// +build ignore

// gen-testdata generates an STL file for the given voxels.
// voxel coordinates are given on the command line.
// 0,0,0 is *always* set, and is the central voxel.
// files are written in the format testdata/golden-010-0_10.stl.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	gl "github.com/fogleman/fauxgl"
	"github.com/gmlewis/stldice/binvox"
)

const (
	filePrefix = "testdata/golden"
)

var (
	compactRE    = regexp.MustCompile(`^g(\d+)$`)
	digitsRE     = regexp.MustCompile(`^x?(0|1|-1),?y?(0|1|-1),?z?(0|1|-1)$`)
	dumpTemplate = template.Must(template.New("dump").Parse(meshTemplate))
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage of gen-testdata:
gen-testdata can be used either to generate a single test file:
    go run gen-testdata.go -- '-1,0,1' '0,1,1'
or regenerate all the testdata at once:
    go run gen-testsdata.go testdata/golden*.stl

Options:
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		if err := processVoxels(filePrefix); err != nil { // golden.stl - single voxel
			log.Fatalf("processVoxels(): %v", err)
		}
		log.Print("Done.")
		return
	}

	// try testing the first argument as the compact form (e.g. "g0127")
	m := compactRE.FindStringSubmatch(flag.Arg(0))
	if len(m) == 2 {
		log.Printf("arguments are not filenames - processing compact form...")
		if err := processCompactForm(filePrefix+m[1]+".stl", m[1]); err != nil {
			log.Fatalf("processCompactForm(%v): %v", m[1], err)
		}
		log.Print("Done.")
		return
	}

	// try testing the first argument as a filename.
	if _, _, err := parseFilename(flag.Arg(0)); err != nil {
		log.Printf("arguments are not filenames - processing voxels...")
		if err := processVoxels(filePrefix, flag.Args()...); err != nil {
			log.Fatalf("processVoxels(%+v): %v", flag.Args(), err)
		}
	} else {
		for _, arg := range flag.Args() {
			log.Printf("\n\nProcessing %v ...", arg)
			prefix, args, err := parseFilename(arg)
			if err != nil {
				log.Fatalf("parseFilename(%q): %v", arg, err)
			}
			if err := processVoxels(prefix, args...); err != nil {
				log.Fatalf("processVoxels(%+v): %v", args, err)
			}
		}
	}
	log.Print("Done.")
}

func parseFilename(arg string) (prefix string, args []string, err error) {
	if !strings.HasSuffix(arg, ".stl") {
		return "", nil, fmt.Errorf("not a filename: %v", arg)
	}
	parts := strings.Split(strings.TrimSuffix(arg, ".stl"), "-")
	if len(parts) == 1 { // e.g. golden.stl - just the center voxel.
		return parts[0], nil, nil
	}
	for _, part := range parts[1:] {
		args = append(args, strings.Replace(part, "_", "-", -1))
	}
	return parts[0], args, nil
}

func processCompactForm(filename string, digits string) error {
	voxels := []gl.Voxel{}
	for _, d := range digits {
		switch d {
		case '0':
			voxels = append(voxels, gl.Voxel{X: 0, Y: 0, Z: 0, Color: gl.White})
		case '1':
			voxels = append(voxels, gl.Voxel{X: 1, Y: 0, Z: 0, Color: gl.White})
		case '2':
			voxels = append(voxels, gl.Voxel{X: 1, Y: -1, Z: 0, Color: gl.White})
		case '3':
			voxels = append(voxels, gl.Voxel{X: 0, Y: -1, Z: 0, Color: gl.White})
		case '4':
			voxels = append(voxels, gl.Voxel{X: 0, Y: 0, Z: 1, Color: gl.White})
		case '5':
			voxels = append(voxels, gl.Voxel{X: 1, Y: 0, Z: 1, Color: gl.White})
		case '6':
			voxels = append(voxels, gl.Voxel{X: 1, Y: -1, Z: 1, Color: gl.White})
		case '7':
			voxels = append(voxels, gl.Voxel{X: 0, Y: -1, Z: 1, Color: gl.White})
		}
	}

	bv := &binvox.BinVOX{Scale: 0, Voxels: voxels} // 1 vpmm
	mesh := bv.ManifoldMesh()

	// Sort mesh triangles so that the testdata/golden* files don't change.
	sort.Slice(mesh.Triangles, binvox.TriangleLess(mesh.Triangles))

	// dumpMesh(keys, mesh)

	log.Printf("Writing file %v (%v triangles)...", filename, len(mesh.Triangles))
	if err := mesh.SaveSTL(filename); err != nil {
		return fmt.Errorf("SaveSTL: %v", err)
	}
	return nil
}

func processVoxels(filePrefix string, args ...string) error {
	voxels := []gl.Voxel{{X: 0, Y: 0, Z: 0, Color: gl.White}}
	keys := []key{} // {0,0,0} is implicit in the unit tests

	filename := filePrefix
	for _, arg := range args {
		v, err := getVoxel(arg)
		if err != nil {
			return fmt.Errorf("getVoxel(%q): %v", arg, err)
		}
		arg = strings.Replace(arg, "-", "_", -1)
		arg = strings.Replace(arg, ",", "", -1)
		arg = strings.Replace(arg, "x", "", -1)
		arg = strings.Replace(arg, "y", "", -1)
		arg = strings.Replace(arg, "z", "", -1)
		filename += "-" + arg
		voxels = append(voxels, *v)
		keys = append(keys, key{X: v.X, Y: v.Y, Z: v.Z})
	}
	filename += ".stl"

	bv := &binvox.BinVOX{Scale: 0, Voxels: voxels} // 1 vpmm
	mesh := bv.ManifoldMesh()

	// Sort mesh triangles so that the testdata/golden* files don't change.
	sort.Slice(mesh.Triangles, binvox.TriangleLess(mesh.Triangles))

	dumpMesh(keys, mesh)

	log.Printf("Writing file %v (%v triangles)...", filename, len(mesh.Triangles))
	if err := mesh.SaveSTL(filename); err != nil {
		return fmt.Errorf("SaveSTL: %v", err)
	}
	return nil
}

func getVoxel(s string) (*gl.Voxel, error) {
	s = strings.Replace(s, "_", "-", -1)
	parts := digitsRE.FindStringSubmatch(s)
	if len(parts) != 4 {
		return nil, fmt.Errorf("unable to parse %q", s)
	}
	x, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse %q", parts[1])
	}
	y, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("unable to parse %q", parts[2])
	}
	z, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("unable to parse %q", parts[3])
	}
	return &gl.Voxel{X: x, Y: y, Z: z, Color: gl.White}, nil
}

func dumpMesh(keys []key, mesh *gl.Mesh) {
	out := &struct {
		Keys []key
		Mesh *gl.Mesh
	}{
		Keys: keys,
		Mesh: mesh,
	}
	if err := dumpTemplate.Execute(os.Stdout, out); err != nil {
		log.Printf("dumpTemplate.Execute: %v", err)
	}
}

var meshTemplate = `{
  keys: []key{ {{- range .Keys -}}{ {{- .X}},{{.Y}},{{.Z -}} },{{end -}} },
  want: []*gl.Triangle{
		{{- range .Mesh.Triangles}}
    gl.NewTriangleForPoints(gl.V({{.V1.Position.X}}, {{.V1.Position.Y}}, {{.V1.Position.Z}}), gl.V({{.V2.Position.X}}, {{.V2.Position.Y}}, {{.V2.Position.Z}}), gl.V({{.V3.Position.X}}, {{.V3.Position.Y}}, {{.V3.Position.Z}})),
    {{- end}}
  },
},
`
