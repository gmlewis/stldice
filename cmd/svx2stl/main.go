// svx2stl reads a .svx file and writes a simple
// .stl file using the marching cubes algorithm.
//
// Usage:
//   svx2stl infile.svx ...
package main

import (
	"archive/zip"
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gmlewis/stldice/v4/binvox"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\t%v infile.svx ...\n", filepath.Base(os.Args[0]))
	}
	flag.Parse()

	for _, svxFile := range flag.Args() {
		stlFile := strings.TrimSuffix(svxFile, ".svx") + ".stl"

		model, err := read(svxFile)
		if err != nil {
			log.Fatal(err)
		}

		mesh := model.MarchingCubes()
		log.Printf("Writing file %q...", stlFile)
		if err := mesh.SaveSTL(stlFile); err != nil {
			log.Fatalf("SaveSTL: %v", err)
		}
	}

	log.Println("Done.")
}

func read(filename string) (*binvox.BinVOX, error) {
	log.Printf("Loading file %q...", filename)
	var model *binvox.BinVOX

	r, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("zip.OpenReader: %v", err)
	}
	defer r.Close()

	var sliceNameFmt string
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		if f.Name == "manifest.xml" {
			model, sliceNameFmt, err = parseManifest(rc)
			if err != nil {
				return nil, fmt.Errorf("parseManifest: %v", err)
			}
			rc.Close()
			continue
		}

		var sliceNum int
		if n, err := fmt.Sscanf(f.Name, sliceNameFmt, &sliceNum); err != nil || n != 1 {
			return nil, fmt.Errorf("unable to parse %q using %q", f.Name, sliceNameFmt)
		}

		img, err := png.Decode(rc)
		if err != nil {
			return nil, fmt.Errorf("png.Decode: %v", err)
		}
		rc.Close()

		scanImage(img, model, sliceNum)
	}

	return model, nil
}

func scanImage(img image.Image, model *binvox.BinVOX, z int) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if c.Y >= 128 {
				model.Add(x, y, z)
			}
		}
	}
}

func parseManifest(f io.Reader) (*binvox.BinVOX, string, error) {
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, "", fmt.Errorf("ioutil.Reader: %v", err)
	}

	type Entry struct {
		Key   string `xml:"key,attr"`
		Value string `xml:"value,attr"`
	}

	type Metadata struct {
		Entry []Entry `xml:"entry"`
	}

	type Material struct {
		ID  string `xml:"id,attr"`
		URN string `xml:"urn,attr"`
	}

	type Materials struct {
		Material []Material `xml:"material"`
	}

	type Channel struct {
		Type   string `xml:"type,attr"`
		Bits   int    `xml:"bits,attr"`
		Slices string `xml:"slices,attr"`
	}

	type Channels struct {
		Channel []Channel `xml:"channel"`
	}

	type Grid struct {
		GridSizeX int       `xml:"gridSizeX,attr"`
		GridSizeY int       `xml:"gridSizeY,attr"`
		GridSizeZ int       `xml:"gridSizeZ,attr"`
		VoxelSize float64   `xml:"voxelSize,attr"`
		Channels  Channels  `xml:"channels"`
		Materials Materials `xml:"materials"`
		Metadata  Metadata  `xml:"metadata"`
	}

	type Result struct {
		Grid
	}

	v := Result{}
	if err := xml.Unmarshal(buf, &v); err != nil {
		return nil, "", fmt.Errorf("xml.Unmarshal: %v", err)
	}
	if len(v.Grid.Channels.Channel) == 0 {
		return nil, "", fmt.Errorf("expected at least one grid channel")
	}

	model := &binvox.BinVOX{
		NX:    v.Grid.GridSizeX,
		NY:    v.Grid.GridSizeY,
		NZ:    v.Grid.GridSizeZ,
		Scale: v.Grid.VoxelSize * 1000.0,

		WhiteVoxels: binvox.WhiteVoxelMap{},
	}

	var densitySlices string
	for _, channel := range v.Grid.Channels.Channel {
		if strings.EqualFold(channel.Type, "DENSITY") {
			densitySlices = channel.Slices
		}
	}
	if densitySlices == "" {
		return nil, "", fmt.Errorf("could not find DENSITY slices")
	}

	return model, densitySlices, nil
}
