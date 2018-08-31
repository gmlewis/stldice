// Package vshell provides functions for reading and writing .vsh files.
package vshell

import (
	"fmt"
	"log"
	"strings"

	gl "github.com/fogleman/fauxgl"
	"github.com/gmlewis/stldice/v2/binvox"
)

// VShell represents a voxel model (or subregion) in vshell format.
// Voxels have uniform dimensions in X, Y, and Z and can be thought
// of as 3D "pixels".
//
// NX, NY, and NZ are the number of voxels in each dimension and must
// be positive (non-zero).
//
// Scale represents the uniform scale of this subregion's largest
// dimension (in millimeters). So if "dim = max(NX,NY,NZ)", then there
// are dim/Scale voxels per millimeter in each direction.
//
// Translating the subregion to (TX,TY,TZ) will correctly place this
// region in world space (in millimeters).
//
// VShells differ from BinVOX as follows:
//   * BinVOX lists the presence of each voxel as a sparse matrix
//   * VShell lists only the voxels exposed on the outer shell of
//     the model. That is, all "internal" voxels are elided. Along
//     with each outer "shell" voxel, its NeighborBitMap is also retained.
type VShell struct {
	NX, NY, NZ int     // number of voxels in each dimension
	TX, TY, TZ float64 // translation (location of origin in world space)
	Scale      float64 // uniform scale in millimeters
	Voxels     []VShVoxel
}

// VShVoxel represents a VShell Voxel.
type VShVoxel struct {
	X, Y, Z int
	N       NeighborBitMap
}

// NeighborBitMap represents the neighbors of a voxel and can be used to
// calculate the normal of the voxel.
type NeighborBitMap uint32

// key is used in voxel lookup maps to identify a voxel.
type key struct {
	X, Y, Z int
}

// New returns a new VShell from a BinVOX.
func New(bv *binvox.BinVOX) (*VShell, error) {
	vs := &VShell{
		NX: bv.NX, NY: bv.NY, NZ: bv.NZ,
		TX: bv.TX, TY: bv.TY, TZ: bv.TZ,
		Scale: bv.Scale,
	}
	if len(bv.Voxels) > 0 {
		if err := vs.Add(bv); err != nil {
			return nil, err
		}
	}
	return vs, nil
}

// String returns a summary string of the VShell.
func (v *VShell) String() string {
	mbb := v.MBB()
	return fmt.Sprintf("VShell(n=[%v,%v,%v], t=[%v,%v,%v], mbb=(%v,%v,%v)-(%v,%v,%v), scale=%v, %v vpmm, %v voxels)", v.NX, v.NY, v.NZ, v.TX, v.TY, v.TZ, mbb.Min.X, mbb.Min.Y, mbb.Min.Z, mbb.Max.X, mbb.Max.Y, mbb.Max.Z, v.Scale, v.VoxelsPerMM(), len(v.Voxels))
}

// String returns a summary string of a VshVoxel.
func (v VShVoxel) String() string {
	return fmt.Sprintf("{X:%v,Y:%v,Z:%v,N:%v}", v.X, v.Y, v.Z, v.N)
}

// Dim returns the maximum dimension (the max of NX, NY, and NZ).
func (v *VShell) Dim() int {
	dim := v.NX
	if v.NY > dim {
		dim = v.NY
	}
	if v.NZ > dim {
		dim = v.NZ
	}
	return dim
}

// VoxelsPerMM returns the number of voxels per millimeter.
func (v *VShell) VoxelsPerMM() float64 {
	if v.Scale <= 0 {
		log.Printf("VoxelsPerMM: bad scale in VShell(n=[%v,%v,%v], t=[%v,%v,%v], scale=%v, %v voxels)", v.NX, v.NY, v.NZ, v.TX, v.TY, v.TZ, v.Scale, len(v.Voxels))
		return 1
	}
	return float64(v.Dim()) / v.Scale
}

// MBB returns the minimum bounding box of the subregion in millimeters.
func (v *VShell) MBB() *gl.Box {
	s := 1.0 / v.VoxelsPerMM()
	min := gl.V(v.TX, v.TY, v.TZ)
	max := gl.V(v.TX+s*float64(v.NX), v.TY+s*float64(v.NY), v.TZ+s*float64(v.NZ))
	return &gl.Box{Min: min, Max: max}
}

const ( // NeighborBitMap values
	X_1Y_1Z_1 NeighborBitMap = 1 << iota
	X_1Y_1Z0
	X_1Y_1Z1
	X_1Y0Z_1
	X_1Y0Z0
	X_1Y0Z1
	X_1Y1Z_1
	X_1Y1Z0
	X_1Y1Z1
	X0Y_1Z_1
	X0Y_1Z0
	X0Y_1Z1
	X0Y0Z_1
	x0y0z0 // Unused - the center voxel - but needed as a place-holder for the bit pattern.
	X0Y0Z1
	X0Y1Z_1
	X0Y1Z0
	X0Y1Z1
	X1Y_1Z_1
	X1Y_1Z0
	X1Y_1Z1
	X1Y0Z_1
	X1Y0Z0
	X1Y0Z1
	X1Y1Z_1
	X1Y1Z0
	X1Y1Z1
)

// String returns an easy-to-read (bit possibly long) representation of NeighborBitMap.
func (n NeighborBitMap) String() string {
	var result []string
	var neg []string
	if n&X_1Y_1Z_1 != 0 {
		result = append(result, "X_1Y_1Z_1")
	} else {
		neg = append(neg, "^X_1Y_1Z_1")
	}
	if n&X_1Y_1Z0 != 0 {
		result = append(result, "X_1Y_1Z0")
	} else {
		neg = append(neg, "^X_1Y_1Z0")
	}
	if n&X_1Y_1Z1 != 0 {
		result = append(result, "X_1Y_1Z1")
	} else {
		neg = append(neg, "^X_1Y_1Z1")
	}
	if n&X_1Y0Z_1 != 0 {
		result = append(result, "X_1Y0Z_1")
	} else {
		neg = append(neg, "^X_1Y0Z_1")
	}
	if n&X_1Y0Z0 != 0 {
		result = append(result, "X_1Y0Z0")
	} else {
		neg = append(neg, "^X_1Y0Z0")
	}
	if n&X_1Y0Z1 != 0 {
		result = append(result, "X_1Y0Z1")
	} else {
		neg = append(neg, "^X_1Y0Z1")
	}
	if n&X_1Y1Z_1 != 0 {
		result = append(result, "X_1Y1Z_1")
	} else {
		neg = append(neg, "^X_1Y1Z_1")
	}
	if n&X_1Y1Z0 != 0 {
		result = append(result, "X_1Y1Z0")
	} else {
		neg = append(neg, "^X_1Y1Z0")
	}
	if n&X_1Y1Z1 != 0 {
		result = append(result, "X_1Y1Z1")
	} else {
		neg = append(neg, "^X_1Y1Z1")
	}
	if n&X0Y_1Z_1 != 0 {
		result = append(result, "X0Y_1Z_1")
	} else {
		neg = append(neg, "^X0Y_1Z_1")
	}
	if n&X0Y_1Z0 != 0 {
		result = append(result, "X0Y_1Z0")
	} else {
		neg = append(neg, "^X0Y_1Z0")
	}
	if n&X0Y_1Z1 != 0 {
		result = append(result, "X0Y_1Z1")
	} else {
		neg = append(neg, "^X0Y_1Z1")
	}
	if n&X0Y0Z_1 != 0 {
		result = append(result, "X0Y0Z_1")
	} else {
		neg = append(neg, "^X0Y0Z_1")
	}
	if n&x0y0z0 != 0 {
		result = append(result, "x0y0z0")
	} // "allNeighbors &" handles the ^x0y0z0 case.
	if n&X0Y0Z1 != 0 {
		result = append(result, "X0Y0Z1")
	} else {
		neg = append(neg, "^X0Y0Z1")
	}
	if n&X0Y1Z_1 != 0 {
		result = append(result, "X0Y1Z_1")
	} else {
		neg = append(neg, "^X0Y1Z_1")
	}
	if n&X0Y1Z0 != 0 {
		result = append(result, "X0Y1Z0")
	} else {
		neg = append(neg, "^X0Y1Z0")
	}
	if n&X0Y1Z1 != 0 {
		result = append(result, "X0Y1Z1")
	} else {
		neg = append(neg, "^X0Y1Z1")
	}
	if n&X1Y_1Z_1 != 0 {
		result = append(result, "X1Y_1Z_1")
	} else {
		neg = append(neg, "^X1Y_1Z_1")
	}
	if n&X1Y_1Z0 != 0 {
		result = append(result, "X1Y_1Z0")
	} else {
		neg = append(neg, "^X1Y_1Z0")
	}
	if n&X1Y_1Z1 != 0 {
		result = append(result, "X1Y_1Z1")
	} else {
		neg = append(neg, "^X1Y_1Z1")
	}
	if n&X1Y0Z_1 != 0 {
		result = append(result, "X1Y0Z_1")
	} else {
		neg = append(neg, "^X1Y0Z_1")
	}
	if n&X1Y0Z0 != 0 {
		result = append(result, "X1Y0Z0")
	} else {
		neg = append(neg, "^X1Y0Z0")
	}
	if n&X1Y0Z1 != 0 {
		result = append(result, "X1Y0Z1")
	} else {
		neg = append(neg, "^X1Y0Z1")
	}
	if n&X1Y1Z_1 != 0 {
		result = append(result, "X1Y1Z_1")
	} else {
		neg = append(neg, "^X1Y1Z_1")
	}
	if n&X1Y1Z0 != 0 {
		result = append(result, "X1Y1Z0")
	} else {
		neg = append(neg, "^X1Y1Z0")
	}
	if n&X1Y1Z1 != 0 {
		result = append(result, "X1Y1Z1")
	} else {
		neg = append(neg, "^X1Y1Z1")
	}
	if len(result) <= len(neg) {
		return fmt.Sprintf("0x%X = %v", uint32(n), strings.Join(result, " | "))
	}
	return fmt.Sprintf("0x%X = allNeighbors & %v", uint32(n), strings.Join(neg, " & "))
}

const allNeighbors = X_1Y_1Z_1 |
	X_1Y_1Z0 |
	X_1Y_1Z1 |
	X_1Y0Z_1 |
	X_1Y0Z0 |
	X_1Y0Z1 |
	X_1Y1Z_1 |
	X_1Y1Z0 |
	X_1Y1Z1 |
	X0Y_1Z_1 |
	X0Y_1Z0 |
	X0Y_1Z1 |
	X0Y0Z_1 |
	X0Y0Z1 |
	X0Y1Z_1 |
	X0Y1Z0 |
	X0Y1Z1 |
	X1Y_1Z_1 |
	X1Y_1Z0 |
	X1Y_1Z1 |
	X1Y0Z_1 |
	X1Y0Z0 |
	X1Y0Z1 |
	X1Y1Z_1 |
	X1Y1Z0 |
	X1Y1Z1
