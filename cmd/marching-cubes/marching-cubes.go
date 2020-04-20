// marching-cubes reads a .binvox file and writes a simple
// .stl file using the marching cubes algorithm.
//
// Usage:
//   marching-cubes infile.binvox outfile.stl
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gmlewis/stldice/v4/binvox"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\t%v infile.binvox outfile.stl\n\nOptions:\n", filepath.Base(os.Args[0]))
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

	mesh := model.MarchingCubes()
	log.Printf("Writing file %q...", stlFile)
	if err := mesh.SaveSTL(stlFile); err != nil {
		log.Fatalf("SaveSTL: %v", err)
	}

	log.Println("Done.")
}
