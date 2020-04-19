package binvox

import (
	"log"

	gl "github.com/fogleman/fauxgl"
)

// ToMesh converts a BinVOX to a mesh.
func (b *BinVOX) ToMesh() *gl.Mesh {
	log.Printf("Generating mesh for %v voxels...", len(b.Voxels))
	voxels := []gl.Voxel{}
	for v := range b.Voxels {
		voxels = append(voxels, gl.Voxel{X: v.X, Y: v.Y, Z: v.Z})
	}

	mesh := gl.NewVoxelMesh(voxels)
	log.Printf("Done generating mesh with %v triangles and %v lines.", len(mesh.Triangles), len(mesh.Lines))
	mbb := mesh.BoundingBox()
	log.Printf("Mesh MBB in voxel units = %v", mbb)

	log.Println("Moving the mesh back into its original position...")
	t1 := gl.V(0.5, 0.5, 0.5)
	log.Printf("Translating by %v", t1)
	mbb.Min = mbb.Min.Add(t1)
	mbb.Max = mbb.Max.Add(t1)
	log.Printf("Translated mesh MBB in voxel units = %v", mbb)

	vpmm := b.VoxelsPerMM()
	mmpv := 1.0 / vpmm
	s := gl.V(mmpv, mmpv, mmpv) // uniform scaling
	log.Printf("Scaling by %v millimeters per voxel", mmpv)

	t2 := gl.V(b.TX, b.TY, b.TZ)
	log.Printf("Translating by %v", t2)

	m := gl.Identity().Translate(t1).Scale(s).Translate(t2)
	mesh.Transform(m)
	mbb = mesh.BoundingBox()
	log.Printf("Done moving mesh to MBB=%v", mbb)

	return mesh
}
