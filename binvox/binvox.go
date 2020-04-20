// Package binvox provides functions for reading and writing binvox files.
// See: http://www.patrickmin.com/binvox/binvox.html for more information.
package binvox

import (
	"fmt"
	"log"

	gl "github.com/fogleman/fauxgl"
)

// BinVOX represents a voxel model (or subregion) in binvox format.
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
type BinVOX struct {
	NX, NY, NZ int     // number of voxels in each dimension
	TX, TY, TZ float64 // translation (location of origin in world space)
	Scale      float64 // uniform scale in millimeters

	// Use either WhiteVoxels or ColorVoxels, but not both.

	// WhiteVoxels represents a map of (white) voxels.
	WhiteVoxels WhiteVoxelMap

	// ColorVoxels represents a map of full-color voxels.
	ColorVoxels ColorVoxelMap
}

// Key represents the location of a voxel.
type Key struct { // For the template dump
	X, Y, Z int
}

// Color represents a voxel color.
type Color struct {
	R, G, B, A float64
}

var (
	Black = Color{0, 0, 0, 1}
	White = Color{1, 1, 1, 1}
)

// WhiteVoxelMap respresents the voxel model as a map of (white) voxels.
type WhiteVoxelMap map[Key]struct{}

// ColorVoxelMap respresents the voxel model as a map of full-color voxels.
type ColorVoxelMap map[Key]Color

// New returns a new BinVOX struct.
func New(nx, ny, nz int, offx, offy, offz, scale float64, fullColor bool) *BinVOX {
	var whiteVoxels WhiteVoxelMap
	var colorVoxels ColorVoxelMap

	if fullColor {
		colorVoxels = ColorVoxelMap{}
	} else {
		whiteVoxels = WhiteVoxelMap{}
	}

	return &BinVOX{
		NX:          nx,
		NY:          ny,
		NZ:          nz,
		TX:          offx,
		TY:          offy,
		TZ:          offz,
		Scale:       scale,
		WhiteVoxels: whiteVoxels,
		ColorVoxels: colorVoxels,
	}
}

// String returns a summary string of the BinVOX.
func (b *BinVOX) String() string {
	mbb := b.MBB()
	return fmt.Sprintf("BinVOX(n=[%v,%v,%v], t=[%v,%v,%v], mbb=(%v,%v,%v)-(%v,%v,%v), scale=%v, %v vpmm, %v white voxels, %v color voxels)",
		b.NX, b.NY, b.NZ,
		b.TX, b.TY, b.TZ,
		mbb.Min.X, mbb.Min.Y, mbb.Min.Z,
		mbb.Max.X, mbb.Max.Y, mbb.Max.Z,
		b.Scale, b.VoxelsPerMM(),
		len(b.WhiteVoxels),
		len(b.ColorVoxels),
	)
}

// Get gets a voxel from either the WhiteVoxelMap (first), or the ColorVoxel Map.
func (b *BinVOX) Get(x, y, z int) (color Color, ok bool) {
	key := Key{X: x, Y: y, Z: z}
	if _, ok := b.WhiteVoxels[key]; ok {
		return White, true
	}
	if c, ok := b.ColorVoxels[key]; ok {
		return c, true
	}
	return color, false
}

// Add adds a (white) voxel to the BinVOX WhiteVoxels map.
func (b *BinVOX) Add(x, y, z int) {
	key := Key{X: x, Y: y, Z: z}
	b.WhiteVoxels[key] = struct{}{}
}

// AddColor adds a full-color voxel to the BinVOX ColorVoxels map.
func (b *BinVOX) AddColor(x, y, z int, c Color) {
	key := Key{X: x, Y: y, Z: z}
	b.ColorVoxels[key] = c
}

// Dim returns the maximum dimension (the max of NX, NY, and NZ).
func (b *BinVOX) Dim() int {
	dim := b.NX
	if b.NY > dim {
		dim = b.NY
	}
	if b.NZ > dim {
		dim = b.NZ
	}
	return dim
}

// VoxelsPerMM returns the number of voxels per millimeter.
func (b *BinVOX) VoxelsPerMM() float64 {
	if b.Scale <= 0 {
		log.Printf("VoxelsPerMM: bad scale in BinVOX %v", *b)
		return 1
	}
	return float64(b.Dim()) / b.Scale
}

// MBB returns the minimum bounding box of the subregion in millimeters.
func (b *BinVOX) MBB() *gl.Box {
	s := 1.0 / b.VoxelsPerMM()
	min := gl.V(b.TX, b.TY, b.TZ)
	max := gl.V(b.TX+s*float64(b.NX), b.TY+s*float64(b.NY), b.TZ+s*float64(b.NZ))
	return &gl.Box{Min: min, Max: max}
}
