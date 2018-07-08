package binvox

import (
	"reflect"
	"sort"
	"testing"

	gl "github.com/fogleman/fauxgl"
)

func TestVoxelize(t *testing.T) {
	cube2x2x2 := gl.NewCube()
	c := gl.White

	tests := []struct {
		bv   *BinVOX
		want []gl.Voxel
	}{
		{
			bv:   &BinVOX{NX: 1, NY: 1, NZ: 1, TX: -1, TY: -1, TZ: -1, Scale: 2},
			want: []gl.Voxel{{X: 0, Y: 0, Z: 0, Color: c}},
		},
		{
			bv: &BinVOX{NX: 2, NY: 2, NZ: 2, TX: -1, TY: -1, TZ: -1, Scale: 2},
			want: []gl.Voxel{
				{X: 0, Y: 0, Z: 0, Color: c},
				{X: 1, Y: 0, Z: 0, Color: c},
				{X: 0, Y: 1, Z: 0, Color: c},
				{X: 1, Y: 1, Z: 0, Color: c},
				{X: 0, Y: 0, Z: 1, Color: c},
				{X: 1, Y: 0, Z: 1, Color: c},
				{X: 0, Y: 1, Z: 1, Color: c},
				{X: 1, Y: 1, Z: 1, Color: c},
			},
		},
	}

	for _, test := range tests {
		if err := test.bv.Voxelize(cube2x2x2); err != nil {
			t.Fatalf("Voxelize(%v): %v", test.bv, err)
		}
		sortVoxels(test.bv.Voxels)
		if !reflect.DeepEqual(test.bv.Voxels, test.want) {
			t.Errorf("Voxelize(%v) =\n%#v\nwant\n%#v", test.bv, test.bv.Voxels, test.want)
		}
	}
}

func sortVoxels(v []gl.Voxel) {
	sort.Slice(v, func(a, b int) bool {
		if v[a].Z < v[b].Z {
			return true
		}
		if v[a].Z > v[b].Z {
			return false
		}
		if v[a].Y < v[b].Y {
			return true
		}
		if v[a].Y > v[b].Y {
			return false
		}
		return v[a].X < v[b].X
	})
}
