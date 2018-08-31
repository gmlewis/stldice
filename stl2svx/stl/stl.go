// Package stl gets processes pb.STLFiles.
package stl

import (
	"fmt"
	"math"

	gl "github.com/fogleman/fauxgl"
	pb "github.com/gmlewis/stldice/v2/stl2svx/proto"
)

// STL represents a converted STL file to a mesh.
type STL struct {
	// MBB is the minimum bounding box for the entire STL file.
	MBB gl.Box
	// Mesh is the mesh of triangles.
	Mesh *gl.Mesh
	// The dimensions for the model.
	ModelDimX, ModelDimY, ModelDimZ int
	// The dimensions for each subregion.
	DimX, DimY, DimZ int
	// MMPV is the millimeters per voxel of the model.
	MMPV float64
	// SubregionScale represents the scale for each subregion.
	SubregionScale float64
}

// New parses a pb.STLFile and returns an STL.
// dim represents the number of voxels in the widest dimension.
// nX, nY, nZ represent the number of subdivisions in each dimension.
func New(p *pb.STLFile, dim, nX, nY, nZ int64) (*STL, error) {
	var tris []*gl.Triangle
	for _, t := range p.GetTriangles() {
		tris = append(tris, gl.NewTriangleForPoints(
			gl.V(t.V1.X, t.V1.Y, t.V1.Z),
			gl.V(t.V2.X, t.V2.Y, t.V2.Z),
			gl.V(t.V3.X, t.V3.Y, t.V3.Z)))
	}
	mesh := gl.NewTriangleMesh(tris)
	mbb := mesh.BoundingBox()

	scale := mbb.Max.X - mbb.Min.X
	if dy := mbb.Max.Y - mbb.Min.Y; dy > scale {
		scale = dy
	}
	if dz := mbb.Max.Z - mbb.Min.Z; dz > scale {
		scale = dz
	}
	vpmm := float64(dim) / scale // voxels per millimeter
	mmpv := 1.0 / vpmm           // millimeters per voxel
	modelDimInMM := mbb.Size()
	newModelDimX := int(math.Ceil(modelDimInMM.X * vpmm))
	newModelDimY := int(math.Ceil(modelDimInMM.Y * vpmm))
	newModelDimZ := int(math.Ceil(modelDimInMM.Z * vpmm))
	dimX := newModelDimX / int(nX)
	dimY := newModelDimY / int(nY)
	dimZ := newModelDimZ / int(nZ)
	if dimX == 0 || dimY == 0 || dimZ == 0 {
		return nil, fmt.Errorf("too many divisions: region dimensions = (%v,%v,%v)", dimX, dimY, dimZ)
	}
	maxDim := dimX
	if dimY > maxDim {
		maxDim = dimY
	}
	if dimZ > maxDim {
		maxDim = dimZ
	}
	subregionScale := float64(maxDim) * mmpv

	return &STL{
		MBB:       mbb,
		Mesh:      mesh,
		ModelDimX: newModelDimX, ModelDimY: newModelDimY, ModelDimZ: newModelDimZ,
		DimX: dimX, DimY: dimY, DimZ: dimZ,
		MMPV:           mmpv,
		SubregionScale: subregionScale,
	}, nil
}
