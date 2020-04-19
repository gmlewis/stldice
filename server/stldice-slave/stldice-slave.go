// stldice-slave is a gRPC server that listens for commands from the
// stldice-master server and processes the requests.
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	gl "github.com/fogleman/fauxgl"
	"github.com/gmlewis/stldice/binvox"
	pb "github.com/gmlewis/stldice/stldice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const modelFilePrefix = "model"

var (
	port = flag.String("port", ":5333", "Port to listen on")
)

// server is used to implement stldice.STLDiceServer
type server struct{}

// Add an STL mesh to the voxels (possibly a subregion on the slave).
func (s *server) AddSTLMesh(ctx context.Context, in *pb.AddSTLMeshRequest) (*pb.AddSTLMeshReply, error) {
	// Voxelize the STL file using the provided VoxelRegion.
	mesh, err := stlToMesh(in.GetStlFile())
	if err != nil {
		return nil, err
	}
	addBV, err := voxelize(in.GetVoxelRegion(), mesh)
	if err != nil {
		return nil, err
	}

	// If there is no model on disk with the same coordinates, write the new binvox file.
	modelFilename := filename(in.GetVoxelRegion())
	base, err := binvox.Read(modelFilename, 0, 0, 0, 0, 0, 0)
	if err != nil { // No existing model of the same name. Write the new one.
		if err := mesh.SaveSTL(modelFilename); err != nil {
			return nil, err
		}
		return &pb.AddSTLMeshReply{}, nil // Success.
	}

	// If there already is a model on disk, load it, add the newly-generated voxels to it, then save it back out,
	// overwriting the old version.
	newBase := voxelOp(base, addBV, true)
	if err := newBase.Write(modelFilename, 0, 0, 0, 0, 0, 0); err != nil {
		return nil, err
	}
	return &pb.AddSTLMeshReply{}, nil // Success.
}

// Subtract an STL mesh from the voxels (possibly a subregion on the slave).
func (s *server) SubSTLMesh(ctx context.Context, in *pb.SubSTLMeshRequest) (*pb.SubSTLMeshReply, error) {
	// If there is no model on disk with the same coordinates, return an error.
	modelFilename := filename(in.GetVoxelRegion())
	// Load the model binvox file from disk.
	base, err := binvox.Read(modelFilename, 0, 0, 0, 0, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("unable to subtract, no existing model: %v", err)
	}

	// Voxelize the STL file using the provided VoxelRegion.
	mesh, err := stlToMesh(in.GetStlFile())
	if err != nil {
		return nil, err
	}
	subBV, err := voxelize(in.GetVoxelRegion(), mesh)
	if err != nil {
		return nil, err
	}

	// Cut the base model with this mesh, then save it back out, overwriting the old version.
	newBase := voxelOp(base, subBV, false)
	if err := newBase.Write(modelFilename, 0, 0, 0, 0, 0, 0); err != nil {
		return nil, err
	}
	return &pb.SubSTLMeshReply{}, nil
}

// Get back the model voxels as STL (possibly a subregion on the slave).
// This also resets the server by clearing out the model.
func (s *server) GetSTLMesh(ctx context.Context, in *pb.GetSTLMeshRequest) (*pb.GetSTLMeshReply, error) {
	// Load the model binvox file from disk.
	modelFilename := filename(in.GetVoxelRegion())
	base, err := binvox.Read(modelFilename, 0, 0, 0, 0, 0, 0)
	if err != nil {
		return nil, err
	}

	mesh := base.ManifoldMesh()

	var buf bytes.Buffer
	header := gl.STLHeader{}
	header.Count = uint32(len(mesh.Triangles))
	if err := binary.Write(&buf, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	for _, triangle := range mesh.Triangles {
		n := triangle.Normal()
		d := gl.STLTriangle{}
		d.N[0] = float32(n.X)
		d.N[1] = float32(n.Y)
		d.N[2] = float32(n.Z)
		d.V1[0] = float32(triangle.V1.Position.X)
		d.V1[1] = float32(triangle.V1.Position.Y)
		d.V1[2] = float32(triangle.V1.Position.Z)
		d.V2[0] = float32(triangle.V2.Position.X)
		d.V2[1] = float32(triangle.V2.Position.Y)
		d.V2[2] = float32(triangle.V2.Position.Z)
		d.V3[0] = float32(triangle.V3.Position.X)
		d.V3[1] = float32(triangle.V3.Position.Y)
		d.V3[2] = float32(triangle.V3.Position.Z)
		if err := binary.Write(&buf, binary.LittleEndian, &d); err != nil {
			return nil, err
		}
	}
	return &pb.GetSTLMeshReply{}, nil
}

func main() {
	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSTLDiceServer(s, &server{})
	s.Serve(lis)
}

// filename returns the model filename for the given VoxelRegion.
func filename(voxelRegion *pb.VoxelRegion) string {
	return fmt.Sprintf("%v-%v-%v-%v-%v-%v-%v-%v-%v-%v.binvox", modelFilePrefix,
		voxelRegion.GetLlx(), voxelRegion.GetLly(), voxelRegion.GetLlz(),
		voxelRegion.GetUrx(), voxelRegion.GetUry(), voxelRegion.GetUrz(),
		voxelRegion.GetNx(), voxelRegion.GetNy(), voxelRegion.GetNz())
}

// stlToMesh parses STL data and returns a mesh.
func stlToMesh(buf []byte) (*gl.Mesh, error) {
	size := int64(len(buf))
	f := bytes.NewReader(buf)

	// read header, get expected binary size
	header := gl.STLHeader{}
	if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	expectedSize := int64(header.Count)*50 + 84

	// rewind to start of the buffer
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}

	// parse ascii or binary stl
	if size == expectedSize {
		return loadSTLB(f)
	}
	return loadSTLA(f)
}

// parse ASCII STL data.
func loadSTLA(r io.Reader) (*gl.Mesh, error) {
	var vertexes []gl.Vector
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 4 && fields[0] == "vertex" {
			f := gl.ParseFloats(fields[1:])
			vertexes = append(vertexes, gl.Vector{f[0], f[1], f[2]})
		}
	}
	var triangles []*gl.Triangle
	for i := 0; i < len(vertexes); i += 3 {
		t := gl.Triangle{}
		t.V1.Position = vertexes[i+0]
		t.V2.Position = vertexes[i+1]
		t.V3.Position = vertexes[i+2]
		t.FixNormals()
		triangles = append(triangles, &t)
	}
	return gl.NewTriangleMesh(triangles), scanner.Err()
}

// parse binary STL data.
func loadSTLB(r io.Reader) (*gl.Mesh, error) {
	header := gl.STLHeader{}
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	count := int(header.Count)
	triangles := make([]*gl.Triangle, count)
	for i := 0; i < count; i++ {
		d := gl.STLTriangle{}
		if err := binary.Read(r, binary.LittleEndian, &d); err != nil {
			return nil, err
		}
		t := gl.Triangle{}
		t.V1.Position = gl.Vector{float64(d.V1[0]), float64(d.V1[1]), float64(d.V1[2])}
		t.V2.Position = gl.Vector{float64(d.V2[0]), float64(d.V2[1]), float64(d.V2[2])}
		t.V3.Position = gl.Vector{float64(d.V3[0]), float64(d.V3[1]), float64(d.V3[2])}
		t.FixNormals()
		triangles[i] = &t
	}
	return gl.NewTriangleMesh(triangles), nil
}

// voxelize voxelizes a mesh using the provided VoxelRegion.
func voxelize(voxelRegion *pb.VoxelRegion, mesh *gl.Mesh) (*binvox.BinVOX, error) {
	// TODO(gmlewis); write this.
	return nil, nil
}

// voxelOp performs a boolean union or a boolean difference on the two meshes and returns the result.
func voxelOp(base, opVoxels *binvox.BinVOX, union bool) *binvox.BinVOX {
	return nil
}
