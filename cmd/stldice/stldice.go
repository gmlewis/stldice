// stldice dices up STL meshes into one or more 'binvox' files.
//
// Usage:
//   stldice base.stl all-cuts.stl
//   stldice -mbb "(-80,-80,-2.6)-(80,80,0.6)" all-cuts.stl
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	gl "github.com/fogleman/fauxgl"
	"github.com/gmlewis/stldice/v2/binvox"
)

var (
	mbbString     = flag.String("mbb", "", "MBB of the STL file, e.g. '(-80,-80,-2.6)-(80,80,0.6)'; empty means to calculate it from the first STL file; units are in millimeters")
	dim           = flag.Int("dim", 8192, "Number of voxels along longest axis")
	nX            = flag.Int("nx", 8, "Number of slices along the X dimension")
	nY            = flag.Int("ny", 8, "Number of slices along the Y dimension")
	nZ            = flag.Int("nz", 1, "Number of slices along the Z dimension")
	smoothDegrees = flag.Float64("smooth", 0, "Degrees used for smoothing normals (0=no smoothing)")

	mbbRE = regexp.MustCompile(`^\s*[\(\[]?\s*([^,\s]+)\s*,\s*([^,\s]+)\s*,\s*([^\]\)\s]+)\s*[\]\)]?\s*\-\s*[\(\[]?\s*([^,\s]+)\s*,\s*([^,\s]+)\s*,\s*([^\]\)\s]+)\s*[\)\]]?\s*$`)
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\t%v [options] base.stl all-cuts.stl ...\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Must supply at least one filename")
	}
	if *dim%*nX != 0 || *dim%*nY != 0 || *dim%*nZ != 0 {
		log.Fatal("dim must be an integral multiple of nx, ny, and nz")
	}

	var mbb *gl.Box
	if *mbbString != "" {
		var err error
		if mbb, err = parseMBB(*mbbString); err != nil {
			log.Fatalf("Unable to parse mbb: %v", err)
		}
	}

	log.Printf("stldice -dim %v -nx %v -ny %v -nz %v -mbb %q %v", *dim, *nX, *nY, *nZ, *mbbString, strings.Join(flag.Args(), " "))

	prefixes := map[string]bool{}
	for _, arg := range flag.Args() {
		log.Printf("Loading file %q...", arg)
		mesh, err := gl.LoadSTL(arg)
		if err != nil {
			log.Fatalf("Unable to load file %q: %v", arg, err)
		}
		if *smoothDegrees > 0 {
			log.Printf("Smoothing mesh normals with %v degree threshold...", *smoothDegrees)
			mesh.SmoothNormalsThreshold(gl.Radians(*smoothDegrees))
			log.Println("Done smoothing mesh normals.")
		}
		if mbb == nil {
			box := mesh.BoundingBox()
			mbb = &box
			log.Printf(`Mesh MBB: "(%v,%v,%v)-(%v,%v,%v)"`, mbb.Min.X, mbb.Min.Y, mbb.Min.Z, mbb.Max.X, mbb.Max.Y, mbb.Max.Z)
		}
		scale := mbb.Max.X - mbb.Min.X
		if dy := mbb.Max.Y - mbb.Min.Y; dy > scale {
			scale = dy
		}
		if dz := mbb.Max.Z - mbb.Min.Z; dz > scale {
			scale = dz
		}

		outPrefix := fmt.Sprintf("%v-%v-%v-%v-%v", strings.TrimSuffix(arg, ".stl"), *dim, *nX, *nY, *nZ)
		pfxs, err := diceMesh(mesh, mbb, scale, *dim, *nX, *nY, *nZ, outPrefix)
		if err != nil {
			log.Fatalf("Unable to dice mesh: %v", err)
		}
		for _, v := range pfxs {
			prefixes[v] = true // dedupe
		}
	}

	// Print out the commands needed to complete the operation.
	fmt.Println("#!/bin/bash -x") // Do not use -e due to unwritten (empty) binvox files stopping the script.
	var pfxs []string
	for k := range prefixes {
		pfxs = append(pfxs, k)
	}
	sort.Strings(pfxs)
	stlOutPrefix := "out"
	if flag.NArg() == 1 {
		stlOutPrefix = strings.TrimSuffix(flag.Arg(0), ".stl")
	}
	common := fmt.Sprintf("%v-%v-%v-%v", *dim, *nX, *nY, *nZ)
	for _, v := range pfxs {
		fmt.Printf("voxcut -ostl %[1]v-%[4]v%[3]v.stl %[2]v-%[4]v%[3]v.binvox", stlOutPrefix, strings.TrimSuffix(flag.Arg(0), ".stl"), v, common)
		for _, arg := range flag.Args()[1:] {
			fmt.Printf(" %v-%v%v.binvox", strings.TrimSuffix(arg, ".stl"), common, v)
		}
		fmt.Println()
	}
	if len(pfxs) > 1 { // Merge all the STL files back together
		fmt.Printf("merge-stl -f -prefix %v-%v\n", stlOutPrefix, common)
	}

	log.Println("Done.")
}

// diceMesh dices a mesh into voxelized regions and writes the regions into .binvox files.
// nx, ny, nz are the number of divisions for each dimension.
// prefixes are the file prefixes that were generated.
func diceMesh(mesh *gl.Mesh, mbb *gl.Box, scale float64, dim, nx, ny, nz int, outPrefix string) (prefixes []string, err error) {
	vpmm := float64(dim) / scale // voxels per millimeter
	mmpv := 1.0 / vpmm           // millimeters per voxel
	modelDimInMM := mbb.Size()
	log.Printf("diceMesh(mbb=(%v,%v,%v)-(%v,%v,%v), size=%v, scale=%v, dim=%v, n=[%v,%v,%v], outPrefix=%v); %v voxels per millimeter; %v millimeters per voxel", mbb.Min.X, mbb.Min.Y, mbb.Min.Z, mbb.Max.X, mbb.Max.Y, mbb.Max.Z, modelDimInMM, scale, dim, nx, ny, nz, outPrefix, vpmm, mmpv)

	newModelDimX := int(math.Ceil(modelDimInMM.X * vpmm))
	newModelDimY := int(math.Ceil(modelDimInMM.Y * vpmm))
	newModelDimZ := int(math.Ceil(modelDimInMM.Z * vpmm))
	log.Printf("number of voxels needed along each axis to cover entire model: [%v,%v,%v]", newModelDimX, newModelDimY, newModelDimZ)

	// subregion dimensions
	dimX := newModelDimX / nx
	dimY := newModelDimY / ny
	dimZ := newModelDimZ / nz
	if dimX == 0 || dimY == 0 || dimZ == 0 {
		return nil, fmt.Errorf("too many divisions: region dimensions = (%v,%v,%v)", dimX, dimY, dimZ)
	}

	maxDim := dimX
	if dimY > maxDim {
		maxDim = dimY
	}
	if dimZ > maxDim {
		maxDim = dimZ
	}
	log.Printf("maximum dimension = %v (the smaller this number, the faster the binvox writing will be)", maxDim)
	subregionScale := float64(maxDim) * mmpv
	log.Printf("subregion scale = %v", subregionScale)

	for zi := 0; zi < nz; zi++ {
		z1 := mbb.Min.Z + float64(zi*dimZ)*mmpv
		for yi := 0; yi < ny; yi++ {
			y1 := mbb.Min.Y + float64(yi*dimY)*mmpv
			for xi := 0; xi < nx; xi++ {
				newPrefix := fmt.Sprintf("%v-%02d-%02d-%02d", outPrefix, xi, yi, zi)
				outFile := newPrefix + ".binvox"
				if _, err := os.Stat(outFile); err == nil || os.IsExist(err) {
					log.Printf("Skipping writing existing file %v", outFile)
					continue
				}

				x1 := mbb.Min.X + float64(xi*dimX)*mmpv
				bv := &binvox.BinVOX{
					NX: dimX, NY: dimY, NZ: dimZ,
					TX: x1, TY: y1, TZ: z1,
					Scale: subregionScale,
				}
				if err := bv.Voxelize(mesh); err != nil {
					return nil, fmt.Errorf("voxelize: %v", err)
				}

				if len(bv.Voxels) == 0 {
					log.Printf("No voxels created; skipping writing empty file %v", outFile)
					continue
				}
				log.Printf("Writing %v voxels to %v", len(bv.Voxels), outFile)
				if err := bv.Write(outFile, 0, 0, 0, 0, 0, 0); err != nil {
					return nil, err
				}
				prefixes = append(prefixes, strings.TrimPrefix(newPrefix, outPrefix))
			}
		}
	}
	return prefixes, nil
}

// parseMBB parses a string representation of a minimum bounding box (MBB).
func parseMBB(s string) (mbb *gl.Box, err error) {
	mbb = &gl.Box{}
	parts := mbbRE.FindStringSubmatch(s)
	if len(parts) != 7 {
		return nil, fmt.Errorf("incorrect format: %q", s)
	}
	if mbb.Min.X, err = strconv.ParseFloat(parts[1], 64); err != nil {
		return nil, fmt.Errorf("invalid number %v: %v", parts[1], err)
	}
	if mbb.Min.Y, err = strconv.ParseFloat(parts[2], 64); err != nil {
		return nil, fmt.Errorf("invalid number %v: %v", parts[2], err)
	}
	if mbb.Min.Z, err = strconv.ParseFloat(parts[3], 64); err != nil {
		return nil, fmt.Errorf("invalid number %v: %v", parts[3], err)
	}
	if mbb.Max.X, err = strconv.ParseFloat(parts[4], 64); err != nil {
		return nil, fmt.Errorf("invalid number %v: %v", parts[4], err)
	}
	if mbb.Max.Y, err = strconv.ParseFloat(parts[5], 64); err != nil {
		return nil, fmt.Errorf("invalid number %v: %v", parts[5], err)
	}
	if mbb.Max.Z, err = strconv.ParseFloat(parts[6], 64); err != nil {
		return nil, fmt.Errorf("invalid number %v: %v", parts[6], err)
	}
	return mbb, nil
}
