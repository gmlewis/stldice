package binvox

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	const header = `#binvox 1
dim 256 256 256
translate -80 -80 -2.6
scale 160
data
`

	tests := []struct {
		name       string
		header     string
		bytes      []byte
		sx, sy, sz int
		cx, cy, cz int
		want       *BinVOX
	}{
		{
			name:   "no data",
			header: header,
			want:   &BinVOX{NX: 256, NY: 256, NZ: 256, TX: -80, TY: -80, TZ: -2.6, Scale: 160, Voxels: VoxelMap{}},
		},
		{
			name:   "big design",
			header: header,
			bytes:  []byte{0, 5, 1, 3},
			want: &BinVOX{
				NX: 256, NY: 256, NZ: 256, TX: -80, TY: -80, TZ: -2.6, Scale: 160,
				Voxels: VoxelMap{
					Key{X: 0, Y: 5, Z: 0}: White,
					Key{X: 0, Y: 6, Z: 0}: White,
					Key{X: 0, Y: 7, Z: 0}: White,
				},
			},
		},
		{
			name: "small design",
			header: `#binvox 1
dim 2 2 2
translate -80 -80 -2.6
scale 160
data
`,
			bytes: []byte{0, 5, 1, 3},
			want: &BinVOX{
				NX: 2, NY: 2, NZ: 2, TX: -80, TY: -80, TZ: -2.6, Scale: 160,
				Voxels: VoxelMap{
					Key{X: 1, Y: 1, Z: 0}: White,
					Key{X: 1, Y: 0, Z: 1}: White,
					Key{X: 1, Y: 1, Z: 1}: White,
				},
			},
		},
		{
			name: "partial read of small design",
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
				Voxels: VoxelMap{
					Key{X: 1, Y: 1, Z: 1}: White,
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test #%v: %v", i, tt.name), func(t *testing.T) {
			b := bytes.NewBufferString(tt.header)
			for _, v := range tt.bytes {
				if err := b.WriteByte(v); err != nil {
					t.Fatalf("WriteByte(%v): %v", v, err)
				}
			}
			got, err := read(b, tt.sx, tt.sy, tt.sz, tt.cx, tt.cy, tt.cz)
			if err != nil {
				t.Fatalf("read(%q, %+v) = %v, want nil", tt.header, tt.bytes, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("read(%q, %+v) = %#v, want %#v", tt.header, tt.bytes, got, tt.want)
			}
		})
	}
}
