package vshell

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

// Write writes a vshell file.
// sx, sy, sz are the starting indices of the model.
// nx, ny, nz are the number of voxels to write in each direction (0=all).
func (b *VShell) Write(filename string, sx, sy, sz, nx, ny, nz int) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create file %q: %v", filename, err)
	}
	n, err := b.write(f, sx, sy, sz, nx, ny, nz)
	if err != nil {
		f.Close()
		return fmt.Errorf("Write(%q): %v", filename, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("unable to close %q: %v", filename, err)
	}
	log.Printf("Done writing %v white voxels to file %q.", n, filename)
	return nil
}

const headerFMT = `#vshell 1
dim %v %v %v
translate %g %g %g
scale %g
data
`

func (vs *VShell) write(w io.Writer, sx, sy, sz, nx, ny, nz int) (int64, error) {
	if len(vs.Voxels) == 0 {
		fmt.Fprintf(w, headerFMT, 0, 0, 0, vs.TX, vs.TY, vs.TZ, vs.Scale)
		return 0, nil
	}

	if nx == 0 {
		nx = vs.NX
	}
	if ny == 0 {
		ny = vs.NY
	}
	if nz == 0 {
		nz = vs.NZ
	}

	// create lookup table
	lookup := map[key]int{}
	for i, v := range vs.Voxels {
		lookup[key{v.X, v.Y, v.Z}] = i
	}

	header := fmt.Sprintf(headerFMT, nx, ny, nz, vs.TX, vs.TY, vs.TZ, vs.Scale)
	log.Printf("New header: %v", header)
	fmt.Fprint(w, header)

	ch := make(chan rle, 100)
	go vs.encode(ch, lookup, sx, sy, sz, nx, ny, nz)

	var numVoxels int64
	for v := range ch {
		if v.value == 0 || v.value == 0xffffffff {
			if err := binary.Write(w, binary.LittleEndian, v.value); err != nil {
				return 0, fmt.Errorf("unable to write to buf: %v", err)
			}
			if err := binary.Write(w, binary.LittleEndian, v.count); err != nil {
				return 0, fmt.Errorf("unable to write to buf: %v", err)
			}
			continue
		}
		numVoxels += int64(v.count)
		if numMax := v.count / 255; numMax > 0 {
			if _, err := w.Write(bytes.Repeat([]byte{
				byte(v.value & 0xff),
				byte((v.value >> 8) & 0xff),
				byte((v.value >> 16) & 0xff),
				byte((v.value >> 24) & 0xff),
				255,
			}, int(numMax))); err != nil {
				return 0, fmt.Errorf("unable to write to buf: %v", err)
			}
		}
		if rem := v.count % 255; rem > 0 {
			if _, err := w.Write([]byte{
				byte(v.value & 0xff),
				byte((v.value >> 8) & 0xff),
				byte((v.value >> 16) & 0xff),
				byte((v.value >> 24) & 0xff),
				byte(rem),
			}); err != nil {
				return 0, fmt.Errorf("unable to write buf: %v", err)
			}
		}
	}

	return numVoxels, nil
}

// rle represents the run-length encoding for voxel values.
type rle struct {
	value uint32
	count uint16
}

func (vs *VShell) encode(ch chan<- rle, lookup map[key]int, sx, sy, sz, nx, ny, nz int) {
	var value uint32
	var count uint16
	for xi := sx; xi < sx+nx; xi++ {
		for zi := sz; zi < sz+nz; zi++ {
			for yi := sy; yi < sy+ny; yi++ {
				k := key{xi, yi, zi}
				var newValue uint32
				if v, ok := lookup[k]; ok {
					newValue = uint32(vs.Voxels[v].N)
				}
				if newValue != value {
					if count > 0 { // To account for first value being sent.
						ch <- rle{value, count}
					}
					value = newValue
					count = 1
				} else {
					count++
				}
			}
		}
	}
	if count > 0 { // Output last value.
		ch <- rle{value, count}
	}
	close(ch)
}
