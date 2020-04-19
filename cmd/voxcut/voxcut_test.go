package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gmlewis/stldice/v3/binvox"
)

func TestVoxCut(t *testing.T) {
	tests := []struct {
		base binvox.VoxelMap
		cut  binvox.VoxelMap
		want binvox.VoxelMap
	}{
		{
			base: binvox.VoxelMap{binvox.Key{0, 0, 0}: binvox.Black}, // nothing
			cut:  binvox.VoxelMap{binvox.Key{0, 0, 0}: binvox.Black}, // minus nothing
			want: binvox.VoxelMap{},
		},
		{
			base: binvox.VoxelMap{binvox.Key{0, 0, 0}: binvox.Black}, // nothing
			cut:  binvox.VoxelMap{binvox.Key{0, 0, 0}: binvox.White}, // minus something
			want: binvox.VoxelMap{},                                  // is still nothing
		},
		{
			base: binvox.VoxelMap{binvox.Key{0, 0, 0}: binvox.White}, // something
			cut:  binvox.VoxelMap{binvox.Key{0, 0, 0}: binvox.Black}, // minus nothing
			want: binvox.VoxelMap{binvox.Key{0, 0, 0}: binvox.White}, // is still something
		},
		{
			base: binvox.VoxelMap{binvox.Key{0, 0, 0}: binvox.White}, // something
			cut:  binvox.VoxelMap{binvox.Key{0, 0, 0}: binvox.White}, // minus something
			want: binvox.VoxelMap{},                                  // is nothing
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test #%v", i), func(t *testing.T) {
			base := &binvox.BinVOX{Scale: 1, Voxels: test.base}
			cut := &binvox.BinVOX{Scale: 1, Voxels: test.cut}
			got, err := voxcut(base, cut)
			if err != nil {
				t.Fatalf("test #%v: voxcut = %v, want nil", i, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("test #%v: voxcut = %+v, want %+v", i, got, test.want)
			}
		})
	}
}
