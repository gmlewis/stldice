// stldice-master is a gRPC server that listens for commands from the
// dicer command-line tool and communicates with stldice-slave servers
// to process the requests.
package main

import (
	"flag"
	"log"
	"net"

	pb "github.com/gmlewis/stldice/v2/stldice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port = flag.String("port", ":5334", "Port to listen on")
)

// server is used to implement stldice.STLDiceServer
type server struct{}

// Add an STL mesh to the voxels.
func (s *server) AddSTLMesh(ctx context.Context, in *pb.AddSTLMeshRequest) (*pb.AddSTLMeshReply, error) {
	return &pb.AddSTLMeshReply{}, nil
}

// Subtract an STL mesh from the voxels.
func (s *server) SubSTLMesh(ctx context.Context, in *pb.SubSTLMeshRequest) (*pb.SubSTLMeshReply, error) {
	return &pb.SubSTLMeshReply{}, nil
}

// Get back the model voxels as STL.
// This also resets the server by clearing out the model.
func (s *server) GetSTLMesh(ctx context.Context, in *pb.GetSTLMeshRequest) (*pb.GetSTLMeshReply, error) {
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
