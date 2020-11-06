package main

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/gmlewis/stldice/v4/binvox"
)

func TestParseManifest(t *testing.T) {
	data := `<?xml version="1.0"?>

<grid version="1.0" gridSizeX="660" gridSizeY="150" gridSizeZ="150"
   voxelSize="0.000200" subvoxelBits="8" slicesOrientation="Z" >

    <channels>
        <channel type="DENSITY" bits="8" slices="density/slice%04d.png" />
    </channels>

    <materials>
        <material id="1" urn="urn:shapeways:materials/1" />
    </materials>

    <metadata>
        <entry key="author" value="Glenn M. Lewis" />
        <entry key="creationDate" value="2020-10-12" />
    </metadata>
</grid>`

	xmlFile := bytes.NewBufferString(data)
	gotModel, gotSliceNameFmt, err := parseManifest(xmlFile)
	if err != nil {
		t.Fatalf("parseManifest: %v", err)
	}

	if want := "density/slice%04d.png"; gotSliceNameFmt != want {
		t.Errorf("sliceNameFmt = %v, want %v", gotSliceNameFmt, want)
	}

	want := &binvox.BinVOX{
		NX:    660,
		NY:    150,
		NZ:    150,
		Scale: 0.2,

		WhiteVoxels: binvox.WhiteVoxelMap{},
	}
	if !reflect.DeepEqual(gotModel, want) {
		t.Errorf("parseManifest =\n%#v\nwant:\n%#v", *gotModel, *want)
	}
}
