// stldice-mr is a MapReduce that reads one "base" STL file and zero or more
// "cut" STL files, dices them into voxels, then produces a stack of images
// resulting from cutting the "base" model with all the subsequent "cuts".
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrislusf/gleam/distributed"
	"github.com/chrislusf/gleam/flow"
	"github.com/chrislusf/gleam/gio"
	gl "github.com/fogleman/fauxgl"
	"github.com/gmlewis/stldice/v4/binvox"
	"github.com/gmlewis/stldice/v4/mr"
	"github.com/golang/protobuf/proto"
)

var (
	MapperVoxelizer = gio.RegisterMapper(voxelize)
	MapperImager    = gio.RegisterMapper(imager)
	// ReducerImager   = gio.RegisterReducer(imager)

	dim = flag.Int("dim", 8192, "Number of voxels along longest axis")
	nX  = flag.Int("nx", 8, "Number of slices along the X dimension")
	nY  = flag.Int("ny", 8, "Number of slices along the Y dimension")
	nZ  = flag.Int("nz", 1, "Number of slices along the Z dimension")
	stl = flag.String("stl", "", "Comma separated list of STL files to process; first is base (e.g. 'base.stl,cut1.stl...')")
	// Flags to control the MapReduce configuration:
	isDistributed = flag.Bool("distributed", false, "Run in distributed mode")
	dockerMaster  = flag.String("docker", "", "Run in docker cluster with master IP address (e.g. 'master:45326')")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\t%v [options] -stl base.stl[,cut1.stl[,cut2.stl,...]]\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse() // optional, since gio.Init() will call this also.
	gio.Init()   // If the command line invokes the mapper or reducer, execute it and exit.

	if *stl == "" {
		log.Fatalf("empty -stl flag")
	}

	ch, err := generateMR(strings.Split(*stl, ","))
	if err != nil {
		log.Fatal(err)
	}

	f := flow.New("stldice-mr").
		Channel(ch).                       // input source files.
		Map("voxelizer", MapperVoxelizer). // invoke the registered "voxelizer" mapper function.
		GroupBy("", flow.Field(1)).        // send all Z values to the same reducer.
		Map("imager", MapperImager).
		Fprintf(os.Stderr, "Done.\n") // Needed to force pipeline to run.

	if *isDistributed {
		log.Println("Running in distributed mode.")
		f.Run(distributed.Option())
	} else if *dockerMaster != "" {
		log.Printf("Running in docker cluster: %v.", *dockerMaster)
		f.Run(distributed.Option().SetMaster(*dockerMaster))
	} else {
		log.Printf("Running in standalone mode.")
		f.Run()
	}
}

// generateMR controls how the MapReduce will run by defining the jobs to
// be performed by each mapper. It also sends all the data each mapper needs
// through the provided channel.
func generateMR(args []string) (chan interface{}, error) {
	ch := make(chan interface{})
	go func() {
		var (
			mbb                  *gl.Box
			subregionScale, mmpv float64
			dimX, dimY, dimZ     int
		)

		for i, arg := range args {
			log.Printf("voxelize: loading file %q...", arg)
			mesh, err := gl.LoadSTL(arg)
			if err != nil {
				log.Fatalf("Unable to load file %q: %v", arg, err)
			}

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
				maxDim := dimX
				if dimY > maxDim {
					maxDim = dimY
				}
				if dimZ > maxDim {
					maxDim = dimZ
				}
				subregionScale = float64(maxDim) * mmpv
			}

			mapIn := &mr.MapIn{
				Base: i == 0,
			}
			for _, t := range mesh.Triangles {
				mapIn.Triangles = append(mapIn.Triangles, &mr.Triangle{
					V1: &mr.Vertex{t.V1.Position.X, t.V1.Position.Y, t.V1.Position.Z},
					V2: &mr.Vertex{t.V2.Position.X, t.V2.Position.Y, t.V2.Position.Z},
					V3: &mr.Vertex{t.V3.Position.X, t.V3.Position.Y, t.V3.Position.Z},
				})
			}

			// Now dice it up.
			for zi := 0; zi < *nZ; zi++ {
				z1 := mbb.Min.Z + float64(zi*dimZ)*mmpv
				for yi := 0; yi < *nY; yi++ {
					y1 := mbb.Min.Y + float64(yi*dimY)*mmpv
					for xi := 0; xi < *nX; xi++ {
						x1 := mbb.Min.X + float64(xi*dimX)*mmpv

						mapIn.VoxelRegion = &mr.VoxelRegion{
							Nx: int64(dimX), Ny: int64(dimY), Nz: int64(dimZ),
							Tx: x1, Ty: y1, Tz: z1,
							Scale: subregionScale,
						}
						data, err := proto.Marshal(mapIn)
						if err != nil {
							log.Fatalf("generateMR: unable to marshal %#v", mapIn)
						}

						log.Printf("generateMR: sending %v to mapper", arg)
						ch <- data
					}
				}
			}
		}
		close(ch)
	}()
	return ch, nil
}

type voxelInfo struct {
	X, Y int
	Base bool
}

// voxelize takes an STL file and dices it into voxels
// then sends them to the imager.
func voxelize(row []interface{}) error {
	// log.Printf("voxelize: len(row)=%v", len(row))
	m := &mr.MapIn{}
	if err := proto.Unmarshal(row[0].([]byte), m); err != nil {
		return err
	}

	var tris []*gl.Triangle
	for _, t := range m.Triangles {
		tris = append(tris, gl.NewTriangleForPoints(
			gl.V(t.V1.X, t.V1.Y, t.V1.Z),
			gl.V(t.V2.X, t.V2.Y, t.V2.Z),
			gl.V(t.V3.X, t.V3.Y, t.V3.Z)))
	}
	mesh := gl.NewTriangleMesh(tris)

	bv := &binvox.BinVOX{
		NX: int(m.VoxelRegion.Nx), NY: int(m.VoxelRegion.Ny), NZ: int(m.VoxelRegion.Nz),
		TX: m.VoxelRegion.Tx, TY: m.VoxelRegion.Ty, TZ: m.VoxelRegion.Tz,
		Scale: m.VoxelRegion.Scale,
	}
	if err := bv.Voxelize(mesh); err != nil {
		return fmt.Errorf("voxelize: %v", err)
	}

	vType := "cut"
	if m.Base {
		vType = "base"
	}
	log.Printf("voxelize: sending %v %v voxels to imager...", len(bv.WhiteVoxels), vType)

	keyFunc := func(k binvox.Key) {
		if k.X < 0 || k.Y < 0 || k.Z < 0 {
			// log.Printf("key %v out-of-bounds (%v,%v,%v)", k, *nX, *nY, *nZ)
			return // common for a cut to extend beyond the bounds of the base.
		}
		gio.Emit(k.Z, voxelInfo{k.X, k.Y, m.Base})
	}
	for k := range bv.WhiteVoxels {
		keyFunc(k)
	}
	for k := range bv.ColorVoxels {
		keyFunc(k)
	}
	return nil
}

type pixelKey struct {
	X, Y int
}

// imager takes the voxels from voxelize and creates
// a 2D image at the provided z height.
// It outputs an image to disk.
func imager(row []interface{}) error {
	z := row[0].(uint64)
	log.Printf("imager: z=%v", z)
	pixels := make(map[pixelKey]bool)
	for _, r := range row[1].([]interface{}) {
		vim := r.(map[interface{}]interface{})
		// log.Printf("vim=%#v, vim[X]=%#v, vim[Y]=%#v, vim[Base]=%#v", vim, vim["X"], vim["Y"], vim["Base"])
		x := int(vim["X"].(uint64))
		y := int(vim["Y"].(uint64))
		base := vim["Base"].(bool)
		k := pixelKey{x, y}
		if v, ok := pixels[k]; ok {
			if v {
				pixels[k] = base
			}
		} else {
			pixels[k] = base
		}
	}

	// Convert collected pixels into an image.
	rect := image.Rect(0, 0, *dim, *dim)
	img := image.NewNRGBA(rect)
	draw.Draw(img, rect, image.Black, image.ZP, draw.Over)
	for k, v := range pixels {
		c := image.Black
		if v {
			c = image.White
		}
		img.Set(k.X, k.Y, c)
	}
	log.Printf("imager: got %v pixels", len(pixels))

	// Write out the PNG file.
	outFile := fmt.Sprintf("out-%v-%v-%v-%v-%v.png", *dim, *nX, *nY, *nZ, z)
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	log.Printf("Writing file %v ...", outFile)
	if err := png.Encode(f, img); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
