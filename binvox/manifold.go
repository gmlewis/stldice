package binvox

import (
	"fmt"
	"log"
	"strings"

	gl "github.com/fogleman/fauxgl"
)

const (
	g0 neighborBitMap = 1 << iota // Classical Marching Cubes grid points
	g1
	g2
	g3
	g4
	g5
	g6
	g7
)

var (
	verbose = false
)

type neighborBitMap byte

type manifoldMap map[Key]neighborBitMap

func (b *BinVOX) ManifoldMesh() *gl.Mesh {
	gridCells := make(manifoldMap) // grid cell locations
	keyFunc := func(v Key) {
		gridCells[Key{v.X, v.Y, v.Z}] = gridCells[Key{v.X, v.Y, v.Z}] | g0
		gridCells[Key{v.X - 1, v.Y, v.Z}] = gridCells[Key{v.X - 1, v.Y, v.Z}] | g1
		gridCells[Key{v.X, v.Y + 1, v.Z}] = gridCells[Key{v.X, v.Y + 1, v.Z}] | g3
		gridCells[Key{v.X - 1, v.Y + 1, v.Z}] = gridCells[Key{v.X - 1, v.Y + 1, v.Z}] | g2
		gridCells[Key{v.X, v.Y, v.Z - 1}] = gridCells[Key{v.X, v.Y, v.Z - 1}] | g4
		gridCells[Key{v.X - 1, v.Y, v.Z - 1}] = gridCells[Key{v.X - 1, v.Y, v.Z - 1}] | g5
		gridCells[Key{v.X, v.Y + 1, v.Z - 1}] = gridCells[Key{v.X, v.Y + 1, v.Z - 1}] | g7
		gridCells[Key{v.X - 1, v.Y + 1, v.Z - 1}] = gridCells[Key{v.X - 1, v.Y + 1, v.Z - 1}] | g6
	}
	for v := range b.WhiteVoxels {
		keyFunc(v)
	}
	for v := range b.ColorVoxels {
		keyFunc(v)
	}

	var tris []*gl.Triangle
	vpmm := b.VoxelsPerMM()
	mmpv := 1.0 / vpmm
	s := gl.V(mmpv, mmpv, mmpv)
	t := gl.V(b.TX, b.TY, b.TZ)
	voxelToVector := func(k Key, dx, dy, dz float64) gl.Vector {
		x := float64(k.X) + 1
		y := float64(k.Y)
		z := float64(k.Z) + 1
		v := gl.V(x+dx, y+dy, z+dz).Mul(s).Add(t)
		// vlog("voxelToVector(%v, %v, %v, %v) = %v", k, dx, dy, dz, v)
		return v
	}

	for k, v := range gridCells {
		tris = append(tris, grid2tris(k, v, voxelToVector)...)
	}

	return gl.NewMesh(tris, nil)
}

func vlog(fmts string, args ...interface{}) {
	if verbose {
		log.Printf(fmts, args...)
	}
}

type voxelToVectorFunc func(k Key, dx, dy, dz float64) gl.Vector

func apply(k Key, v2v voxelToVectorFunc, in []*gl.Triangle) (out []*gl.Triangle) {
	for _, t := range in {
		p1 := t.V1.Position
		p2 := t.V2.Position
		p3 := t.V3.Position
		out = append(out, gl.NewTriangleForPoints(v2v(k, p1.X, p1.Y, p1.Z), v2v(k, p2.X, p2.Y, p2.Z), v2v(k, p3.X, p3.Y, p3.Z)))
	}
	return out
}

func rotateTrisClockwiseZ(top2v voxelToVectorFunc) voxelToVectorFunc {
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, dy, -dx, dz) }
}

func rotateTrisCounterClockwiseZ(top2v voxelToVectorFunc) voxelToVectorFunc {
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, -dy, dx, dz) }
}

func rotateTris180Z(top2v voxelToVectorFunc) voxelToVectorFunc {
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, -dx, -dy, dz) }
}

func rotateTrisClockwiseX(top2v voxelToVectorFunc) voxelToVectorFunc {
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, dx, dz, -dy) }
}

func rotateTrisCounterClockwiseX(top2v voxelToVectorFunc) voxelToVectorFunc {
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, dx, -dz, dy) }
}

func rotateTris180X(top2v voxelToVectorFunc) voxelToVectorFunc {
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, dx, -dy, -dz) }
}

func rotateTrisClockwiseY(top2v voxelToVectorFunc) voxelToVectorFunc {
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, dz, dy, -dx) }
}

func rotateTrisCounterClockwiseY(top2v voxelToVectorFunc) voxelToVectorFunc {
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, -dz, dy, dx) }
}

func rotateTris180Y(top2v voxelToVectorFunc) voxelToVectorFunc {
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, -dx, dy, -dz) }
}

func mirrorX(top2v voxelToVectorFunc) voxelToVectorFunc { // flips normals
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, -dx, dy, dz) }
}

func mirrorY(top2v voxelToVectorFunc) voxelToVectorFunc { // flips normals
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, dx, -dy, dz) }
}

func mirrorZ(top2v voxelToVectorFunc) voxelToVectorFunc { // flips normals
	return func(k Key, dx, dy, dz float64) gl.Vector { return top2v(k, dx, dy, -dz) }
}

// flipNormals must make copies so that the originals are not messed up.
func flipNormals(tris []*gl.Triangle) (out []*gl.Triangle) {
	for _, t := range tris {
		out = append(out, gl.NewTriangle(t.V1, t.V3, t.V2))
	}
	return out
}

// grid2tris converts grid cells to triangles.
func grid2tris(k Key, n neighborBitMap, v2v voxelToVectorFunc) (tris []*gl.Triangle) {
	vlog("grid2tris: k=%v, n= %v", k, n)
	switch n { // 256 cases
	case 0, 0xff: // no faces - all inside or all outside

	// single corners
	case g0:
		return apply(k, v2v, singleCorner)
	case g1:
		return apply(k, rotateTrisClockwiseZ(v2v), singleCorner)
	case g2:
		return apply(k, rotateTris180Z(v2v), singleCorner)
	case g3:
		return apply(k, rotateTrisCounterClockwiseZ(v2v), singleCorner)
	case g4:
		return apply(k, rotateTrisCounterClockwiseZ(rotateTris180X(v2v)), singleCorner)
	case g5:
		return apply(k, rotateTris180Z(rotateTris180X(v2v)), singleCorner)
	case g6:
		return apply(k, rotateTrisClockwiseZ(rotateTris180X(v2v)), singleCorner)
	case g7:
		return apply(k, rotateTris180X(v2v), singleCorner)
	case g0 | g1 | g2 | g3 | g4 | g6 | g7: // mirrorZ of g1
		return apply(k, rotateTrisClockwiseZ(mirrorZ(v2v)), singleCorner)
	case g0 | g1 | g3 | g4 | g5 | g6 | g7: // mirrorZ of g6, mirrorX of g3, mirrorY of g1
		return apply(k, rotateTrisClockwiseZ(mirrorY(v2v)), singleCorner)
	case g0 | g2 | g3 | g4 | g5 | g6 | g7: // mirrorZ of g5, mirrorX of g0
		return apply(k, mirrorX(v2v), singleCorner)
	case g0 | g1 | g2 | g4 | g5 | g6 | g7: // mirrorZ of g7
		return apply(k, rotateTris180X(mirrorZ(v2v)), singleCorner)
	case g1 | g2 | g3 | g4 | g5 | g6 | g7: // mirrorZ of g4, mirrorX of g1
		return apply(k, rotateTrisClockwiseZ(mirrorX(v2v)), singleCorner)
	case g0 | g1 | g2 | g3 | g5 | g6 | g7: // mirrorX of g5, mirrorZ of g0
		return apply(k, mirrorZ(v2v), singleCorner)
	case g0 | g1 | g2 | g3 | g4 | g5 | g7: // mirrorX of g7
		return apply(k, rotateTris180X(mirrorX(v2v)), singleCorner)
	case g0 | g1 | g2 | g3 | g4 | g5 | g6: // mirrorX of g6, mirrorY of g4, mirrorZ of g3
		return apply(k, rotateTrisCounterClockwiseZ(mirrorZ(v2v)), singleCorner)

	// single faces
	case g0 | g1 | g2 | g3: // bottom
		return apply(k, rotateTrisClockwiseX(v2v), singleFace)
	case g4 | g5 | g6 | g7: // top
		return apply(k, rotateTrisCounterClockwiseX(v2v), singleFace)
	case g0 | g1 | g4 | g5: // back
		return apply(k, v2v, singleFace)
	case g1 | g2 | g5 | g6: // right
		return apply(k, rotateTrisClockwiseZ(v2v), singleFace)
	case g2 | g3 | g6 | g7: // front
		return apply(k, rotateTris180Z(v2v), singleFace)
	case g0 | g3 | g4 | g7: // left
		return apply(k, rotateTrisCounterClockwiseZ(v2v), singleFace)

	// two adjacent corners
	case g0 | g1: // back lower horizontal
		return apply(k, v2v, twoAdjacentCorners)
	case g0 | g3: // left lower horizontal
		return apply(k, rotateTrisCounterClockwiseZ(v2v), twoAdjacentCorners)
	case g1 | g2: // right lower horizontal
		return apply(k, rotateTrisClockwiseZ(v2v), twoAdjacentCorners)
	case g2 | g3: // front lower horizontal
		return apply(k, rotateTris180Z(v2v), twoAdjacentCorners)
	case g0 | g4: // left back vertical
		return apply(k, rotateTrisClockwiseY(v2v), twoAdjacentCorners)
	case g1 | g5: // right back vertical
		return apply(k, rotateTrisClockwiseY(rotateTrisClockwiseZ(v2v)), twoAdjacentCorners)
	case g2 | g6: // front right vertical
		return apply(k, rotateTrisClockwiseY(rotateTris180Z(v2v)), twoAdjacentCorners)
	case g3 | g7: // front left vertical
		return apply(k, rotateTrisClockwiseY(rotateTrisCounterClockwiseZ(v2v)), twoAdjacentCorners)
	case g4 | g5: // back upper horizontal
		return apply(k, rotateTrisCounterClockwiseX(v2v), twoAdjacentCorners)
	case g4 | g7: // left upper horizontal
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisCounterClockwiseZ(v2v)), twoAdjacentCorners)
	case g5 | g6: // right upper horizontal
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseZ(v2v)), twoAdjacentCorners)
	case g6 | g7: // front upper horizontal
		return apply(k, rotateTrisCounterClockwiseX(rotateTris180Z(v2v)), twoAdjacentCorners)
	case g0 | g1 | g2 | g3 | g6 | g7: // vertical mirror of g0 | g1
		return apply(k, mirrorZ(v2v), twoAdjacentCorners)
	case g0 | g1 | g2 | g3 | g5 | g6:
		return apply(k, rotateTrisCounterClockwiseZ(mirrorZ(v2v)), twoAdjacentCorners)
	case g0 | g1 | g4 | g5 | g6 | g7:
		return apply(k, mirrorY(v2v), twoAdjacentCorners)
	case g0 | g3 | g4 | g5 | g6 | g7:
		return apply(k, rotateTrisCounterClockwiseZ(mirrorX(v2v)), twoAdjacentCorners)
	case g1 | g2 | g4 | g5 | g6 | g7:
		return apply(k, rotateTrisClockwiseZ(mirrorX(v2v)), twoAdjacentCorners)
	case g2 | g3 | g4 | g5 | g6 | g7: // mirrorZ of g4 | g5
		return apply(k, rotateTrisCounterClockwiseX(mirrorZ(v2v)), twoAdjacentCorners)
	case g1 | g2 | g3 | g5 | g6 | g7: // mirrorX of g1 | g5, mirrorY of g3 | g7
		return apply(k, rotateTrisClockwiseY(rotateTrisClockwiseZ(mirrorX(v2v))), twoAdjacentCorners)
	case g0 | g1 | g2 | g4 | g5 | g6: // mirrorX of g2 | g6, mirrorY of g0 | g4
		return apply(k, rotateTrisClockwiseY(mirrorY(v2v)), twoAdjacentCorners)
	case g0 | g2 | g3 | g4 | g6 | g7: // mirrorX of g0 | g4
		return apply(k, rotateTrisClockwiseY(mirrorX(v2v)), twoAdjacentCorners)
	case g0 | g1 | g3 | g4 | g5 | g7: // mirrorX of g3 | g7, mirrorY of g1 | g5
		return apply(k, rotateTrisClockwiseY(rotateTrisCounterClockwiseZ(mirrorX(v2v))), twoAdjacentCorners)
	case g0 | g1 | g2 | g3 | g4 | g5: // mirrorY of g4 | g5, mirrorZ of g2 | g3
		return apply(k, rotateTris180Z(mirrorZ(v2v)), twoAdjacentCorners)
	case g0 | g1 | g2 | g3 | g4 | g7: // mirrorZ of g1 | g2
		return apply(k, rotateTrisClockwiseZ(mirrorZ(v2v)), twoAdjacentCorners)

	// two opposite level corners
	case g0 | g2: // bottom
		return apply(k, v2v, twoOppositeLevelCorners)
	case g1 | g3: // bottom
		return apply(k, rotateTrisClockwiseZ(v2v), twoOppositeLevelCorners)
	case g0 | g7: // left
		return apply(k, rotateTrisClockwiseY(rotateTrisClockwiseX(v2v)), twoOppositeLevelCorners)
	case g3 | g4: // left
		return apply(k, rotateTrisClockwiseY(v2v), twoOppositeLevelCorners)
	case g0 | g5: // back
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseY(v2v)), twoOppositeLevelCorners)
	case g1 | g4: // back
		return apply(k, rotateTrisCounterClockwiseX(v2v), twoOppositeLevelCorners)
	case g1 | g6: // right
		return apply(k, rotateTrisCounterClockwiseY(v2v), twoOppositeLevelCorners)
	case g2 | g5: // right
		return apply(k, rotateTrisCounterClockwiseY(rotateTrisClockwiseX(v2v)), twoOppositeLevelCorners)
	case g3 | g6: // front
		return apply(k, rotateTrisClockwiseX(v2v), twoOppositeLevelCorners)
	case g2 | g7: // front
		return apply(k, rotateTrisClockwiseX(rotateTrisClockwiseY(v2v)), twoOppositeLevelCorners)
	case g4 | g6: // top
		return apply(k, rotateTris180X(rotateTrisClockwiseZ(v2v)), twoOppositeLevelCorners)
	case g5 | g7: // top
		return apply(k, rotateTris180X(v2v), twoOppositeLevelCorners)

	// two opposite diagonal corners
	case g0 | g6:
		return apply(k, v2v, twoOppositeDiagonalCorners)
	case g1 | g7:
		return apply(k, rotateTrisClockwiseZ(v2v), twoOppositeDiagonalCorners)
	case g2 | g4:
		return apply(k, rotateTris180Z(v2v), twoOppositeDiagonalCorners)
	case g3 | g5:
		return apply(k, rotateTrisCounterClockwiseZ(v2v), twoOppositeDiagonalCorners)

		// three adjacent corners
	case g0 | g1 | g2:
		return apply(k, v2v, threeAdjacentCorners)
	case g0 | g1 | g3:
		return apply(k, rotateTrisCounterClockwiseZ(v2v), threeAdjacentCorners)
	case g0 | g2 | g3:
		return apply(k, rotateTris180Z(v2v), threeAdjacentCorners)
	case g1 | g2 | g3:
		return apply(k, rotateTrisClockwiseZ(v2v), threeAdjacentCorners)
	case g4 | g5 | g6:
		return apply(k, rotateTris180X(rotateTrisCounterClockwiseZ(v2v)), threeAdjacentCorners)
	case g4 | g5 | g7:
		return apply(k, rotateTris180X(rotateTris180Z(v2v)), threeAdjacentCorners)
	case g4 | g6 | g7:
		return apply(k, rotateTris180X(rotateTrisClockwiseZ(v2v)), threeAdjacentCorners)
	case g5 | g6 | g7:
		return apply(k, rotateTris180X(v2v), threeAdjacentCorners)
	case g0 | g4 | g5:
		return apply(k, rotateTrisCounterClockwiseY(rotateTrisCounterClockwiseZ(v2v)), threeAdjacentCorners)
	case g1 | g2 | g6:
		return apply(k, rotateTrisClockwiseY(rotateTris180Z(v2v)), threeAdjacentCorners)
	case g2 | g3 | g6:
		return apply(k, rotateTrisClockwiseX(v2v), threeAdjacentCorners)
	case g0 | g4 | g7:
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisCounterClockwiseZ(v2v)), threeAdjacentCorners)
	case g2 | g3 | g7:
		return apply(k, rotateTrisClockwiseY(rotateTrisCounterClockwiseZ(v2v)), threeAdjacentCorners)
	case g0 | g3 | g7:
		return apply(k, rotateTrisClockwiseX(rotateTrisClockwiseZ(v2v)), threeAdjacentCorners)
	case g0 | g1 | g5:
		return apply(k, rotateTrisClockwiseY(rotateTrisClockwiseZ(v2v)), threeAdjacentCorners)
	case g1 | g5 | g6:
		return apply(k, rotateTrisCounterClockwiseY(v2v), threeAdjacentCorners)
	case g3 | g6 | g7:
		return apply(k, rotateTrisCounterClockwiseX(rotateTris180Z(v2v)), threeAdjacentCorners)
	case g2 | g6 | g7:
		return apply(k, rotateTrisCounterClockwiseY(rotateTrisClockwiseZ(v2v)), threeAdjacentCorners)
	case g1 | g4 | g5:
		return apply(k, rotateTrisCounterClockwiseX(v2v), threeAdjacentCorners)
	case g0 | g1 | g4:
		return apply(k, rotateTrisClockwiseX(rotateTris180Z(v2v)), threeAdjacentCorners)
	case g3 | g4 | g7:
		return apply(k, rotateTrisCounterClockwiseY(rotateTris180Z(v2v)), threeAdjacentCorners)
	case g2 | g5 | g6:
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseZ(v2v)), threeAdjacentCorners)
	case g0 | g3 | g4:
		return apply(k, rotateTrisClockwiseY(v2v), threeAdjacentCorners)
	case g1 | g2 | g5:
		return apply(k, rotateTrisClockwiseX(rotateTrisCounterClockwiseZ(v2v)), threeAdjacentCorners)
	case g0 | g1 | g2 | g3 | g6:
		return apply(k, rotateTris180X(rotateTris180Z(v2v)), flipNormals(threeAdjacentCorners))
	case g0 | g1 | g2 | g3 | g7:
		return apply(k, rotateTris180X(rotateTrisCounterClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))
	case g1 | g4 | g5 | g6 | g7:
		return apply(k, rotateTris180Z(v2v), flipNormals(threeAdjacentCorners))
	case g0 | g4 | g5 | g6 | g7:
		return apply(k, rotateTrisClockwiseZ(v2v), flipNormals(threeAdjacentCorners))
	case g0 | g3 | g4 | g5 | g7:
		return apply(k, rotateTrisClockwiseY(rotateTris180Z(v2v)), flipNormals(threeAdjacentCorners))
	case g2 | g3 | g5 | g6 | g7:
		return apply(k, rotateTrisClockwiseX(rotateTris180Z(v2v)), flipNormals(threeAdjacentCorners))
	case g2 | g3 | g4 | g6 | g7:
		return apply(k, rotateTrisClockwiseY(rotateTrisClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))
	case g0 | g3 | g4 | g6 | g7:
		return apply(k, rotateTrisClockwiseX(rotateTrisCounterClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))
	case g1 | g2 | g5 | g6 | g7:
		return apply(k, rotateTrisClockwiseY(v2v), flipNormals(threeAdjacentCorners))
	case g0 | g1 | g4 | g5 | g7:
		return apply(k, rotateTrisClockwiseX(v2v), flipNormals(threeAdjacentCorners))
	case g0 | g1 | g4 | g5 | g6:
		return apply(k, rotateTrisClockwiseY(rotateTrisCounterClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))
	case g1 | g2 | g4 | g5 | g6:
		return apply(k, rotateTrisClockwiseX(rotateTrisClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))
	case g2 | g4 | g5 | g6 | g7:
		return apply(k, rotateTrisCounterClockwiseZ(v2v), flipNormals(threeAdjacentCorners))
	case g0 | g2 | g3 | g6 | g7:
		return apply(k, rotateTrisCounterClockwiseX(v2v), flipNormals(threeAdjacentCorners))
	case g0 | g2 | g3 | g4 | g7:
		return apply(k, rotateTrisCounterClockwiseY(v2v), flipNormals(threeAdjacentCorners))
	case g0 | g1 | g2 | g5 | g6:
		return apply(k, rotateTrisCounterClockwiseY(rotateTris180Z(v2v)), flipNormals(threeAdjacentCorners))
	case g1 | g2 | g3 | g5 | g6:
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisCounterClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))
	case g0 | g1 | g2 | g4 | g5: // (inverse = g367 = case g3 | g6 | g7:)
		return apply(k, rotateTrisCounterClockwiseX(rotateTris180Z(v2v)), flipNormals(threeAdjacentCorners))
	case g3 | g4 | g5 | g6 | g7: // (inverse = g012 = case g0 | g1 | g2:)
		return apply(k, v2v, flipNormals(threeAdjacentCorners))
	case g0 | g1 | g2 | g3 | g5: // (inverse = g467 = case g4 | g6 | g7:)
		return apply(k, rotateTris180X(rotateTrisClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))
	case g0 | g1 | g3 | g4 | g7: // (inverse = g256 = case g2 | g5 | g6:)
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))
	case g0 | g1 | g2 | g3 | g4: // (inverse = g567 = case g5 | g6 | g7:)
		return apply(k, rotateTris180X(v2v), flipNormals(threeAdjacentCorners))
	case g1 | g2 | g3 | g6 | g7: // (inverse = g045 = case g0 | g4 | g5:)
		return apply(k, rotateTrisCounterClockwiseY(rotateTrisCounterClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))
	case g0 | g1 | g3 | g4 | g5: // (inverse = g267 = case g2 | g6 | g7:)
		return apply(k, rotateTrisCounterClockwiseY(rotateTrisClockwiseZ(v2v)), flipNormals(threeAdjacentCorners))

	// half corners
	case g0 | g1 | g2 | g5:
		tris = append(tris, apply(k, v2v, halfCorner)...)
		tris = append(tris, apply(k, rotateTrisClockwiseX(rotateTrisCounterClockwiseZ(v2v)), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseY(v2v)), halfCorner)...)
		return tris
	case g0 | g1 | g3 | g4:
		tris = append(tris, apply(k, rotateTrisCounterClockwiseZ(v2v), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisClockwiseX(rotateTris180Z(v2v)), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseY(rotateTrisCounterClockwiseZ(v2v))), halfCorner)...) // could be simplified to two rotations
		return tris
	case g0 | g2 | g3 | g7:
		tris = append(tris, apply(k, rotateTris180Z(v2v), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisClockwiseX(rotateTrisClockwiseZ(v2v)), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseY(rotateTris180Z(v2v))), halfCorner)...) // could be simplified to two rotations
		return tris
	case g0 | g4 | g5 | g7: // ok, now I'm just being lazy. It works.
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisCounterClockwiseZ(v2v)), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisClockwiseX(rotateTrisCounterClockwiseZ(rotateTrisCounterClockwiseX(rotateTrisCounterClockwiseZ(v2v)))), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseY(rotateTrisCounterClockwiseX(rotateTrisCounterClockwiseZ(v2v)))), halfCorner)...)
		return tris
	case g1 | g2 | g3 | g6:
		tris = append(tris, apply(k, rotateTrisClockwiseZ(v2v), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisClockwiseX(v2v), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseY(rotateTrisClockwiseZ(v2v))), halfCorner)...) // could be simplified to two rotations, but let's allow the computer some fun.
		return tris
	case g1 | g4 | g5 | g6:
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(v2v), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisClockwiseX(rotateTrisCounterClockwiseZ(rotateTrisCounterClockwiseX(v2v))), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseY(rotateTrisCounterClockwiseX(v2v))), halfCorner)...)
		return tris
	case g2 | g5 | g6 | g7:
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseZ(v2v)), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisClockwiseX(rotateTrisCounterClockwiseZ(rotateTrisCounterClockwiseX(rotateTrisClockwiseZ(v2v)))), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseY(rotateTrisCounterClockwiseX(rotateTrisClockwiseZ(v2v)))), halfCorner)...)
		return tris
	case g3 | g4 | g6 | g7:
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTris180Z(v2v)), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisClockwiseX(rotateTrisCounterClockwiseZ(rotateTrisCounterClockwiseX(rotateTris180Z(v2v)))), halfCorner)...)
		tris = append(tris, apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseY(rotateTrisCounterClockwiseX(rotateTris180Z(v2v)))), halfCorner)...)
		return tris

	// unusual cases

	case g0 | g3 | g5 | g6 | g7:
		return apply(k, v2v, g0357)
	case g0 | g1 | g2 | g6 | g7: // (inverse = g345 = case g3 | g4 | g5:)
		return apply(k, mirrorZ(rotateTrisCounterClockwiseZ(v2v)), flipNormals(g0357))
	case g3 | g4 | g5: // (inverse = g01267 = case g0 | g1 | g2 | g6 | g7:)
		return apply(k, mirrorZ(rotateTrisCounterClockwiseZ(v2v)), g0357)

	case g2 | g3 | g4:
		return apply(k, v2v, g234)
	case g0 | g1 | g5 | g6 | g7: // (inverse = g234 = case g2 | g3 | g4:)
		return apply(k, v2v, flipNormals(g234))
	case g0 | g6 | g7: // (inverse = g12345 = case g1 | g2 | g3 | g4 | g5:)
		return apply(k, mirrorZ(v2v), flipNormals(g234))
	case g1 | g2 | g3 | g4 | g5: // (inverse = g067 = case g0 | g6 | g7:)
		return apply(k, mirrorZ(v2v), g234)
	case g1 | g6 | g7: // (inverse = g02345 = case g0 | g2 | g3 | g4 | g5:)
		return apply(k, mirrorX(v2v), flipNormals(g234))
	case g0 | g2 | g3 | g4 | g5: // (inverse = g167 = case g1 | g6 | g7:)
		return apply(k, mirrorX(v2v), g234)
	case g0 | g1 | g6: // (inverse = g23457 = case g2 | g3 | g4 | g5 | g7:)
		return apply(k, rotateTris180Z(v2v), g234)
	case g2 | g3 | g4 | g5 | g7: // (inverse = g016 = case g0 | g1 | g6:)
		return apply(k, rotateTris180Z(v2v), flipNormals(g234))
	case g0 | g1 | g7: // (inverse = g23456 = case g2 | g3 | g4 | g5 | g6:)
		return apply(k, rotateTris180Z(mirrorX(v2v)), flipNormals(g234))
	case g2 | g3 | g4 | g5 | g6: // (inverse = g017 = case g0 | g1 | g7:)
		return apply(k, rotateTris180Z(mirrorX(v2v)), g234)
	case g2 | g4 | g5: // (inverse = g01367 = case g0 | g1 | g3 | g6 | g7:)
		return apply(k, mirrorZ(rotateTris180Z(v2v)), flipNormals(g234))
	case g0 | g1 | g3 | g6 | g7: // (inverse = g245 = case g2 | g4 | g5:)
		return apply(k, mirrorZ(rotateTris180Z(v2v)), g234)
	case g2 | g3 | g5: // (inverse = g01467 = case g0 | g1 | g4 | g6 | g7:)
		return apply(k, rotateTris180Z(mirrorX(rotateTris180Z(v2v))), flipNormals(g234))
	case g0 | g1 | g4 | g6 | g7: // (inverse = g235 = case g2 | g3 | g5:)
		return apply(k, rotateTris180Z(mirrorX(rotateTris180Z(v2v))), g234)

	case g0 | g1 | g6 | g7:
		return apply(k, v2v, g0167)
	case g2 | g3 | g4 | g5: // (inverse = g0167 = case g0 | g1 | g6 | g7:)
		return apply(k, v2v, flipNormals(g0167))

	case g0 | g1 | g5 | g6:
		return apply(k, v2v, g0156)
	case g0 | g3 | g6 | g7:
		return apply(k, mirrorX(rotateTrisCounterClockwiseZ(v2v)), flipNormals(g0156))
	case g2 | g3 | g4 | g7:
		return apply(k, rotateTris180Z(v2v), g0156)
	case g0 | g4 | g5 | g6:
		return apply(k, rotateTrisCounterClockwiseY(mirrorX(v2v)), flipNormals(g0156))
	case g3 | g4 | g5 | g7:
		return apply(k, rotateTrisCounterClockwiseY(mirrorX(rotateTrisCounterClockwiseZ(v2v))), flipNormals(g0156))
	case g0 | g1 | g2 | g6: // (inverse = g3457 = case g3 | g4 | g5 | g7:)
		return apply(k, rotateTrisCounterClockwiseY(mirrorX(rotateTrisCounterClockwiseZ(v2v))), g0156)
	case g2 | g3 | g5 | g6:
		return apply(k, mirrorX(rotateTris180Z(v2v)), flipNormals(g0156))
	case g0 | g3 | g4 | g5: // (inverse = g1267 = case g1 | g2 | g6 | g7:)
		return apply(k, rotateTrisCounterClockwiseZ(v2v), g0156)
	case g1 | g2 | g6 | g7: // (inverse = g0345 = case g0 | g3 | g4 | g5:)
		return apply(k, rotateTrisCounterClockwiseZ(v2v), flipNormals(g0156))
	case g0 | g1 | g4 | g7: // (inverse = g2356 = case g2 | g3 | g5 | g6:)
		return apply(k, mirrorX(rotateTris180Z(v2v)), g0156)
	case g1 | g2 | g4 | g5: // (inverse = g0367 = case g0 | g3 | g6 | g7:)
		return apply(k, mirrorX(rotateTrisCounterClockwiseZ(v2v)), g0156)
	case g1 | g4 | g5 | g7: // (inverse = g0236 = case g0 | g2 | g3 | g6:)
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisCounterClockwiseZ(v2v)), g0156)
	case g0 | g2 | g3 | g6:
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisCounterClockwiseZ(v2v)), flipNormals(g0156))
	case g0 | g4 | g6 | g7: // (inverse = g1235 = case g1 | g2 | g3 | g5:)
		return apply(k, rotateTris180Z(rotateTrisClockwiseX(v2v)), g0156)
	case g1 | g2 | g3 | g5:
		return apply(k, rotateTris180Z(rotateTrisClockwiseX(v2v)), flipNormals(g0156))
	case g2 | g4 | g5 | g6: // (inverse = g0137 = case g0 | g1 | g3 | g7:)
		return apply(k, rotateTrisClockwiseZ(rotateTrisClockwiseX(v2v)), g0156)
	case g0 | g1 | g3 | g7: // (inverse = g2456 = case g2 | g4 | g5 | g6:)
		return apply(k, rotateTrisClockwiseZ(rotateTrisClockwiseX(v2v)), flipNormals(g0156))
	case g0 | g2 | g3 | g4: // (inverse = g1567 = case g1 | g5 | g6 | g7:)
		return apply(k, rotateTrisCounterClockwiseY(mirrorX(rotateTrisCounterClockwiseZ(rotateTris180Z(v2v)))), g0156) // could be simplified.
	case g1 | g5 | g6 | g7: // (inverse = g0234 = case g0 | g2 | g3 | g4:)
		return apply(k, rotateTrisCounterClockwiseY(mirrorX(rotateTrisCounterClockwiseZ(rotateTris180Z(v2v)))), flipNormals(g0156))

	case g2 | g4 | g5 | g7:
		return apply(k, v2v, g2457)

	case g1 | g2 | g4 | g6 | g7:
		return apply(k, v2v, g12467)

	case g0 | g2 | g4 | g5 | g6 | g7:
		return apply(k, v2v, g024567)
	case g1 | g3 | g4 | g5 | g6 | g7:
		return apply(k, rotateTrisClockwiseZ(v2v), g024567)
	case g1 | g2 | g3 | g4 | g5 | g6:
		return apply(k, rotateTrisClockwiseY(v2v), g024567)
	case g1 | g2 | g3 | g4 | g6 | g7:
		return apply(k, rotateTrisCounterClockwiseX(v2v), g024567)
	case g0 | g2 | g3 | g4 | g5 | g7:
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisClockwiseZ(v2v)), g024567)
	case g0 | g2 | g3 | g5 | g6 | g7:
		return apply(k, rotateTrisClockwiseY(rotateTrisClockwiseZ(v2v)), g024567)
	case g0 | g1 | g3 | g4 | g6 | g7:
		return apply(k, rotateTrisCounterClockwiseY(v2v), g024567)
	case g0 | g1 | g2 | g5 | g6 | g7:
		return apply(k, rotateTrisCounterClockwiseX(rotateTrisCounterClockwiseZ(v2v)), g024567)

	case g1 | g3 | g4 | g5 | g6:
		return apply(k, v2v, g13456)
	case g0 | g2 | g4 | g5 | g7:
		return apply(k, rotateTrisCounterClockwiseZ(v2v), g13456)

	default:
		log.Printf("grid2tris: k=%v: unhandled:\n%v", k, n)
	}
	return nil
}

// TriangleLess provides a Less function for sort.Slice.
func TriangleLess(t []*gl.Triangle) func(a, b int) bool {
	return func(a, b int) bool {
		if t[a].V1.Position.Z < t[b].V1.Position.Z {
			return true
		}
		if t[a].V1.Position.Z > t[b].V1.Position.Z {
			return false
		}
		if t[a].V1.Position.Y < t[b].V1.Position.Y {
			return true
		}
		if t[a].V1.Position.Y > t[b].V1.Position.Y {
			return false
		}
		if t[a].V1.Position.X < t[b].V1.Position.X {
			return true
		}
		if t[a].V1.Position.X > t[b].V1.Position.X {
			return false
		}

		if t[a].V2.Position.Z < t[b].V2.Position.Z {
			return true
		}
		if t[a].V2.Position.Z > t[b].V2.Position.Z {
			return false
		}
		if t[a].V2.Position.Y < t[b].V2.Position.Y {
			return true
		}
		if t[a].V2.Position.Y > t[b].V2.Position.Y {
			return false
		}
		if t[a].V2.Position.X < t[b].V2.Position.X {
			return true
		}
		if t[a].V2.Position.X > t[b].V2.Position.X {
			return false
		}

		if t[a].V3.Position.Z < t[b].V3.Position.Z {
			return true
		}
		if t[a].V3.Position.Z > t[b].V3.Position.Z {
			return false
		}
		if t[a].V3.Position.Y < t[b].V3.Position.Y {
			return true
		}
		if t[a].V3.Position.Y > t[b].V3.Position.Y {
			return false
		}
		return t[a].V3.Position.X < t[b].V3.Position.X
	}
}

func (n neighborBitMap) String() string {
	test := "g"
	invTest := "g"
	var result []string
	var invResult []string
	if n&g0 != 0 {
		result = append(result, "g0")
		test += "0"
	} else {
		invResult = append(invResult, "g0")
		invTest += "0"
	}
	if n&g1 != 0 {
		result = append(result, "g1")
		test += "1"
	} else {
		invResult = append(invResult, "g1")
		invTest += "1"
	}
	if n&g2 != 0 {
		result = append(result, "g2")
		test += "2"
	} else {
		invResult = append(invResult, "g2")
		invTest += "2"
	}
	if n&g3 != 0 {
		result = append(result, "g3")
		test += "3"
	} else {
		invResult = append(invResult, "g3")
		invTest += "3"
	}
	if n&g4 != 0 {
		result = append(result, "g4")
		test += "4"
	} else {
		invResult = append(invResult, "g4")
		invTest += "4"
	}
	if n&g5 != 0 {
		result = append(result, "g5")
		test += "5"
	} else {
		invResult = append(invResult, "g5")
		invTest += "5"
	}
	if n&g6 != 0 {
		result = append(result, "g6")
		test += "6"
	} else {
		invResult = append(invResult, "g6")
		invTest += "6"
	}
	if n&g7 != 0 {
		result = append(result, "g7")
		test += "7"
	} else {
		invResult = append(invResult, "g7")
		invTest += "7"
	}
	if len(result) == 0 {
		return fmt.Sprintf("%v=%#x=*empty*", int(n), int(n))
	}
	return fmt.Sprintf("%v = case %v: (inverse = %v = case %v:)", test, strings.Join(result, " | "), invTest, strings.Join(invResult, " | "))
}
