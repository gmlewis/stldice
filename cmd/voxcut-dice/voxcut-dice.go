// voxcut-dice reads a binvox model and writes commands to stdout to dice
// the model into subsections to be processed by voxcut.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gmlewis/stldice/v2/binvox"
)

var (
	dryRun       = flag.Bool("n", false, "Perform a dry run and print the commands instead of executing them")
	size         = flag.Int("size", 256, "Maximum size for each dimension when dicing")
	binVOXPrefix = flag.String("obinvox", "", "Prefix for output binvox files")
	stlPrefix    = flag.String("ostl", "", "Prefix for output stl files")
	voxPrefix    = flag.String("ovox", "", "Prefix for output vox files")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\t%v [options] base.binvox all-cuts.binvox ...\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Must supply at least one filename")
	}
	if *binVOXPrefix == "" && *stlPrefix == "" && *voxPrefix == "" {
		log.Fatal("Must specify at least one of -obinvox, -ostl, or -ovox")
	}

	base, err := binvox.Read(flag.Arg(0), 0, 0, 0, 1, 1, 1) // Read only one voxel to get the header info.
	if err != nil {
		log.Fatal(err)
	}

	if *size > base.NX && *size > base.NY && *size > base.NZ {
		log.Fatalf("size (%v) larger than all dimensions (%v,%v,%v) - no need to dice", *size, base.NX, base.NY, base.NZ)
	}

	if *dryRun {
		fmt.Println("#!/bin/bash -x")
	}

	sizeArg := fmt.Sprintf("%04d", *size)
	for zi := 0; zi < base.NZ; zi += *size {
		for yi := 0; yi < base.NY; yi += *size {
			for xi := 0; xi < base.NX; xi += *size {
				cmdLine := fmt.Sprintf("voxcut -sx %v -sy %v -sz %v -cx %[4]v -cy %[4]v -cz %[4]v", xi, yi, zi, *size)
				args := strings.Split(cmdLine, " ")
				if *voxPrefix != "" {
					args = append(args, "-ovox")
					args = append(args, fmt.Sprintf("%v-%04d-%04d-%04d-%[5]v-%[5]v-%[5]v.vox", *voxPrefix, xi, yi, zi, sizeArg))
				}
				if *binVOXPrefix != "" {
					args = append(args, "-obinvox")
					args = append(args, fmt.Sprintf("%v-%04d-%04d-%04d-%[5]v-%[5]v-%[5]v.binvox", *binVOXPrefix, xi, yi, zi, sizeArg))
				}
				if *stlPrefix != "" {
					args = append(args, "-ostl")
					args = append(args, fmt.Sprintf("%v-%04d-%04d-%04d-%[5]v-%[5]v-%[5]v.stl", *stlPrefix, xi, yi, zi, sizeArg))
				}
				args = append(args, flag.Args()...)

				if *dryRun {
					fmt.Println(strings.Join(args, " "))
					continue
				}
				log.Printf("\n\nRunning: %v", strings.Join(args, " "))
				cmd := exec.Command(args[0], args[1:]...)
				buf, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatalf("Unable to run command: %v\n%s", err, buf)
				}
				log.Printf("%s\n", buf)
			}
		}
	}
}
