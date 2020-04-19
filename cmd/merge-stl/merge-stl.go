// merge-stl takes a filename prefix and reads in all files matching prefix*.stl.
// It merges the STL files together, then writes out prefix.stl as a merged model.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gmlewis/stldice/stl"
)

var (
	force      = flag.Bool("f", false, "Force overwrite of output file")
	numWorkers = flag.Int("num", 10, "Number of workers to use for work pool (default=10)")
	prefix     = flag.String("prefix", "out", "Prefix of files to merge and name of .stl file to write (default='out')")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\t%v -prefix out\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	files, err := filepath.Glob(*prefix + "*.stl")
	if err != nil {
		log.Fatalf("No files found matching %v*.stl", *prefix)
	}
	outFile := *prefix + ".stl"
	var newFiles []string
	for _, f := range files { // Quick sanity check
		if f != outFile {
			newFiles = append(newFiles, f)
			continue
		}
		if *force {
			log.Printf("%v already exists. Overwriting due to -f flag.", outFile)
		} else {
			log.Fatalf("%v already exists. To overwrite, use -f flag.", outFile)
		}
	}

	log.Printf("Merging %v STL files into %v using %v workers...", len(newFiles), outFile, *numWorkers)
	mesh, err := stl.Merge(newFiles, *numWorkers)
	if err != nil {
		log.Fatalf("Merge: %v", err)
	}

	if err := mesh.SaveSTL(outFile); err != nil {
		log.Fatalf("SaveSTL: %v", err)
	}

	log.Printf("Done.")
}
