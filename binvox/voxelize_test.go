package binvox

import (
	"reflect"
	"sort"
	"testing"

	gl "github.com/fogleman/fauxgl"
)

func TestVoxelize(t *testing.T) {
	cube2x2x2 := gl.NewCube()

	tests := []struct {
		bv   *BinVOX
		want []gl.Voxel
	}{
		{
			bv:   &BinVOX{NX: 1, NY: 1, NZ: 1, TX: -1, TY: -1, TZ: -1, Scale: 2},
			want: []gl.Voxel{{X: 0, Y: 0, Z: 0, Color: gl.White}},
		},
		{
			bv: &BinVOX{NX: 2, NY: 2, NZ: 2, TX: -1, TY: -1, TZ: -1, Scale: 2},
			want: []gl.Voxel{
				{X: 0, Y: 0, Z: 0, Color: gl.White},
				{X: 1, Y: 0, Z: 0, Color: gl.White},
				{X: 0, Y: 1, Z: 0, Color: gl.White},
				{X: 1, Y: 1, Z: 0, Color: gl.White},
				{X: 0, Y: 0, Z: 1, Color: gl.White},
				{X: 1, Y: 0, Z: 1, Color: gl.White},
				{X: 0, Y: 1, Z: 1, Color: gl.White},
				{X: 1, Y: 1, Z: 1, Color: gl.White},
			},
		},
	}

	for _, test := range tests {
		if err := test.bv.Voxelize(cube2x2x2); err != nil {
			t.Fatalf("Voxelize(%v): %v", test.bv, err)
		}

		got := sortVoxels(test.bv.Voxels)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Voxelize(%v) =\n%#v\nwant\n%#v", test.bv, got, test.want)
		}
	}
}

func sortVoxels(voxels VoxelMap) []gl.Voxel {
	var result []gl.Voxel
	for v := range voxels {
		result = append(result, gl.Voxel{X: v.X, Y: v.Y, Z: v.Z, Color: gl.White})
	}

	sort.Slice(result, func(a, b int) bool {
		if result[a].Z < result[b].Z {
			return true
		}
		if result[a].Z > result[b].Z {
			return false
		}
		if result[a].Y < result[b].Y {
			return true
		}
		if result[a].Y > result[b].Y {
			return false
		}
		return result[a].X < result[b].X
	})

	return result
}
