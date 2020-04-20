// -*- compile-command: "CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' ./stl2svx.go"; -*-

// stl2svx is a MapReduce that reads one "base" STL file and zero or more
// "cut" STL files, dices them into voxels, then produces a stack of images
// resulting from cutting the "base" model with all the subsequent "cuts".
// It packs all these slices into an SVX file. See https://abfab3d.com/svx-format
package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	gl "github.com/fogleman/fauxgl"
	"github.com/gmlewis/stldice/v4/binvox"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var (
	dim    = flag.Int("dim", 8192, "Number of voxels along longest axis")
	nX     = flag.Int("nx", 8, "Number of slices along the X dimension")
	nY     = flag.Int("ny", 8, "Number of slices along the Y dimension")
	nZ     = flag.Int("nz", 1, "Number of slices along the Z dimension")
	stl    = flag.String("stl", "", "Comma separated list of STL files to process; first is base (e.g. 'base.stl,cut1.stl...')")
	bucket = flag.String("bucket", "", "Google Cloud Storage bucket in which to save images (e.g. 'gmlewis.appspot.com')")

	storageBucket *storage.BucketHandle
	zipFile       *zip.Writer
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\t%v [options] -stl base.stl[,cut1.stl[,cut2.stl,...]]\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *bucket != "" {
		ctx := context.Background()
		ts, err := google.DefaultTokenSource(ctx, storage.ScopeFullControl)
		if err != nil {
			log.Fatalf("could not get token source: %v", err)
		}
		client, err := storage.NewClient(ctx, option.WithTokenSource(ts))
		if err != nil {
			log.Fatalf("could not connect to Google Cloud Storage: %v", err)
		}
		storageBucket = client.Bucket(*bucket)
		defer client.Close()
	}

	// If any STL files are passed in, this is the master.
	if *stl != "" {
		master()
	}

	if err := zipFile.Close(); err != nil {
		log.Printf("error: %v", err)
	}

	log.Print("Done.")
}

func master() {
	_, dimZ, err := generateMR(strings.Split(*stl, ","), true) // just get dimZ
	if err != nil {
		log.Fatal(err)
	}

	// For Shapeways, create a black base and a black top.
	rect := image.Rect(0, 0, *dim+2, *dim+2)
	img := image.NewGray(rect)
	draw.Draw(img, rect, image.Black, image.ZP, draw.Over)
	outf := func(fn string) {
		f, err := zipFile.Create(fn)
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		if err := png.Encode(f, img); err != nil {
			log.Printf("error: %v", err)
		}
	}
	outf(fmt.Sprintf("%v/out-%v-%v-%v-%v-%04d.png", *dim, *dim, *nX, *nY, *nZ, 0))

	for i := 0; i < dimZ; i++ {
		agent(i)
	}

	outf(fmt.Sprintf("%v/out-%v-%v-%v-%v-%04d.png", *dim, *dim, *nX, *nY, *nZ, dimZ+1))
}

func agent(agentID int) {
	log.Printf("Agent %v to process: %v...", agentID, *stl)

	ch, _, err := generateMR(strings.Split(*stl, ","), false)
	if err != nil {
		log.Fatal(err)
	}

	vxCh := make(chan voxelInfo, 1000)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Println(imager(agentID, vxCh))
		wg.Done()
	}()

	for bv := range ch {
		voxelize(bv, agentID, vxCh)
	}
	close(vxCh)
	wg.Wait()
	log.Printf("Agent %v done.", agentID)
}

type bvInfo struct {
	dx, dy, dz int
	bv         *binvox.BinVOX
	base       bool
	mesh       *gl.Mesh
}

// generateMR controls how the MapReduce will run by defining the jobs to
// be performed by each mapper. It also sends all the data each mapper needs
// through the provided channel.
// dimZ is the number of agents needed to output the images.
func generateMR(args []string, dimOnly bool) (ch chan *bvInfo, dimZ int, err error) {
	ch = make(chan *bvInfo)
	var wg sync.WaitGroup // Wait to return when dimZ has been calculated
	wg.Add(1)
	go func() {
		var (
			mbb                  *gl.Box
			subregionScale, mmpv float64
			dimX, dimY           int
		)

		for i, arg := range args {
			log.Printf("generateMR: loading file %q...", arg)
			mesh, err := gl.LoadSTL(arg)
			if err != nil {
				log.Fatalf("Unable to load file %q: %v", arg, err)
			}
			log.Printf("generateMR: loaded %v triangles", len(mesh.Triangles))

			if i == 0 {
				box := mesh.BoundingBox()
				mbb = &box
				scale := mbb.Max.X - mbb.Min.X
				if dy := mbb.Max.Y - mbb.Min.Y; dy > scale {
					scale = dy
				}
				if dz := mbb.Max.Z - mbb.Min.Z; dz > scale {
					scale = dz
				}

				vpmm := float64(*dim) / scale // voxels per millimeter
				mmpv = 1.0 / vpmm             // millimeters per voxel
				modelDimInMM := mbb.Size()
				newModelDimX := int(math.Ceil(modelDimInMM.X * vpmm))
				newModelDimY := int(math.Ceil(modelDimInMM.Y * vpmm))
				newModelDimZ := int(math.Ceil(modelDimInMM.Z * vpmm))
				dimX = newModelDimX / *nX
				dimY = newModelDimY / *nY
				dimZ = newModelDimZ / *nZ
				if dimX == 0 || dimY == 0 || dimZ == 0 {
					log.Fatalf("too many divisions: region dimensions = (%v,%v,%v)", dimX, dimY, dimZ)
				}
				// dimZ has now been calculated... parent function can return.
				wg.Done()
				if dimOnly { // generate the manifest file.
					// Note that "Up" in Shapeways is +Y, so swap Y and Z (and keep mirroring Y when writing the image).
					// Shapeways will *not* accept slicesOrientation="Z" despite documentation.
					// Also swap the originY/Z values.
					b := fmt.Sprintf(`<?xml version="1.0"?>
<grid version="1.0" gridSizeX="%v" gridSizeY="%v" gridSizeZ="%v" voxelSize="%v" subvoxelBits="8" originX="%v" originY="%v" originZ="%v" slicesOrientation="Y" >
  <channels>
    <channel type="DENSITY" bits="8" slices="%v/out-%v-%v-%v-%v-%%04d.png" />
  </channels>
</grid>
`, newModelDimX+2, newModelDimZ+2, newModelDimY+2, mmpv*1e-3, mbb.Min.X, mbb.Min.Z, mbb.Min.Y, *dim, *dim, *nX, *nY, *nZ)
					log.Printf("SVX manifest.xml file:\n\n%v\n\n", b)
					writeManifest(b, *dim)
					break // don't continue
				}
				maxDim := dimX
				if dimY > maxDim {
					maxDim = dimY
				}
				if dimZ > maxDim {
					maxDim = dimZ
				}
				subregionScale = float64(maxDim) * mmpv
			}

			// Now dice it up.
			for zi := 0; zi < *nZ; zi++ {
				z1 := mbb.Min.Z + float64(zi*dimZ)*mmpv
				for yi := 0; yi < *nY; yi++ {
					y1 := mbb.Min.Y + float64(yi*dimY)*mmpv
					for xi := 0; xi < *nX; xi++ {
						x1 := mbb.Min.X + float64(xi*dimX)*mmpv

						mapIn := &bvInfo{
							dx: xi * dimX,
							dy: yi * dimY,
							dz: zi * dimZ,
							bv: &binvox.BinVOX{
								NX: dimX, NY: dimY, NZ: dimZ,
								TX: x1, TY: y1, TZ: z1,
								Scale: subregionScale,
							},
							base: i == 0,
							mesh: mesh,
						}

						log.Printf("generateMR: sending %v to mapper", arg)
						ch <- mapIn
					}
				}
			}
		}
		close(ch)
	}()
	wg.Wait()
	return ch, dimZ, nil
}

type voxelInfo struct {
	X, Y int
	Base bool
}

// voxelize takes an STL file and dices it into voxels
// then sends them to the imager.
func voxelize(bvi *bvInfo, zi int, ch chan<- voxelInfo) {
	bv := bvi.bv
	if err := bv.VoxelizeZ(bvi.mesh, zi); err != nil {
		log.Printf("voxelize(%v): %v", zi, err)
		return
	}

	vType := "cut"
	if bvi.base {
		vType = "base"
	}
	log.Printf("voxelize(%v): sending %v %v voxels to imager(%v)...", zi, len(bv.WhiteVoxels), vType, zi)

	keyFunc := func(k binvox.Key) {
		k.X += bvi.dx
		k.Y += bvi.dy
		k.Z += bvi.dz
		if zi != k.Z {
			log.Fatalf("voxelize(%v): k{%v,%v,%v} does not match zi=%v", zi, k.X, k.Y, k.Z, zi)
		}
		if k.X < 0 || k.Y < 0 || k.Z < 0 {
			return // common for a cut to extend beyond the bounds of the base.
		}
		ch <- voxelInfo{X: k.X, Y: k.Y, Base: bvi.base}
	}
	for k := range bv.WhiteVoxels {
		keyFunc(k)
	}
	for k := range bv.ColorVoxels {
		keyFunc(k)
	}
}

type pixelKey struct {
	X, Y int
}

// imager takes the voxels from voxelize and creates
// a 2D image at the provided z height.
// It outputs an image to disk.
func imager(z int, ch <-chan voxelInfo) string {
	log.Printf("imager: z=%v", z)
	pixels := make(map[pixelKey]bool)
	for value := range ch {
		base := value.Base
		// k := pixelKey{value.X, value.Y}
		k := pixelKey{value.X, *dim - 1 - value.Y} // mirror the Y axis
		if v, ok := pixels[k]; ok {
			if v {
				pixels[k] = base
			}
		} else {
			pixels[k] = base
		}
	}
	log.Printf("imager(%v): processed %v voxels", z, len(pixels))

	// Convert collected pixels into an image.
	rect := image.Rect(0, 0, *dim+2, *dim+2)
	img := image.NewGray(rect)
	// Shapeways requires a single pixel boundary in every dimension!!!
	draw.Draw(img, rect, image.Black, image.ZP, draw.Over)
	for k, v := range pixels {
		c := image.Black
		if v {
			c = image.White
		}
		img.Set(k.X+1, k.Y+1, c)
	}
	log.Printf("imager(%v): writing %v pixels", z, len(pixels))

	// Write out the PNG file.
	outFile := fmt.Sprintf("%v/out-%v-%v-%v-%v-%04d.png", *dim, *dim, *nX, *nY, *nZ, z+1)

	log.Printf("imager(%v): writing %v to zip file...", z, outFile)
	f, err := zipFile.Create(outFile)
	if err != nil {
		return fmt.Sprintf("imager(%v) error: %v", z, err)
	}

	if err := png.Encode(f, img); err != nil {
		return fmt.Sprintf("imager(%v) error: %v", z, err)
	}
	return fmt.Sprintf("imager(%v) done.", z)
}

func writeManifest(buf string, dim int) {
	var f io.WriteCloser
	const outFile = "manifest.xml"
	zipName := fmt.Sprintf("out%v.svx", dim)
	if *bucket != "" {
		log.Printf("writing %v to bucket %v...", zipName, *bucket)
		ctx := context.Background()
		w := storageBucket.Object(zipName).NewWriter(ctx)
		w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
		w.ContentType = "archive/zip"
		// Entries are immutable, be aggressive about caching (1 day).
		w.CacheControl = "public, max-age=86400"
		f = w
	} else {
		log.Printf("writing %v locally...", zipName)
		w, err := os.Create(zipName)
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		f = w
	}

	zipFile = zip.NewWriter(f)
	log.Printf("Writing file %v ...", outFile)
	mf, err := zipFile.Create(outFile)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	fmt.Fprint(mf, buf)
}
