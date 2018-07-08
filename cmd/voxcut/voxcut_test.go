package main

import (
	"reflect"
	"testing"

	gl "github.com/fogleman/fauxgl"
	"github.com/gmlewis/stldice/binvox"
)

func TestVoxCut(t *testing.T) {
	tests := []struct {
		base []gl.Voxel
		cut  []gl.Voxel
		want []gl.Voxel
	}{
		{
			base: []gl.Voxel{{0, 0, 0, gl.Black}}, // nothing
			cut:  []gl.Voxel{{0, 0, 0, gl.Black}}, // minus nothing
			want: nil,                             // is nothing
		},
		{
			base: []gl.Voxel{{0, 0, 0, gl.Black}}, // nothing
			cut:  []gl.Voxel{{0, 0, 0, gl.White}}, // minus something
			want: nil,                             // is still nothing
		},
		{
			base: []gl.Voxel{{0, 0, 0, gl.White}}, // something
			cut:  []gl.Voxel{{0, 0, 0, gl.Black}}, // minus nothing
			want: []gl.Voxel{{0, 0, 0, gl.White}}, // is still something
		},
		{
			base: []gl.Voxel{{0, 0, 0, gl.White}}, // something
			cut:  []gl.Voxel{{0, 0, 0, gl.White}}, // minus something
			want: nil,                             // is nothing
		},
	}

	for i, test := range tests {
		base := &binvox.BinVOX{Scale: 1, Voxels: test.base}
		cut := &binvox.BinVOX{Scale: 1, Voxels: test.cut}
		got, err := voxcut(base, cut)
		if err != nil {
			t.Errorf("test #%v: voxcut = %v, want nil", i, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("test #%v: voxcut = %+v, want %+v", i, got, test.want)
		}
	}
}
