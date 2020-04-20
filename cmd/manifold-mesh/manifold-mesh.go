// manifold-mesh reads a .binvox file and writes a simple
// .stl file that is topologically closed (manifold).
//
// Usage:
//   manifold-mesh infile.binvox outfile.stl
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gmlewis/stldice/v4/binvox"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\t%v infile.binvox outfile.stl\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatalf("Must provide two filenames: infile.binvox outfile.stl")
	}
	binvoxFile := flag.Arg(0)
	stlFile := flag.Arg(1)

	model, err := binvox.Read(binvoxFile, 0, 0, 0, 0, 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	mesh := model.ManifoldMesh()
	log.Printf("Writing file %q...", stlFile)
	if err := mesh.SaveSTL(stlFile); err != nil {
		log.Fatalf("SaveSTL: %v", err)
	}

	log.Println("Done.")
}
