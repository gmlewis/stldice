// stl2svx-client sends jobs to stl2svx-server and writes the
// results to local disk.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	gl "github.com/fogleman/fauxgl"
	pb "github.com/gmlewis/stldice/v4/stl2svx/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	maxSize = 1 << 30 // 1GB
)

var (
	dim           = flag.Int("dim", 8192, "Number of voxels along longest axis")
	nX            = flag.Int("nx", 8, "Number of slices along the X dimension")
	nY            = flag.Int("ny", 8, "Number of slices along the Y dimension")
	nZ            = flag.Int("nz", 1, "Number of slices along the Z dimension")
	stl           = flag.String("stl", "", "Comma separated list of STL files to process; first is base (e.g. 'base.stl,cut1.stl...')")
	prefix        = flag.String("prefix", "out", "Prefix for output SVX file")
	masterAddress = flag.String("master", "", "Address used by agent to contact master")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\t%v [options] -stl base.stl[,cut1.stl[,cut2.stl,...]]\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *stl == "" {
		log.Fatal("must specify -stl file(s)")
	}

	conn, err := grpc.Dial(*masterAddress, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxSize)), grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxSize)))
	if err != nil {
		log.Fatalf("Unable to communicate with master %v: %v", masterAddress, err)
	}
	defer conn.Close()

	req := &pb.NewJobRequest{
		Dim:       int64(*dim),
		NX:        int64(*nX),
		NY:        int64(*nY),
		NZ:        int64(*nZ),
		OutPrefix: *prefix,
	}
	req.StlFiles, err = loadSTL(strings.Split(*stl, ","))
	if err != nil {
		log.Fatalf("loadSTL: %v", err)
	}

	c := pb.NewMasterClient(conn)
	resp, err := c.NewJob(context.Background(), req)
	if err != nil {
		log.Fatalf("master: %v", err)
	}

	outFile := fmt.Sprintf("%v-%v-%v-%v-%v.svx", *prefix, *dim, *nX, *nY, *nZ)
	log.Printf("Writing SVX file %v ...", outFile)
	if err := ioutil.WriteFile(outFile, resp.GetSvxFile(), 0644); err != nil {
		log.Fatalf("write: %v", err)
	}

	log.Print("Done.")
}

func loadSTL(args []string) (out []*pb.STLFile, err error) {
	for i, arg := range args {
		log.Printf("loadSTL: loading file #%v: %v ...", i, arg)
		mesh, err := gl.LoadSTL(arg)
		if err != nil {
			return nil, err
		}
		log.Printf("loadSTL: loaded %v triangles", len(mesh.Triangles))

		stl := &pb.STLFile{}
		out = append(out, stl)
		for _, t := range mesh.Triangles {
			stl.Triangles = append(stl.Triangles, &pb.Triangle{
				V1: &pb.Vertex{t.V1.Position.X, t.V1.Position.Y, t.V1.Position.Z},
				V2: &pb.Vertex{t.V2.Position.X, t.V2.Position.Y, t.V2.Position.Z},
				V3: &pb.Vertex{t.V3.Position.X, t.V3.Position.Y, t.V3.Position.Z},
			})
		}
	}
	return out, nil
}
