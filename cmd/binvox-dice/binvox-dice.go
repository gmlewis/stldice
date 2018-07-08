// binvox-dice takes an STL bounding box and writes commands to stdout to dice
// the model into subsections to be processed by binvox.
// Note that the maximum reliable dice resolution of binvox is 1024 despite the
// documentation. If the CPU is used, all internal voxels are remove with or
// without the '-ri' flag. This means that the resulting .binvox files are all
// hollow if the '-e' is used.
//
// Usage:
//   binvox-dice -factor 8 -mbb "(-80,-80,-2.6)-(80,80,0.6)" base.stl all-cuts.stl ...
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	gl "github.com/fogleman/fauxgl"
)

const (
	maxD = 1024 // maximum voxel grid size that binvox can reliably calculate
)

var (
	factor    = flag.Int("factor", 8, "Multiplier factor (power of two) for subdividing the STL file")
	mbbString = flag.String("mbb", "", "MBB of the STL file, e.g. '(-80,-80,-2.6)-(80,80,0.6)'")
	subX      = flag.Bool("subx", true, "Subdivide in the X dimension (default=true)")
	subY      = flag.Bool("suby", true, "Subdivide in the Y dimension (default=true)")
	subZ      = flag.Bool("subz", false, "Subdivide in the Z dimension (default=false)")

	mbbRE = regexp.MustCompile(`^\s*[\(\[]?\s*([^,\s]+)\s*,\s*([^,\s]+)\s*,\s*([^\]\)\s]+)\s*[\]\)]?\s*\-\s*[\(\[]?\s*([^,\s]+)\s*,\s*([^,\s]+)\s*,\s*([^\]\)\s]+)\s*[\)\]]?\s*$`)
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\t%v -factor 8 -mbb '(-80,-80,-2.6)-(80,80,0.6)' base.stl all-cuts.stl ...\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Must supply at least one filename")
	}

	parts := mbbRE.FindStringSubmatch(*mbbString)
	if len(parts) != 7 {
		log.Fatalf("Unable to parse mbb %q", *mbbString)
	}
	llx, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		log.Fatalf("unable to parse mbb %v: %v", parts[1], err)
	}
	lly, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		log.Fatalf("unable to parse mbb %v: %v", parts[2], err)
	}
	llz, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		log.Fatalf("unable to parse mbb %v: %v", parts[3], err)
	}
	urx, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		log.Fatalf("unable to parse mbb %v: %v", parts[4], err)
	}
	ury, err := strconv.ParseFloat(parts[5], 64)
	if err != nil {
		log.Fatalf("unable to parse mbb %v: %v", parts[5], err)
	}
	urz, err := strconv.ParseFloat(parts[6], 64)
	if err != nil {
		log.Fatalf("unable to parse mbb %v: %v", parts[6], err)
	}

	fmt.Println("#!/bin/bash -x")
	fmt.Println("# Generated with command line:")
	fmt.Printf("# binvox-dice \\\n#\t-factor %v \\\n#\t-mbb %q \\\n#\t-subx %v \\\n#\t-suby %v \\\n#\t-subz %v \\\n#\t%v\n", *factor, *mbbString, *subX, *subY, *subZ, strings.Join(flag.Args(), " "))

	xf := *factor
	if !*subX {
		xf = 1
	}
	yf := *factor
	if !*subY {
		yf = 1
	}
	zf := *factor
	if !*subZ {
		zf = 1
	}

	sFmt := "voxcut -ostl out%[1]v.stl"
	for _, arg := range flag.Args() {
		a := strings.TrimSuffix(arg, ".stl")
		sFmt += " " + a + "%[1]v.binvox"
	}

	var numCuts int
	for zi := 0; zi < zf; zi++ {
		z1 := gl.V(0, 0, llz).Lerp(gl.V(0, 0, urz), float64(zi)/float64(zf)).Z
		z2 := gl.V(0, 0, llz).Lerp(gl.V(0, 0, urz), float64(zi+1)/float64(zf)).Z
		for yi := 0; yi < yf; yi++ {
			y1 := gl.V(0, 0, lly).Lerp(gl.V(0, 0, ury), float64(yi)/float64(yf)).Z
			y2 := gl.V(0, 0, lly).Lerp(gl.V(0, 0, ury), float64(yi+1)/float64(yf)).Z
			for xi := 0; xi < xf; xi++ {
				x1 := gl.V(0, 0, llx).Lerp(gl.V(0, 0, urx), float64(xi)/float64(xf)).Z
				x2 := gl.V(0, 0, llx).Lerp(gl.V(0, 0, urx), float64(xi+1)/float64(xf)).Z
				fmt.Println()
				for _, arg := range flag.Args() {
					fmt.Printf("binvox -pb -d %v -bb %g %g %g %g %g %g %v\n", maxD, x1, y1, z1, x2, y2, z2, arg)
				}

				// Now output the voxcut commands to process all the .binvox files...
				c := ""
				if numCuts > 0 {
					c = fmt.Sprintf("_%v", numCuts)
				}
				fmt.Printf(sFmt+"\n", c)

				numCuts++
			}
		}
	}
}
