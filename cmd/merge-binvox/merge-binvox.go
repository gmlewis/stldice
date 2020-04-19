// merge-binvox takes a filename prefix and reads in all files matching
// prefix*.binvox. It merges the binvox files together as VShells, then writes
// out prefix.vsh as a merged model.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gmlewis/stldice/vshell"
)

var (
	force  = flag.Bool("f", false, "Force overwrite of output file")
	prefix = flag.String("prefix", "out", "Prefix of files to merge and name of .vsh file to write (default='out')")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\t%v -prefix out\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	files, err := filepath.Glob(*prefix + "*.binvox")
	if err != nil {
		log.Fatalf("No files found matching %v*.binvox", *prefix)
	}
	outFile := *prefix + ".vsh"
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

	log.Printf("Merging %v binvox files into %v...", len(newFiles), outFile)
	vsh, err := vshell.Merge(newFiles)
	if err != nil {
		log.Fatalf("Merge: %v", err)
	}

	if err := vsh.Write(outFile, 0, 0, 0, 0, 0, 0); err != nil {
		log.Fatalf("Save: %v", err)
	}

	log.Printf("Done.")
}
