package vshell

import (
	"reflect"
	"sort"
	"testing"

	"github.com/gmlewis/stldice/v2/binvox"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name    string
		vs      *VShell
		bv      *binvox.BinVOX
		want    *VShell
		wantErr bool
	}{
		{name: "null pointers", wantErr: true},
		{
			name:    "mismatched vpmm",
			vs:      &VShell{NX: 1, NY: 1, NZ: 1, Scale: 1, Voxels: []VShVoxel{{}}},
			bv:      binvox.New(1, 1, 1, 0, 0, 0, 2),
			wantErr: true,
		},
		{
			name: "empty voxels",
			vs:   &VShell{Scale: 1},
			bv:   &binvox.BinVOX{Scale: 1},
			want: &VShell{Scale: 1},
		},
		{
			name: "no adjustment necessary",
			vs: &VShell{
				Scale:  1,
				Voxels: []VShVoxel{{}}, // Single voxel at origin
			},
			bv: &binvox.BinVOX{Scale: 1},
			want: &VShell{
				Scale:  1,
				Voxels: []VShVoxel{{}},
			},
		},
		{
			name: "binvox offset by 1,1,1",
			vs: &VShell{
				NX: 1, NY: 1, NZ: 1,
				Scale:  1,
				Voxels: []VShVoxel{{}}, // single voxel at origin
			},
			bv: &binvox.BinVOX{
				NX: 1, NY: 1, NZ: 1,
				TX: 1, TY: 1, TZ: 1,
				Scale:  1,
				Voxels: binvox.VoxelMap{binvox.Key{}: binvox.White}, // single voxel at origin
			},
			want: &VShell{
				NX: 2, NY: 2, NZ: 2,
				Scale:  2, // preserve VPMM
				Voxels: []VShVoxel{{N: X1Y1Z1}, {X: 1, Y: 1, Z: 1, N: X_1Y_1Z_1}},
			},
		},
		{
			name: "vshell update needed",
			vs: &VShell{
				NX: 1, NY: 1, NZ: 1,
				Scale:  1,
				Voxels: []VShVoxel{{}}, // single voxel at origin, no neighbors
			},
			bv: &binvox.BinVOX{
				NX: 1, NY: 1, NZ: 1,
				TX: -1, TY: -1, TZ: -1,
				Scale:  1,
				Voxels: binvox.VoxelMap{binvox.Key{}: binvox.White}, // single voxel at (-1,-1,-1), no neighbors
			},
			want: &VShell{
				NX: 2, NY: 2, NZ: 2, // dimensions updated
				TX: -1, TY: -1, TZ: -1, // translation updated
				Scale:  2, // preserve VPMM
				Voxels: []VShVoxel{{X: 1, Y: 1, Z: 1, N: X_1Y_1Z_1}, {N: X1Y1Z1}},
			},
		},
		{
			name: "remove middle voxel in 3x3x3 cube of voxels",
			vs: &VShell{ // no voxels to begin with
				NX: 3, NY: 3, NZ: 3,
				Scale: 1,
			},
			bv: &binvox.BinVOX{
				NX: 3, NY: 3, NZ: 3,
				Scale: 1,
				Voxels: binvox.VoxelMap{
					binvox.Key{X: 0, Y: 0, Z: 0}: binvox.White,
					binvox.Key{X: 0, Y: 1, Z: 0}: binvox.White,
					binvox.Key{X: 0, Y: 2, Z: 0}: binvox.White,
					binvox.Key{X: 0, Y: 0, Z: 1}: binvox.White,
					binvox.Key{X: 0, Y: 1, Z: 1}: binvox.White,
					binvox.Key{X: 0, Y: 2, Z: 1}: binvox.White,
					binvox.Key{X: 0, Y: 0, Z: 2}: binvox.White,
					binvox.Key{X: 0, Y: 1, Z: 2}: binvox.White,
					binvox.Key{X: 0, Y: 2, Z: 2}: binvox.White,

					binvox.Key{X: 1, Y: 0, Z: 0}: binvox.White,
					binvox.Key{X: 1, Y: 1, Z: 0}: binvox.White,
					binvox.Key{X: 1, Y: 2, Z: 0}: binvox.White,
					binvox.Key{X: 1, Y: 0, Z: 1}: binvox.White,
					binvox.Key{X: 1, Y: 1, Z: 1}: binvox.White,
					binvox.Key{X: 1, Y: 2, Z: 1}: binvox.White,
					binvox.Key{X: 1, Y: 0, Z: 2}: binvox.White,
					binvox.Key{X: 1, Y: 1, Z: 2}: binvox.White,
					binvox.Key{X: 1, Y: 2, Z: 2}: binvox.White,

					binvox.Key{X: 2, Y: 0, Z: 0}: binvox.White,
					binvox.Key{X: 2, Y: 1, Z: 0}: binvox.White,
					binvox.Key{X: 2, Y: 2, Z: 0}: binvox.White,
					binvox.Key{X: 2, Y: 0, Z: 1}: binvox.White,
					binvox.Key{X: 2, Y: 1, Z: 1}: binvox.White,
					binvox.Key{X: 2, Y: 2, Z: 1}: binvox.White,
					binvox.Key{X: 2, Y: 0, Z: 2}: binvox.White,
					binvox.Key{X: 2, Y: 1, Z: 2}: binvox.White,
					binvox.Key{X: 2, Y: 2, Z: 2}: binvox.White,
				},
			},
			want: &VShell{
				NX: 3, NY: 3, NZ: 3,
				Scale: 1,
				Voxels: []VShVoxel{
					{X: 0, Y: 0, Z: 0, N: X0Y0Z1 | X0Y1Z0 | X0Y1Z1 | X1Y0Z0 | X1Y0Z1 | X1Y1Z0 | X1Y1Z1},
					{X: 1, Y: 0, Z: 0, N: X_1Y0Z0 | X_1Y0Z1 | X_1Y1Z0 | X_1Y1Z1 | X0Y0Z1 | X0Y1Z0 | X0Y1Z1 | X1Y0Z0 | X1Y0Z1 | X1Y1Z0 | X1Y1Z1},
					{X: 2, Y: 0, Z: 0, N: X_1Y0Z0 | X_1Y0Z1 | X_1Y1Z0 | X_1Y1Z1 | X0Y0Z1 | X0Y1Z0 | X0Y1Z1},
					{X: 0, Y: 1, Z: 0, N: X0Y_1Z0 | X0Y_1Z1 | X0Y0Z1 | X0Y1Z0 | X0Y1Z1 | X1Y_1Z0 | X1Y_1Z1 | X1Y0Z0 | X1Y0Z1 | X1Y1Z0 | X1Y1Z1},
					{X: 1, Y: 1, Z: 0, N: allNeighbors & ^X_1Y_1Z_1 & ^X_1Y0Z_1 & ^X_1Y1Z_1 & ^X0Y_1Z_1 & ^X0Y0Z_1 & ^X0Y1Z_1 & ^X1Y_1Z_1 & ^X1Y0Z_1 & ^X1Y1Z_1},
					{X: 2, Y: 1, Z: 0, N: X_1Y_1Z0 | X_1Y_1Z1 | X_1Y0Z0 | X_1Y0Z1 | X_1Y1Z0 | X_1Y1Z1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z1 | X0Y1Z0 | X0Y1Z1},
					{X: 0, Y: 2, Z: 0, N: X0Y_1Z0 | X0Y_1Z1 | X0Y0Z1 | X1Y_1Z0 | X1Y_1Z1 | X1Y0Z0 | X1Y0Z1},
					{X: 1, Y: 2, Z: 0, N: X_1Y_1Z0 | X_1Y_1Z1 | X_1Y0Z0 | X_1Y0Z1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z1 | X1Y_1Z0 | X1Y_1Z1 | X1Y0Z0 | X1Y0Z1},
					{X: 2, Y: 2, Z: 0, N: X_1Y_1Z0 | X_1Y_1Z1 | X_1Y0Z0 | X_1Y0Z1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z1},
					{X: 0, Y: 0, Z: 1, N: X0Y0Z_1 | X0Y0Z1 | X0Y1Z_1 | X0Y1Z0 | X0Y1Z1 | X1Y0Z_1 | X1Y0Z0 | X1Y0Z1 | X1Y1Z_1 | X1Y1Z0 | X1Y1Z1},
					{X: 1, Y: 0, Z: 1, N: allNeighbors & ^X_1Y_1Z_1 & ^X_1Y_1Z0 & ^X_1Y_1Z1 & ^X0Y_1Z_1 & ^X0Y_1Z0 & ^X0Y_1Z1 & ^X1Y_1Z_1 & ^X1Y_1Z0 & ^X1Y_1Z1},
					{X: 2, Y: 0, Z: 1, N: X_1Y0Z_1 | X_1Y0Z0 | X_1Y0Z1 | X_1Y1Z_1 | X_1Y1Z0 | X_1Y1Z1 | X0Y0Z_1 | X0Y0Z1 | X0Y1Z_1 | X0Y1Z0 | X0Y1Z1},
					{X: 0, Y: 1, Z: 1, N: allNeighbors & ^X_1Y_1Z_1 & ^X_1Y_1Z0 & ^X_1Y_1Z1 & ^X_1Y0Z_1 & ^X_1Y0Z0 & ^X_1Y0Z1 & ^X_1Y1Z_1 & ^X_1Y1Z0 & ^X_1Y1Z1},
					// removed: {X: 1, Y: 1, Z: 1},
					{X: 2, Y: 1, Z: 1, N: allNeighbors & ^X1Y_1Z_1 & ^X1Y_1Z0 & ^X1Y_1Z1 & ^X1Y0Z_1 & ^X1Y0Z0 & ^X1Y0Z1 & ^X1Y1Z_1 & ^X1Y1Z0 & ^X1Y1Z1},
					{X: 0, Y: 2, Z: 1, N: X0Y_1Z_1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z_1 | X0Y0Z1 | X1Y_1Z_1 | X1Y_1Z0 | X1Y_1Z1 | X1Y0Z_1 | X1Y0Z0 | X1Y0Z1},
					{X: 1, Y: 2, Z: 1, N: allNeighbors & ^X_1Y1Z_1 & ^X_1Y1Z0 & ^X_1Y1Z1 & ^X0Y1Z_1 & ^X0Y1Z0 & ^X0Y1Z1 & ^X1Y1Z_1 & ^X1Y1Z0 & ^X1Y1Z1},
					{X: 2, Y: 2, Z: 1, N: X_1Y_1Z_1 | X_1Y_1Z0 | X_1Y_1Z1 | X_1Y0Z_1 | X_1Y0Z0 | X_1Y0Z1 | X0Y_1Z_1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z_1 | X0Y0Z1},
					{X: 0, Y: 0, Z: 2, N: X0Y0Z_1 | X0Y1Z_1 | X0Y1Z0 | X1Y0Z_1 | X1Y0Z0 | X1Y1Z_1 | X1Y1Z0},
					{X: 1, Y: 0, Z: 2, N: X_1Y0Z_1 | X_1Y0Z0 | X_1Y1Z_1 | X_1Y1Z0 | X0Y0Z_1 | X0Y1Z_1 | X0Y1Z0 | X1Y0Z_1 | X1Y0Z0 | X1Y1Z_1 | X1Y1Z0},
					{X: 2, Y: 0, Z: 2, N: X_1Y0Z_1 | X_1Y0Z0 | X_1Y1Z_1 | X_1Y1Z0 | X0Y0Z_1 | X0Y1Z_1 | X0Y1Z0},
					{X: 0, Y: 1, Z: 2, N: X0Y_1Z_1 | X0Y_1Z0 | X0Y0Z_1 | X0Y1Z_1 | X0Y1Z0 | X1Y_1Z_1 | X1Y_1Z0 | X1Y0Z_1 | X1Y0Z0 | X1Y1Z_1 | X1Y1Z0},
					{X: 1, Y: 1, Z: 2, N: allNeighbors & ^X_1Y_1Z1 & ^X_1Y0Z1 & ^X_1Y1Z1 & ^X0Y_1Z1 & ^X0Y0Z1 & ^X0Y1Z1 & ^X1Y_1Z1 & ^X1Y0Z1 & ^X1Y1Z1},
					{X: 2, Y: 1, Z: 2, N: X_1Y_1Z_1 | X_1Y_1Z0 | X_1Y0Z_1 | X_1Y0Z0 | X_1Y1Z_1 | X_1Y1Z0 | X0Y_1Z_1 | X0Y_1Z0 | X0Y0Z_1 | X0Y1Z_1 | X0Y1Z0},
					{X: 0, Y: 2, Z: 2, N: X0Y_1Z_1 | X0Y_1Z0 | X0Y0Z_1 | X1Y_1Z_1 | X1Y_1Z0 | X1Y0Z_1 | X1Y0Z0},
					{X: 1, Y: 2, Z: 2, N: X_1Y_1Z_1 | X_1Y_1Z0 | X_1Y0Z_1 | X_1Y0Z0 | X0Y_1Z_1 | X0Y_1Z0 | X0Y0Z_1 | X1Y_1Z_1 | X1Y_1Z0 | X1Y0Z_1 | X1Y0Z0},
					{X: 2, Y: 2, Z: 2, N: X_1Y_1Z_1 | X_1Y_1Z0 | X_1Y0Z_1 | X_1Y0Z0 | X0Y_1Z_1 | X0Y_1Z0 | X0Y0Z_1},
				},
			},
		},
		{
			name: "remove previous vshell voxel due to abutted edge voxels",
			vs: &VShell{
				NX: 1, NY: 3, NZ: 3,
				Scale: 1,
				Voxels: []VShVoxel{ // a right-facing "wall" of 9 voxels with no neighbors to the right - all will be removed due to them only missing neighbors to the right.
					{X: 0, Y: 0, Z: 0, N: allNeighbors & ^X1Y0Z0},
					{X: 0, Y: 1, Z: 0, N: allNeighbors & ^X1Y0Z0},
					{X: 0, Y: 2, Z: 0, N: allNeighbors & ^X1Y0Z0},
					{X: 0, Y: 0, Z: 1, N: allNeighbors & ^X1Y0Z0},
					{X: 0, Y: 1, Z: 1, N: allNeighbors & ^X1Y0Z0},
					{X: 0, Y: 2, Z: 1, N: allNeighbors & ^X1Y0Z0},
					{X: 0, Y: 0, Z: 2, N: allNeighbors & ^X1Y0Z0},
					{X: 0, Y: 1, Z: 2, N: allNeighbors & ^X1Y0Z0},
					{X: 0, Y: 2, Z: 2, N: allNeighbors & ^X1Y0Z0},
				},
			},
			bv: &binvox.BinVOX{
				NX: 3, NY: 3, NZ: 3,
				Scale: 1,
				Voxels: binvox.VoxelMap{ // a left-facing "wall" of 9 voxels with no neighbors to the left and 1 to the right.
					binvox.Key{X: 1, Y: 0, Z: 0}: binvox.White,
					binvox.Key{X: 1, Y: 1, Z: 0}: binvox.White,
					binvox.Key{X: 1, Y: 2, Z: 0}: binvox.White,
					binvox.Key{X: 1, Y: 0, Z: 1}: binvox.White,
					binvox.Key{X: 1, Y: 1, Z: 1}: binvox.White,
					binvox.Key{X: 1, Y: 2, Z: 1}: binvox.White,
					binvox.Key{X: 1, Y: 0, Z: 2}: binvox.White,
					binvox.Key{X: 1, Y: 1, Z: 2}: binvox.White,
					binvox.Key{X: 1, Y: 2, Z: 2}: binvox.White,
					binvox.Key{X: 2, Y: 0, Z: 0}: binvox.White,
					binvox.Key{X: 2, Y: 1, Z: 0}: binvox.White,
					binvox.Key{X: 2, Y: 2, Z: 0}: binvox.White,
					binvox.Key{X: 2, Y: 0, Z: 1}: binvox.White,
					binvox.Key{X: 2, Y: 1, Z: 1}: binvox.White,
					binvox.Key{X: 2, Y: 2, Z: 1}: binvox.White,
					binvox.Key{X: 2, Y: 0, Z: 2}: binvox.White,
					binvox.Key{X: 2, Y: 1, Z: 2}: binvox.White,
					binvox.Key{X: 2, Y: 2, Z: 2}: binvox.White,
				},
			},
			// want the old 0,1,1 and the new 1,1,1 voxels to be removed due to being fully enclosed.
			want: &VShell{
				NX: 3, NY: 3, NZ: 3,
				Scale: 1,
				Voxels: []VShVoxel{
					{X: 1, Y: 0, Z: 0, N: X_1Y0Z0 | X_1Y0Z1 | X_1Y1Z0 | X_1Y1Z1 | X0Y0Z1 | X0Y1Z0 | X0Y1Z1 | X1Y0Z0 | X1Y0Z1 | X1Y1Z0 | X1Y1Z1},
					{X: 1, Y: 0, Z: 1, N: allNeighbors & ^X_1Y_1Z_1 & ^X_1Y_1Z0 & ^X_1Y_1Z1 & ^X0Y_1Z_1 & ^X0Y_1Z0 & ^X0Y_1Z1 & ^X1Y_1Z_1 & ^X1Y_1Z0 & ^X1Y_1Z1},
					{X: 1, Y: 0, Z: 2, N: X_1Y0Z_1 | X_1Y0Z0 | X_1Y1Z_1 | X_1Y1Z0 | X0Y0Z_1 | X0Y1Z_1 | X0Y1Z0 | X1Y0Z_1 | X1Y0Z0 | X1Y1Z_1 | X1Y1Z0},
					{X: 1, Y: 1, Z: 0, N: allNeighbors & ^X_1Y_1Z_1 & ^X_1Y0Z_1 & ^X_1Y1Z_1 & ^X0Y_1Z_1 & ^X0Y0Z_1 & ^X0Y1Z_1 & ^X1Y_1Z_1 & ^X1Y0Z_1 & ^X1Y1Z_1},
					// removed {X:1,Y:1,Z:1},
					{X: 1, Y: 1, Z: 2, N: allNeighbors & ^X_1Y_1Z1 & ^X_1Y0Z1 & ^X_1Y1Z1 & ^X0Y_1Z1 & ^X0Y0Z1 & ^X0Y1Z1 & ^X1Y_1Z1 & ^X1Y0Z1 & ^X1Y1Z1},
					{X: 1, Y: 2, Z: 0, N: X_1Y_1Z0 | X_1Y_1Z1 | X_1Y0Z0 | X_1Y0Z1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z1 | X1Y_1Z0 | X1Y_1Z1 | X1Y0Z0 | X1Y0Z1},
					{X: 1, Y: 2, Z: 1, N: allNeighbors & ^X_1Y1Z_1 & ^X_1Y1Z0 & ^X_1Y1Z1 & ^X0Y1Z_1 & ^X0Y1Z0 & ^X0Y1Z1 & ^X1Y1Z_1 & ^X1Y1Z0 & ^X1Y1Z1},
					{X: 1, Y: 2, Z: 2, N: X_1Y_1Z_1 | X_1Y_1Z0 | X_1Y0Z_1 | X_1Y0Z0 | X0Y_1Z_1 | X0Y_1Z0 | X0Y0Z_1 | X1Y_1Z_1 | X1Y_1Z0 | X1Y0Z_1 | X1Y0Z0},
					{X: 2, Y: 0, Z: 0, N: X_1Y0Z0 | X_1Y0Z1 | X_1Y1Z0 | X_1Y1Z1 | X0Y0Z1 | X0Y1Z0 | X0Y1Z1},
					{X: 2, Y: 0, Z: 1, N: X_1Y0Z_1 | X_1Y0Z0 | X_1Y0Z1 | X_1Y1Z_1 | X_1Y1Z0 | X_1Y1Z1 | X0Y0Z_1 | X0Y0Z1 | X0Y1Z_1 | X0Y1Z0 | X0Y1Z1},
					{X: 2, Y: 0, Z: 2, N: X_1Y0Z_1 | X_1Y0Z0 | X_1Y1Z_1 | X_1Y1Z0 | X0Y0Z_1 | X0Y1Z_1 | X0Y1Z0},
					{X: 2, Y: 1, Z: 0, N: X_1Y_1Z0 | X_1Y_1Z1 | X_1Y0Z0 | X_1Y0Z1 | X_1Y1Z0 | X_1Y1Z1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z1 | X0Y1Z0 | X0Y1Z1},
					{X: 2, Y: 1, Z: 1, N: X_1Y_1Z_1 | X_1Y_1Z0 | X_1Y_1Z1 | X_1Y0Z_1 | X_1Y0Z0 | X_1Y0Z1 | X_1Y1Z_1 | X_1Y1Z0 | X_1Y1Z1 | X0Y_1Z_1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z_1 | X0Y0Z1 | X0Y1Z_1 | X0Y1Z0 | X0Y1Z1},
					{X: 2, Y: 1, Z: 2, N: X_1Y_1Z_1 | X_1Y_1Z0 | X_1Y0Z_1 | X_1Y0Z0 | X_1Y1Z_1 | X_1Y1Z0 | X0Y_1Z_1 | X0Y_1Z0 | X0Y0Z_1 | X0Y1Z_1 | X0Y1Z0},
					{X: 2, Y: 2, Z: 0, N: X_1Y_1Z0 | X_1Y_1Z1 | X_1Y0Z0 | X_1Y0Z1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z1},
					{X: 2, Y: 2, Z: 1, N: X_1Y_1Z_1 | X_1Y_1Z0 | X_1Y_1Z1 | X_1Y0Z_1 | X_1Y0Z0 | X_1Y0Z1 | X0Y_1Z_1 | X0Y_1Z0 | X0Y_1Z1 | X0Y0Z_1 | X0Y0Z1},
					{X: 2, Y: 2, Z: 2, N: X_1Y_1Z_1 | X_1Y_1Z0 | X_1Y0Z_1 | X_1Y0Z0 | X0Y_1Z_1 | X0Y_1Z0 | X0Y0Z_1},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.vs.Add(tc.bv); err != nil {
				if !tc.wantErr {
					t.Fatalf("Add = %v, want nil", err)
				}
				return
			}
			if tc.wantErr {
				t.Fatalf("Add = nil, want err")
			}

			sortVoxels(tc.vs.Voxels)
			sortVoxels(tc.want.Voxels)
			if !reflect.DeepEqual(tc.vs, tc.want) {
				t.Errorf("got %v\nwant %v", tc.vs, tc.want)
				if len(tc.vs.Voxels) != len(tc.want.Voxels) {
					for i, v := range tc.vs.Voxels {
						t.Errorf("got voxel[%v]=%v", i, v)
					}
					for i, v := range tc.want.Voxels {
						t.Errorf("want voxel[%v]=%v", i, v)
					}
				} else if !reflect.DeepEqual(tc.vs.Voxels, tc.want.Voxels) {
					for i, got := range tc.vs.Voxels {
						want := tc.want.Voxels[i]
						if got != want {
							t.Errorf("voxel %v:\n got {%v,%v,%v,%v},\nwant {%v,%v,%v,%v}", i, got.X, got.Y, got.Z, got.N, want.X, want.Y, want.Z, want.N)
						}
					}
				}
			}
		})
	}
}

func sortVoxels(v []VShVoxel) {
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

func TestDiffKeys(t *testing.T) {
	tests := []struct {
		name string
		o, n key
		want NeighborBitMap
	}{
		{name: "not neighbors"},
		{
			name: "left right",
			o:    key{0, 0, 0},
			n:    key{1, 0, 0},
			want: X1Y0Z0,
		},
		{
			name: "up down",
			o:    key{0, 0, 0},
			n:    key{0, 0, 1},
			want: X0Y0Z1,
		},
		{
			name: "front back",
			o:    key{0, 0, 0},
			n:    key{0, 1, 0},
			want: X0Y1Z0,
		},
		{
			name: "right left",
			o:    key{0, 0, 0},
			n:    key{-1, 0, 0},
			want: X_1Y0Z0,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := diffKeys(tc.o, tc.n); got != tc.want {
				t.Errorf("diffKeys(%v,%v) = %v, want %v", tc.o, tc.n, got, tc.want)
			}
		})
	}
}
