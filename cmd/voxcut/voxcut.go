// voxcut performs boolean operations on 'binvox' files.
//
// Note that the binvox files must be based on the same voxel 3D grid
// meaning that all vox files have the same voxels per milliemeter.
//
// Since binvox models can be very large and possibly not fit into
// memory, voxcut supports dicing the model into smaller chunks.
// To facilitate this, start indices and counts for each dimension
// can be provided to process only a smaller section of the model.
//
// Usage:
//   voxcut [options] base.binvox [cut1.binvox [cut2.binvox ...]]
package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	gl "github.com/fogleman/fauxgl"
	"github.com/gmlewis/stldice/v4/binvox"
)

var (
	binVOXFile    = flag.String("obinvox", "", "The output binvox filename to create")
	stlFile       = flag.String("ostl", "", "The output stl filename to create")
	voxFile       = flag.String("ovox", "", "The output vox filename to create")
	startX        = flag.Int("sx", 0, "The start X index (default=0)")
	startY        = flag.Int("sy", 0, "The start Y index (default=0)")
	startZ        = flag.Int("sz", 0, "The start Z index (default=0)")
	countX        = flag.Int("cx", 0, "The number of voxels to process in the X direction (default=0=all)")
	countY        = flag.Int("cy", 0, "The number of voxels to process in the Y direction (default=0=all)")
	countZ        = flag.Int("cz", 0, "The number of voxels to process in the Z direction (default=0=all)")
	smoothDegrees = flag.Float64("smooth", 0, "Degrees used for smoothing normals (0=no smoothing)")
	manifold      = flag.Bool("manifold", false, "Output manifold mesh - useful for low-res cutouts")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\t%v [options] base.binvox [cut1.binvox [cut2.binvox ...]]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Must supply at least one filename")
	}
	if *binVOXFile == "" && *stlFile == "" && *voxFile == "" {
		log.Fatal("Must specify at least one of -obinvox, -ostl, or -ovox")
	}

	base, err := binvox.Read(flag.Arg(0), *startX, *startY, *startZ, *countX, *countY, *countZ)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i < flag.NArg(); i++ {
		cut, err := binvox.Read(flag.Arg(i), *startX, *startY, *startZ, *countX, *countY, *countZ)
		if err != nil {
			log.Printf("skipping: %v", err)
			continue
		}

		log.Printf("\n\nCutting voxel model with %q...", flag.Arg(i))
		base.WhiteVoxels, err = voxcut(base, cut)
		if err != nil {
			log.Fatalf("arg #%v: cut with %q: %v", i, flag.Arg(i), err)
		}
		if len(base.WhiteVoxels) == 0 {
			log.Fatal("result of cut leaves no non-zero voxels... no need to write file")
		}
		log.Printf("Done cutting voxel model with %q.", flag.Arg(i))
	}

	if *binVOXFile != "" {
		log.Printf("Writing file %q...", *binVOXFile)
		err := base.Write(*binVOXFile, *startX, *startY, *startZ, *countX, *countY, *countZ)
		if err != nil {
			log.Fatalf("binvox.Write(%q): %v", *binVOXFile, err)
		}
	}

	if *voxFile != "" {
		log.Printf("Writing file %q...", *voxFile)
		f, err := os.Create(*voxFile)
		if err != nil {
			log.Fatalf("Unable to create file %v: %v", *voxFile, err)
		}
		if err := writeVOX(f, base.WhiteVoxels); err != nil {
			log.Fatalf("writeVOX: %v", err)
		}
		if err := f.Close(); err != nil {
			log.Fatalf("Unable to close file %v: %v", *voxFile, err)
		}
	}

	if *stlFile != "" {
		var mesh *gl.Mesh
		if *manifold {
			mesh = base.ManifoldMesh()
		} else {
			mesh = base.ToMesh()
		}

		if *smoothDegrees > 0 {
			log.Printf("Smoothing mesh normals with %v degree threshold...", *smoothDegrees)
			mesh.SmoothNormalsThreshold(gl.Radians(*smoothDegrees))
			log.Println("Done smoothing mesh normals.")
		}

		log.Printf("Writing file %v ...", *stlFile)
		if err := mesh.SaveSTL(*stlFile); err != nil {
			log.Fatalf("SaveSTL: %v", err)
		}
	}

	log.Println("Done.")
}

// voxcut cuts the base voxels by the cut voxels and returns the newVoxels.
// It accounts for different translation settings in the base and cutting voxels.
func voxcut(base, cut *binvox.BinVOX) (newVoxels binvox.WhiteVoxelMap, err error) {
	if len(cut.WhiteVoxels) == 0 {
		return base.WhiteVoxels, nil // Nothing to cut.
	}
	if len(base.WhiteVoxels) == 0 {
		return nil, errors.New("base must not be empty")
	}

	newVoxels = binvox.WhiteVoxelMap{}
	vpmm := base.VoxelsPerMM()

	// Map base indices to cut indices by taking into account the translations.
	dx := int((base.TX - cut.TX) * vpmm)
	dy := int((base.TY - cut.TY) * vpmm)
	dz := int((base.TZ - cut.TZ) * vpmm)
	log.Printf("Translating cut voxels by [%v,%v,%v]", dx, dy, dz)

	keyFunc := func(v binvox.Key, vc binvox.Color) {
		c, ok := cut.Get(v.X, v.Y, v.Z)
		if !ok { // Nothing to cut - keep voxel.
			newVoxels[v] = struct{}{}
			return
		}

		// TODO: What does it mean here to start supporting color?!?
		r := gl.Clamp(vc.R-c.A*c.R, 0, 1)
		g := gl.Clamp(vc.G-c.A*c.G, 0, 1)
		b := gl.Clamp(vc.B-c.A*c.B, 0, 1)
		a := vc.A
		if a > 0 && (r > 0 || g > 0 || b > 0) {
			// newVoxels[binvox.Key{v.X, v.Y, v.Z}] = binvox.Color{r, g, b, a}
			newVoxels[binvox.Key{v.X, v.Y, v.Z}] = struct{}{}
		}
	}
	for v := range base.WhiteVoxels {
		keyFunc(v, binvox.White)
	}
	for v, vc := range base.ColorVoxels {
		keyFunc(v, vc)
	}

	return newVoxels, nil
}

// writeVOX writes a vox file from the base voxels.
func writeVOX(f io.Writer, base binvox.WhiteVoxelMap) error {
	header := gl.VOXHeader{Magic: [4]byte{'V', 'O', 'X', ' '}, Version: 150}
	if err := binary.Write(f, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("header: %v", err)
	}

	chunk := gl.VOXChunk{ID: [4]byte{'M', 'A', 'I', 'N'}, ChildrenBytes: int32(12 + len(base)*4 + 4)}
	if err := binary.Write(f, binary.LittleEndian, &chunk); err != nil {
		return fmt.Errorf("MAIN: %v", err)
	}
	chunk = gl.VOXChunk{ID: [4]byte{'S', 'I', 'Z', 'E'}, ContentBytes: 12}
	if err := binary.Write(f, binary.LittleEndian, &chunk); err != nil {
		return fmt.Errorf("SIZE: %v", err)
	}

	var maxX, maxY, maxZ int
	for v := range base {
		if v.X > maxX {
			maxX = v.X
		}
		if v.Y > maxY {
			maxY = v.Y
		}
		if v.Z > maxZ {
			maxZ = v.Z
		}
	}
	if err := binary.Write(f, binary.LittleEndian, int32(maxX)); err != nil {
		return fmt.Errorf("maxX: %v", err)
	}
	if err := binary.Write(f, binary.LittleEndian, int32(maxY)); err != nil {
		return fmt.Errorf("maxY: %v", err)
	}
	if err := binary.Write(f, binary.LittleEndian, int32(maxZ)); err != nil {
		return fmt.Errorf("maxZ: %v", err)
	}

	chunk = gl.VOXChunk{ID: [4]byte{'X', 'Y', 'Z', 'I'}, ContentBytes: int32(len(base)*4 + 4)}
	if err := binary.Write(f, binary.LittleEndian, &chunk); err != nil {
		return fmt.Errorf("XYZI: %v", err)
	}
	numVoxels := int32(len(base))
	if err := binary.Write(f, binary.LittleEndian, &numVoxels); err != nil {
		return fmt.Errorf("numVoxels: %v", err)
	}
	// TODO(gmlewis): support full color palette.
	// for v, vc := range base {
	for v := range base {
		// i := int32(0)
		// if vc.A >= 0.5 && vc.R+vc.G+vc.B > 1.5 {  // WHAT?!?
		// 	i = 1
		// }
		if err := binary.Write(f, binary.LittleEndian, int32(v.X)); err != nil {
			return fmt.Errorf("X: %v", err)
		}
		if err := binary.Write(f, binary.LittleEndian, int32(v.Y)); err != nil {
			return fmt.Errorf("Y: %v", err)
		}
		if err := binary.Write(f, binary.LittleEndian, int32(v.Z)); err != nil {
			return fmt.Errorf("Z: %v", err)
		}
		i := int32(1)
		if err := binary.Write(f, binary.LittleEndian, i); err != nil {
			return fmt.Errorf("I: %v", err)
		}
	}
	return nil
}
