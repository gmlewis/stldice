// Packate master implements the master part of the stl2svx-server.
package master

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "github.com/gmlewis/stldice/v2/stl2svx/proto"
	"github.com/gmlewis/stldice/v2/stl2svx/stl"
	"golang.org/x/net/context"
)

const (
	idleWaitTime = 100 * time.Millisecond
	maxSize      = 1 << 30 // 1GB
)

// Master implements the pb.MasterServer interface.
type master struct {
	amu        sync.Mutex          // protects agents and idleAgents
	idleAgents map[string]struct{} // just a map of idle agents

	zmu     sync.Mutex  // protects zipFile
	zipFile *zip.Writer // resulting SVX file
}

func New() *master {
	return &master{idleAgents: make(map[string]struct{})}
}

func (m *master) NewJob(ctx context.Context, in *pb.NewJobRequest) (*pb.NewJobResponse, error) {
	if len(in.GetStlFiles()) == 0 || len(in.GetStlFiles()[0].GetTriangles()) == 0 {
		return nil, fmt.Errorf("Must pass at least one STL file to NewJob")
	}
	base, err := stl.New(in.GetStlFiles()[0], in.GetDim(), in.GetNX(), in.GetNY(), in.GetNZ())
	if err != nil {
		return nil, err
	}

	var zbuf bytes.Buffer
	m.zipFile = zip.NewWriter(&zbuf)

	m.writeManifest(in, base)
	m.writeBlankImages(in, base)

	var wg sync.WaitGroup
	for z := 0; z < base.DimZ; z++ {
		wg.Add(1)
		go func(z int) {
			m.runAgent(ctx, in, z)
			wg.Done()
		}(z)
	}
	wg.Wait()

	if err := m.zipFile.Close(); err != nil {
		return nil, err
	}

	log.Print("Sending SVX file back to client...")
	return &pb.NewJobResponse{SvxFile: zbuf.Bytes()}, nil
}

func (m *master) RegisterAgent(ctx context.Context, in *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	address := in.GetAddress()
	m.amu.Lock()
	m.idleAgents[address] = struct{}{}
	m.amu.Unlock()
	log.Printf("Successfully registered agent %v", address)
	return &pb.RegisterAgentResponse{}, nil
}

// runAgent runs the agent on the given z slice.
func (m *master) runAgent(ctx context.Context, in *pb.NewJobRequest, z int) {
	req := &pb.SliceJobRequest{NewJobRequest: in, Z: int64(z)}
	var conn *grpc.ClientConn
	var address string
	for {
		address = m.getIdleAgent()
		log.Printf("Assigning agent %v to slice z=%v", address, z)
		var err error
		conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxSize)), grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxSize)))
		if err != nil {
			log.Printf("Unable to communicate with agent %v, finding another...", address)
			continue
		}
		break
	}
	defer conn.Close()
	c := pb.NewAgentClient(conn)
	resp, err := c.SliceJob(ctx, req)
	if err != nil {
		log.Printf("Error running SliceJob with agent %v: %v", address, err)
		return
	}
	m.amu.Lock()
	m.idleAgents[address] = struct{}{} // Mark agent as idle
	m.amu.Unlock()

	// Write the PNG to the zip file.
	outFile := fmt.Sprintf("%v/out-%v-%v-%v-%v-%04d.png", in.GetDim(), in.GetDim(), in.GetNX(), in.GetNY(), in.GetNZ(), z+1)

	log.Printf("runAgent(%v): writing %v to zip file...", z, outFile)
	m.zmu.Lock()
	defer m.zmu.Unlock()
	f, err := m.zipFile.Create(outFile)
	if err != nil {
		log.Printf("runAgent(%v) error: %v", z, err)
		return
	}
	if _, err := f.Write(resp.GetPngFile()); err != nil {
		log.Printf("runAgent(%v) error: %v", z, err)
	}
}

// getIdleAgent sits in a loop waiting for an agent to become idle, then marks it busy and returns its address.
func (m *master) getIdleAgent() string {
	for {
		m.amu.Lock()
		var address string
		if len(m.idleAgents) > 0 {
			for address = range m.idleAgents {
				break // get any address from the map
			}
			delete(m.idleAgents, address)
			m.amu.Unlock()
			return address
		}
		m.amu.Unlock()
		time.Sleep(idleWaitTime)
	}
}

func (m *master) writeManifest(in *pb.NewJobRequest, base *stl.STL) error {
	const outFile = "manifest.xml"
	log.Printf("Writing file %v ... need %v agents for job...", outFile, base.DimZ)
	m.zmu.Lock()
	defer m.zmu.Unlock()
	mf, err := m.zipFile.Create(outFile)
	if err != nil {
		return err
	}
	fmt.Fprintf(mf, `<?xml version="1.0"?>
<grid version="1.0" gridSizeX="%v" gridSizeY="%v" gridSizeZ="%v" voxelSize="%v" subvoxelBits="8" originX="%v" originY="%v" originZ="%v" slicesOrientation="Y" >
  <channels>
    <channel type="DENSITY" bits="8" slices="%v/%v-%v-%v-%v-%v-%%04d.png" />
  </channels>
</grid>
`, base.ModelDimX+2, base.ModelDimZ+2, base.ModelDimY+2, base.MMPV*1e-3, base.MBB.Min.X, base.MBB.Min.Z, base.MBB.Min.Y, in.GetDim(), in.GetOutPrefix(), in.GetDim(), in.GetNX(), in.GetNY(), in.GetNZ())
	return nil
}

func (m *master) writeBlankImages(in *pb.NewJobRequest, base *stl.STL) {
	// For Shapeways, create a black base and a black top.
	rect := image.Rect(0, 0, int(in.GetDim()+2), int(in.GetDim()+2))
	img := image.NewGray(rect)
	draw.Draw(img, rect, image.Black, image.ZP, draw.Over)
	m.zmu.Lock()
	defer m.zmu.Unlock()
	outf := func(fn string) {
		f, err := m.zipFile.Create(fn)
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		if err := png.Encode(f, img); err != nil {
			log.Printf("error: %v", err)
		}
	}
	outf(fmt.Sprintf("%v/out-%v-%v-%v-%v-%04d.png", in.GetDim(), in.GetDim(), in.GetNX(), in.GetNY(), in.GetNZ(), 0))
	outf(fmt.Sprintf("%v/out-%v-%v-%v-%v-%04d.png", in.GetDim(), in.GetDim(), in.GetNX(), in.GetNY(), in.GetNZ(), base.DimZ+1))
}
