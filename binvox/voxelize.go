package binvox

import (
	"fmt"
	"log"
	"math"
	"sync"

	gl "github.com/fogleman/fauxgl"
)

const epsilon = 1e-9

// setMap respresents the voxel shell of mesh intersections.
type setMap map[Key][]*gl.Triangle

// addTri adds a triangle to the set if it has not already been added.
func (s setMap) addTri(k Key, tri *gl.Triangle) {
	for _, t := range s[k] {
		if t == tri {
			return
		}
	}
	s[k] = append(s[k], tri)
}

// Voxelize voxelizes a subregion of the mesh using (b.TX,b.TY,b.TZ) as the origin.
// The voxelized subregion will be: (0,0,0)-(b.NX-1,b.NY-1,b.NZ-1) (inclusive).
// b.Scale determines the scale of the voxelization. See Dim and VoxelsPerMM.
//
// Voxelize overwrites the Voxels slice in b.
func (b *BinVOX) Voxelize(mesh *gl.Mesh) error {
	if b.NX == 0 || b.NY == 0 || b.NZ == 0 {
		return fmt.Errorf("mesh dimensions must be non-zero (%v,%v,%v)", b.NX, b.NY, b.NZ)
	}

	b.Voxels = nil
	log.Printf("\n\nVoxelizing %v...", b)

	vpmm := b.VoxelsPerMM() // voxels per millimeter
	dz := 1.0 / vpmm        // millimeters per voxel

	var mu sync.Mutex // protects the voxels map
	voxels := make(VoxelMap)
	setVoxelFunc := func(k Key) {
		mu.Lock()
		voxels[k] = White
		mu.Unlock()
	}

	// Compute shell of voxel intersections.
	var wg sync.WaitGroup
	for zi := 0; zi < b.NZ; zi++ {
		wg.Add(1)
		go func(zi int) {
			b.voxelizeZ(mesh, zi, dz, vpmm, setVoxelFunc)
			// set := make(setMap)
			// z := b.TZ + (0.5+float64(zi))*dz
			// // log.Printf("horizontal slice @ zi=%v, z=%v", zi, z)
			// pairs := intersectZPlane(mesh, z)
			// // log.Printf("zi=%v: got %v intersection pairs", zi, len(pairs))
			// xMinMax := make(xMinMaxMap)
			// for _, pair := range pairs {
			// 	b.rasterizeShellPair(pair, zi, z, dz, vpmm, set, xMinMax)
			// }
			// // log.Printf("zi=%v: flood filling %v raster lines", zi, len(xMinMax))
			// b.floodFill(zi, set, setVoxelFunc, xMinMax)
			wg.Done()
		}(zi)
	}
	wg.Wait()

	b.Voxels = voxels

	log.Printf("Done creating %v voxels.", len(b.Voxels))
	return nil
}

func (b *BinVOX) VoxelizeZ(mesh *gl.Mesh, zi int) error {
	if b.NX == 0 || b.NY == 0 || b.NZ == 0 {
		return fmt.Errorf("mesh dimensions must be non-zero (%v,%v,%v)", b.NX, b.NY, b.NZ)
	}

	b.Voxels = nil
	log.Printf("\n\nVoxelizing z=%v of %v...", zi, b)

	vpmm := b.VoxelsPerMM() // voxels per millimeter
	dz := 1.0 / vpmm        // millimeters per voxel

	var mu sync.Mutex // protects the voxels map
	voxels := make(VoxelMap)
	setVoxelFunc := func(k Key) {
		mu.Lock()
		voxels[k] = White
		mu.Unlock()
	}

	// Compute shell of voxel intersections.
	b.voxelizeZ(mesh, zi, dz, vpmm, setVoxelFunc)

	b.Voxels = voxels

	log.Printf("Done creating %v voxels at z=%v.", len(b.Voxels), zi)
	return nil
}

func (b *BinVOX) voxelizeZ(mesh *gl.Mesh, zi int, dz, vpmm float64, setVoxelFunc func(k Key)) {
	// log.Printf("voxelizeZ(%v): %v triangles, dz=%v, vpmm=%v", zi, len(mesh.Triangles), dz, vpmm)
	set := make(setMap)
	z := b.TZ + (0.5+float64(zi))*dz
	// log.Printf("voxelizeZ(%v): horizontal slice @ z=%v", zi, z)
	pairs := intersectZPlane(mesh, z)
	// log.Printf("voxelizeZ(%v): got %v intersection pairs", zi, len(pairs))
	xMinMax := make(xMinMaxMap)
	for _, pair := range pairs {
		// log.Printf("voxelizeZ(%v): pair=%#v", zi, pair)
		b.rasterizeShellPair(pair, zi, z, dz, vpmm, set, xMinMax)
	}
	// log.Printf("voxelizeZ(%v): flood filling %v raster lines", zi, len(xMinMax))
	b.floodFill(zi, set, setVoxelFunc, xMinMax)
}

// intersectionPair represents an intersection between a triangle and the z-plane.
type intersectionPair struct {
	v1  gl.Vector // first intersection with z-plane
	v2  gl.Vector // second intersection point with z-plane
	tri *gl.Triangle
}

// intersectZPlane returns intersections between the mesh and the Z-plane at z.
func intersectZPlane(mesh *gl.Mesh, z float64) (result []*intersectionPair) {
	for _, t := range mesh.Triangles {
		if v, ok := intersectTriZPlane(t, z); ok {
			result = append(result, v)
		}
	}
	return result
}

// intersectTriZPlane returns an intersection between a triangle and the z-plane.
func intersectTriZPlane(t *gl.Triangle, z float64) (*intersectionPair, bool) {
	v1, ok1 := intersectSegment(t.V1.Position, t.V2.Position, z)
	v2, ok2 := intersectSegment(t.V2.Position, t.V3.Position, z)
	if ok1 && ok2 {
		return &intersectionPair{v1: v1, v2: v2, tri: t}, true
	}
	v3, ok3 := intersectSegment(t.V3.Position, t.V1.Position, z)
	if ok1 && ok3 {
		return &intersectionPair{v1: v1, v2: v3, tri: t}, true
	}
	if ok2 && ok3 {
		return &intersectionPair{v1: v2, v2: v3, tri: t}, true
	}
	return nil, false
}

// intersectSegment returns an intersection between two points and the z-plane.
func intersectSegment(v0, v1 gl.Vector, z float64) (gl.Vector, bool) {
	u := v1.Sub(v0)
	w := v0.Sub(gl.V(0, 0, z))
	d := gl.V(0, 0, 1).Dot(u)
	n := gl.V(0, 0, -1).Dot(w)
	if d > -epsilon && d < epsilon {
		return gl.Vector{}, false
	}
	t := n / d
	if t < 0 || t > 1 {
		return gl.Vector{}, false
	}
	v := v0.Add(u.MulScalar(t))
	return v, true
}

// firstOctantBressenham implements the Bressenham line drawing algorithm in the first octant.
func firstOctantBressenham(x1i, y1i, x2i, y2i int, plotFunc func(x, y int)) {
	dx := x2i - x1i
	dy := y2i - y1i
	y := y1i
	var eps int

	for x := x1i; x <= x2i; x++ {
		plotFunc(x, y)
		eps += dy
		if (eps << 1) >= dx {
			y++
			eps -= dx
		}
	}
}

// rasterizeShellPair rasterizes the shell voxels that intersect the mesh.
func (b *BinVOX) rasterizeShellPair(pair *intersectionPair, zi int, z, dz, vpmm float64, set setMap, xMinMax xMinMaxMap) {
	x1i := int(math.Floor(vpmm * (pair.v1.X - b.TX)))
	x2i := int(math.Floor(vpmm * (pair.v2.X - b.TX)))
	y1i := int(math.Floor(vpmm * (pair.v1.Y - b.TY)))
	y2i := int(math.Floor(vpmm * (pair.v2.Y - b.TY)))

	plotFunc := func(x, y int) {
		// log.Printf("plotFunc(%v,%v)", x, y)
		if y >= 0 && y < b.NY {
			k := Key{x, y, zi}
			xMinMax.update(x, y)
			set.addTri(k, pair.tri)
		}
	}

	// log.Printf("interpolating between (%v,%v) and (%v,%v)", x1i, y1i, x2i, y2i)
	dy := y2i - y1i
	dx := x2i - x1i
	switch {
	case dx >= 0 && dy >= 0 && dx >= dy: // 1st octant
		// log.Printf("dx=%v, dy=%v, 1st octant", dx, dy)
		firstOctantBressenham(x1i, y1i, x2i, y2i, plotFunc)
	case dx >= 0 && dy >= 0: // 2nd octant - swap x and y
		// log.Printf("dx=%v, dy=%v, 2nd octant", dx, dy)
		firstOctantBressenham(y1i, x1i, y2i, x2i, func(x, y int) { plotFunc(y, x) })
	case dx < 0 && dy >= 0 && dy > -dx: // 3rd octant - swap -x and y
		// log.Printf("dx=%v, dy=%v, 3rd octant", dx, dy)
		firstOctantBressenham(y1i, -x1i, y2i, -x2i, func(x, y int) { plotFunc(-y, x) })
	case dx < 0 && dy >= 0: // 4th octant - use -x
		// log.Printf("dx=%v, dy=%v, 4th octant", dx, dy)
		firstOctantBressenham(-x1i, y1i, -x2i, y2i, func(x, y int) { plotFunc(-x, y) })
	case dx < 0 && dy < 0 && -dx >= -dy: // 5th octant - use -x and -y
		// log.Printf("dx=%v, dy=%v, 5th octant", dx, dy)
		firstOctantBressenham(-x1i, -y1i, -x2i, -y2i, func(x, y int) { plotFunc(-x, -y) })
	case dx < 0 && dy < 0: // 6th octant - swap -x and -y
		// log.Printf("dx=%v, dy=%v, 6th octant", dx, dy)
		firstOctantBressenham(-y1i, -x1i, -y2i, -x2i, func(x, y int) { plotFunc(-y, -x) })
	case dx >= 0 && dy < 0 && -dy > dx: // 7th octant - swap x and -y
		// log.Printf("dx=%v, dy=%v, 7th octant", dx, dy)
		firstOctantBressenham(-y1i, x1i, -y2i, x2i, func(x, y int) { plotFunc(y, -x) })
	default: // 8th octant - use x and -y
		// log.Printf("dx=%v, dy=%v, 8th octant", dx, dy)
		firstOctantBressenham(x1i, -y1i, x2i, -y2i, func(x, y int) { plotFunc(x, -y) })
	}
}

// minMax holds the min and max x values.
type minMax struct {
	min, max int
}

// xMinMaxMap maps the y coordinate to min and max x values for that y scanline.
type xMinMaxMap map[int]*minMax

func (x xMinMaxMap) update(xi, yi int) {
	if v, ok := x[yi]; ok {
		if xi < v.min {
			v.min = xi
		}
		if xi > v.max {
			v.max = xi
		}
	} else {
		x[yi] = &minMax{xi, xi}
	}
}

// floodFill flood-fills the internal voxels within the voxel shell.
func (b *BinVOX) floodFill(zi int, set setMap, setVoxelFunc func(k Key), xMinMax xMinMaxMap) {
	var wg sync.WaitGroup
	for yi, mm := range xMinMax { // yi >= 0 && yi < b.NY
		wg.Add(1)
		go func(yi, min, max int) {
			var inside bool
			seenTris := make(map[*gl.Triangle]struct{}) // Only process each triangle once.
			for xi := min; xi <= max; xi++ {
				if xi >= b.NX { // Don't care past the subregion of interest.
					break
				}
				k := Key{xi, yi, zi}
				if len(set[k]) > 0 { // triangles intersect this voxel.
					setVoxelFunc(k)
					var inCount, outCount int
					for _, tri := range set[k] {
						if _, ok := seenTris[tri]; ok {
							continue
						}
						seenTris[tri] = struct{}{}
						switch d := tri.V1.Normal.Dot(gl.V(-1, 0, 0)); {
						case d > epsilon:
							inCount++
						case d < -epsilon:
							outCount++
						default:
						}
					}
					switch {
					case inCount > outCount:
						inside = true
					case inCount < outCount:
						inside = false
						// otherwise, keep previous value
					}
					continue
				}
				if inside && xi >= 0 {
					setVoxelFunc(k)
				}
			}
			wg.Done()
		}(yi, mm.min, mm.max)
	}
	wg.Wait()
}
