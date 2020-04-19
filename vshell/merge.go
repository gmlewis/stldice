package vshell

import (
	"fmt"
	"log"

	"github.com/gmlewis/stldice/v2/binvox"
)

// Merge merges one or more binvox files into a single VShell.
func Merge(binfiles []string) (*VShell, error) {
	var vs *VShell
	for _, filename := range binfiles {
		bv, err := binvox.Read(filename, 0, 0, 0, 0, 0, 0)
		if err != nil {
			return nil, err
		}
		log.Printf("BinVOX=%v", bv)

		if vs == nil {
			var err error
			if vs, err = New(bv); err != nil {
				return nil, fmt.Errorf("New: %v", err)
			}
		} else {
			if err := vs.Add(bv); err != nil {
				return nil, fmt.Errorf("Add: %v", err)
			}
		}
		log.Printf("VShell=%v", vs)
	}
	return vs, nil
}

type lookupMap map[key]struct{}
type shellMap map[key]*VShVoxel

// Add adds shell voxels from bv to vs.
func (vs *VShell) Add(bv *binvox.BinVOX) (err error) {
	if vs == nil || bv == nil {
		return fmt.Errorf("empty vs or bv")
	}
	var dx, dy, dz int
	if len(vs.Voxels) > 0 {
		if dx, dy, dz, err = vs.update(bv); err != nil {
			return err
		}
	}

	log.Printf("creating lookup table for %v new voxels...", len(bv.Voxels))
	lookup := make(lookupMap)
	beforeVPMM := vs.VoxelsPerMM()
	// log.Printf("GML: beforeVPMM=%v, before scale=%v", beforeVPMM, vs.Scale)
	for v := range bv.Voxels {
		nx := v.X + dx
		ny := v.Y + dy
		nz := v.Z + dz
		lookup[key{nx, ny, nz}] = struct{}{}
		if nx >= vs.NX {
			vs.NX = nx + 1
		}
		if ny >= vs.NY {
			vs.NY = ny + 1
		}
		if nz >= vs.NZ {
			vs.NZ = nz + 1
		}
	}
	afterVPMM := vs.VoxelsPerMM()
	// log.Printf("GML: afterVPMM=%v", afterVPMM)
	if beforeVPMM != afterVPMM { // Adjust the Scale to preserve the vpmm.
		vs.Scale *= afterVPMM / beforeVPMM
		// log.Printf("GML: after scale=%v", vs.Scale)
	}
	log.Printf("New shell dimensions are: (%v,%v,%v)", vs.NX, vs.NY, vs.NZ)

	log.Printf("creating lookup table for %v old shell voxels...", len(vs.Voxels))
	old := make(shellMap)
	for i, v := range vs.Voxels {
		old[key{v.X, v.Y, v.Z}] = &vs.Voxels[i]
	}

	log.Printf("processing %v voxels...", len(lookup))
	var shellCount int
	toRemove := make(lookupMap)
	for k := range lookup {
		neighbors, toCheck := lookup.findNeighbors(k, old)
		// log.Printf("GML: Update old neighbors of %v: %v (new neighbors=%v)", k, toCheck, neighbors)
		for _, ck := range toCheck {
			dk := diffKeys(ck, k)
			old[ck].N = old[ck].N | dk
			if old[ck].N == allNeighbors {
				// log.Printf("GML: Removing old shell key %v", ck)
				toRemove[ck] = struct{}{}
			}
		}
		if neighbors == allNeighbors { // ignore completely enclosed voxels
			// log.Printf("GML: Removing new voxel %v...", k)
			continue
		}

		shellCount++
		vs.Voxels = append(vs.Voxels, VShVoxel{X: k.X, Y: k.Y, Z: k.Z, N: neighbors})
	}
	log.Printf("added %v voxels to shell", shellCount)
	if len(toRemove) > 0 {
		log.Printf("removing %v voxels from old shell...", len(toRemove))
		newVoxels := make([]VShVoxel, 0, len(vs.Voxels)-len(toRemove))
		for _, vv := range vs.Voxels {
			if _, ok := toRemove[key{vv.X, vv.Y, vv.Z}]; !ok {
				newVoxels = append(newVoxels, vv)
			}
		}
		vs.Voxels = newVoxels
		log.Printf("done removing %v voxels from old shell; new total voxel count = %v", len(toRemove), len(vs.Voxels))
	}
	return nil
}

// update adjusts the locations of the current VShVoxels to accomodate
// the addition of the BinVOX model.
func (vs *VShell) update(bv *binvox.BinVOX) (dx, dy, dz int, err error) {
	vpmm := vs.VoxelsPerMM()
	if int(vpmm) != int(bv.VoxelsPerMM()) { // Course test for large discrepancies
		return 0, 0, 0, fmt.Errorf("vshell.VoxelsPerMM (%v) != binvox.VoxelsPerMM (%v)", vpmm, bv.VoxelsPerMM())
	}

	dx = int((bv.TX - vs.TX) * vpmm)
	dy = int((bv.TY - vs.TY) * vpmm)
	dz = int((bv.TZ - vs.TZ) * vpmm)
	if bv.TX < vs.TX {
		vs.TX = bv.TX
	} else {

	}
	if bv.TY < vs.TY {
		vs.TY = bv.TY
	}
	if bv.TZ < vs.TZ {
		vs.TZ = bv.TZ
	}

	if dx >= 0 && dy >= 0 && dz >= 0 {
		log.Printf("shifting %v new voxels by (%v,%v,%v) to merge with shell", len(bv.Voxels), dx, dy, dz)
		return dx, dy, dz, nil
	}

	log.Printf("moving %v voxels by (%v,%v,%v)", len(vs.Voxels), -dx, -dy, -dz)
	beforeVPMM := vs.VoxelsPerMM()
	// log.Printf("GML: beforeVPMM=%v, before scale=%v", beforeVPMM, vs.Scale)
	for i := range vs.Voxels {
		vs.Voxels[i].X -= dx
		vs.Voxels[i].Y -= dy
		vs.Voxels[i].Z -= dz
		if vs.Voxels[i].X >= vs.NX {
			vs.NX = vs.Voxels[i].X + 1
		}
		if vs.Voxels[i].Y >= vs.NY {
			vs.NY = vs.Voxels[i].Y + 1
		}
		if vs.Voxels[i].Z >= vs.NZ {
			vs.NZ = vs.Voxels[i].Z + 1
		}
	}
	afterVPMM := vs.VoxelsPerMM()
	// log.Printf("GML: afterVPMM=%v", afterVPMM)
	if beforeVPMM != afterVPMM { // Adjust the Scale to preserve the vpmm.
		vs.Scale *= afterVPMM / beforeVPMM
		// log.Printf("GML: after scale=%v", vs.Scale)
	}
	log.Printf("done moving %v voxels by (%v,%v,%v)", len(vs.Voxels), -dx, -dy, -dz)
	log.Printf("New shell dimensions are: (%v,%v,%v)", vs.NX, vs.NY, vs.NZ)

	return 0, 0, 0, nil
}

func (m lookupMap) findNeighbors(k key, old shellMap) (result NeighborBitMap, toCheck []key) {
	mask := X_1Y_1Z_1
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			for z := -1; z <= 1; z++ {
				if mask != x0y0z0 {
					tk := key{k.X + x, k.Y + y, k.Z + z}
					if _, ok := m[tk]; ok {
						result = result | mask
					} else if _, ok := old[tk]; ok {
						result = result | mask
						toCheck = append(toCheck, tk)
					}
				}
				mask = mask << 1
			}
		}
	}
	return result, toCheck
}

// diffKeys compares o to n and returns a NeighborBitMap representing the difference.
// e.g. o={0,0,0}, n={1,0,0}, result=X1Y0Z0
func diffKeys(o, n key) (result NeighborBitMap) {
	dx := n.X - o.X
	dy := n.Y - o.Y
	dz := n.Z - o.Z
	result = X_1Y_1Z_1
	switch dx {
	case -1: // no shift necessary
	case 0:
		result = result << 9
	case 1:
		result = result << 18
	default:
		log.Printf("diffKeys(%v,%v) - not neighbors at all", o, n)
		return 0
	}
	switch dy {
	case -1: // no shift necessary
	case 0:
		result = result << 3
	case 1:
		result = result << 6
	default:
		log.Printf("diffKeys(%v,%v) - not neighbors at all", o, n)
		return 0
	}
	switch dz {
	case -1: // no shift necessary
	case 0:
		result = result << 1
	case 1:
		result = result << 2
	default:
		log.Printf("diffKeys(%v,%v) - not neighbors at all", o, n)
		return 0
	}
	return result & allNeighbors // strip out x0y0z0
}
