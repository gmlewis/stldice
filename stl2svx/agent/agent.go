// Packate agent implements the agent part of the stl2svx-server.
package agent

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"sync"
	"time"

	gl "github.com/fogleman/fauxgl"
	"github.com/gmlewis/stldice/v2/binvox"
	pb "github.com/gmlewis/stldice/v2/stl2svx/proto"
	"github.com/gmlewis/stldice/v2/stl2svx/stl"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	idleWaitTime = 100 * time.Millisecond
	maxSize      = 1 << 30 // 1GB
)

// Agent implements the pb.AgentServer interface.
type agent struct{}

// New creates and returns a new agent after registering itself
// with the master.
func New(ctx context.Context, address, masterAddress string) (*agent, error) {
	// Register with the master.
	var conn *grpc.ClientConn
	for {
		var err error
		conn, err = grpc.Dial(masterAddress, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxSize)), grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxSize)))
		if err != nil {
			log.Printf("Agent %v unable to communicate with master, sleeping...", address)
			time.Sleep(idleWaitTime)
			continue
		}
		break
	}
	defer conn.Close()
	c := pb.NewMasterClient(conn)
	req := &pb.RegisterAgentRequest{Address: address}
	_, err := c.RegisterAgent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to register with master: %v", err)
	}

	log.Printf("Successfully registered agent %v with master %v", address, masterAddress)
	return &agent{}, nil
}

func (m *agent) SliceJob(ctx context.Context, in *pb.SliceJobRequest) (resp *pb.SliceJobResponse, imErr error) {
	log.Printf("Agent processing slice z=%v...", in.GetZ())

	ch, err := generateMR(in)
	if err != nil {
		return nil, err
	}

	vxCh := make(chan voxelInfo, 1000)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		resp, imErr = imager(in, vxCh)
		wg.Done()
	}()

	for bv := range ch {
		voxelize(bv, int(in.GetZ()), vxCh)
	}
	close(vxCh)
	wg.Wait()
	if imErr != nil {
		log.Printf("imager: %v", imErr)
		return nil, imErr
	}
	log.Printf("Agent done for z=%v.", in.GetZ())

	return resp, nil
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
func generateMR(in *pb.SliceJobRequest) (ch chan *bvInfo, err error) {
	ch = make(chan *bvInfo)
	go func() {
		job := in.GetNewJobRequest()
		var base *stl.STL
		for i, stlFile := range job.GetStlFiles() {
			stlMesh, err := stl.New(stlFile, job.GetDim(), job.GetNX(), job.GetNY(), job.GetNZ())
			if err != nil {
				log.Fatalf("stl.New: %v", err)
			}
			if i == 0 {
				base = stlMesh
			}
			log.Printf("generateMR: i=%v, %v triangles", i, len(stlMesh.Mesh.Triangles))

			// Now dice it up.
			for zi := 0; zi < int(job.GetNZ()); zi++ {
				z1 := base.MBB.Min.Z + float64(zi*base.DimZ)*base.MMPV
				for yi := 0; yi < int(job.GetNY()); yi++ {
					y1 := base.MBB.Min.Y + float64(yi*base.DimY)*base.MMPV
					for xi := 0; xi < int(job.GetNX()); xi++ {
						x1 := base.MBB.Min.X + float64(xi*base.DimX)*base.MMPV

						mapIn := &bvInfo{
							dx: xi * base.DimX,
							dy: yi * base.DimY,
							dz: zi * base.DimZ,
							bv: &binvox.BinVOX{
								NX: base.DimX, NY: base.DimY, NZ: base.DimZ,
								TX: x1, TY: y1, TZ: z1,
								Scale: base.SubregionScale,
							},
							base: i == 0,
							mesh: stlMesh.Mesh,
						}

						log.Printf("generateMR: sending STL #%v to mapper", i)
						ch <- mapIn
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
	log.Printf("voxelize(%v): sending %v %v voxels to imager(%v)...", zi, len(bv.Voxels), vType, zi)

	for k := range bv.Voxels {
		k.X += bvi.dx
		k.Y += bvi.dy
		k.Z += bvi.dz
		if zi != k.Z {
			log.Fatalf("voxelize(%v): k{%v,%v,%v} does not match zi=%v", zi, k.X, k.Y, k.Z, zi)
		}
		if k.X < 0 || k.Y < 0 || k.Z < 0 {
			continue // common for a cut to extend beyond the bounds of the base.
		}
		ch <- voxelInfo{X: k.X, Y: k.Y, Base: bvi.base}
	}
}

type pixelKey struct {
	X, Y int
}

// imager takes the voxels from voxelize and creates
// a 2D image at the provided z height.
// It outputs an image to disk.
func imager(in *pb.SliceJobRequest, ch <-chan voxelInfo) (*pb.SliceJobResponse, error) {
	job := in.GetNewJobRequest()
	z := in.GetZ()
	log.Printf("imager: z=%v", z)
	pixels := make(map[pixelKey]bool)
	for value := range ch {
		base := value.Base
		// k := pixelKey{value.X, value.Y}
		k := pixelKey{value.X, int(job.GetDim()) - 1 - value.Y} // mirror the Y axis
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
	sz := int(job.GetDim())
	rect := image.Rect(0, 0, sz+2, sz+2)
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
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("imager(%v) error: %v", z, err)
	}
	return &pb.SliceJobResponse{PngFile: buf.Bytes()}, nil
}
