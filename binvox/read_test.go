package binvox

import (
	"bytes"
	"reflect"
	"testing"

	gl "github.com/fogleman/fauxgl"
)

func TestRead(t *testing.T) {
	const header = `#binvox 1
dim 256 256 256
translate -80 -80 -2.6
scale 160
data
`

	tests := []struct {
		header     string
		bytes      []byte
		sx, sy, sz int
		cx, cy, cz int
		want       *BinVOX
	}{
		{ // no data
			header: header,
			want:   &BinVOX{NX: 256, NY: 256, NZ: 256, TX: -80, TY: -80, TZ: -2.6, Scale: 160},
		},
		{ // big design
			header: header,
			bytes:  []byte{0, 5, 1, 3},
			want: &BinVOX{
				NX: 256, NY: 256, NZ: 256, TX: -80, TY: -80, TZ: -2.6, Scale: 160,
				Voxels: []gl.Voxel{
					{X: 0, Y: 5, Z: 0, Color: gl.White},
					{X: 0, Y: 6, Z: 0, Color: gl.White},
					{X: 0, Y: 7, Z: 0, Color: gl.White},
				},
			},
		},
		{ // small design
			header: `#binvox 1
dim 2 2 2
translate -80 -80 -2.6
scale 160
data
`,
			bytes: []byte{0, 5, 1, 3},
			want: &BinVOX{
				NX: 2, NY: 2, NZ: 2, TX: -80, TY: -80, TZ: -2.6, Scale: 160,
				Voxels: []gl.Voxel{
					{X: 1, Y: 1, Z: 0, Color: gl.White},
					{X: 1, Y: 0, Z: 1, Color: gl.White},
					{X: 1, Y: 1, Z: 1, Color: gl.White},
				},
			},
		},
		{ // partial read of small design
			header: `#binvox 1
dim 2 2 2
translate -80 -80 -2.6
scale 160
data
`,
			bytes: []byte{0, 5, 1, 3},
			sx:    1, sy: 1, sz: 1,
			cx: 1, cy: 1, cz: 1,
			want: &BinVOX{
				NX: 2, NY: 2, NZ: 2, TX: -80, TY: -80, TZ: -2.6, Scale: 160,
				Voxels: []gl.Voxel{
					{X: 1, Y: 1, Z: 1, Color: gl.White},
				},
			},
		},
	}

	for _, test := range tests {
		b := bytes.NewBufferString(test.header)
		for _, v := range test.bytes {
			if err := b.WriteByte(v); err != nil {
				t.Fatalf("WriteByte(%v): %v", v, err)
			}
		}
		got, err := read(b, test.sx, test.sy, test.sz, test.cx, test.cy, test.cz)
		if err != nil {
			t.Errorf("read(%q, %+v) = %v, want nil", test.header, test.bytes, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("read(%q, %+v) = %#v, want %#v", test.header, test.bytes, got, test.want)
		}
	}
}
