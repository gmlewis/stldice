// stl2svx-server contains the master and agent services
// that are run remotely.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/gmlewis/stldice/v3/stl2svx/agent"
	"github.com/gmlewis/stldice/v3/stl2svx/master"
	pb "github.com/gmlewis/stldice/v3/stl2svx/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	maxSize = 1 << 30 // 1GB
)

var (
	port          = flag.String("port", "0.0.0.0:0", "Port to use (0.0.0.0:0 is any available port)")
	masterAddress = flag.String("master", "", "Address used by agent to contact master")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.MaxRecvMsgSize(maxSize), grpc.MaxSendMsgSize(maxSize))

	switch {
	case *masterAddress == "": // This is the master
		log.Printf("Master using port: %v", lis.Addr().(*net.TCPAddr).String())
		pb.RegisterMasterServer(s, master.New())
	default: // This is an agent
		address := getAddress(lis.Addr().(*net.TCPAddr))
		log.Printf("Agent using port: %v, master address: %v", address, *masterAddress)
		agent, err := agent.New(context.Background(), address, *masterAddress)
		if err != nil {
			log.Fatal(err)
		}
		pb.RegisterAgentServer(s, agent)
	}
	s.Serve(lis)
}

func getAddress(tcpAddr *net.TCPAddr) string {
	localPort := tcpAddr.String()
	port := strings.Split(localPort, ":")

	var result string
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("net.Interfaces: %v", err)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("Addrs: %v", err)
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
				result = ip.String()
				if !strings.HasPrefix(result, "127.") && !strings.Contains(result, ":") {
					return fmt.Sprintf("%v:%v", result, port[len(port)-1])
				}
			case *net.IPAddr:
				ip = v.IP
				result = ip.String()
			}
			log.Printf("got address: %v", result)
		}
	}
	return fmt.Sprintf("%v:%v", result, port[len(port)-1])
}
