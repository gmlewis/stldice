package binvox

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	gl "github.com/fogleman/fauxgl"
)

// Write writes a binvox file.
// sx, sy, sz are the starting indices of the model.
// nx, ny, nz are the number of voxels to write in each direction (0=all).
func (b *BinVOX) Write(filename string, sx, sy, sz, nx, ny, nz int) error {
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

const headerFMT = `#binvox 1
dim %v %v %v
translate %g %g %g
scale %g
data
`

func (b *BinVOX) write(w io.Writer, sx, sy, sz, nx, ny, nz int) (int64, error) {
	if len(b.Voxels) == 0 {
		fmt.Fprintf(w, headerFMT, 0, 0, 0, b.TX, b.TY, b.TZ, b.Scale)
		return 0, nil
	}

	if nx == 0 {
		nx = b.NX
	}
	if ny == 0 {
		ny = b.NY
	}
	if nz == 0 {
		nz = b.NZ
	}

	header := fmt.Sprintf(headerFMT, nx, ny, nz, b.TX, b.TY, b.TZ, b.Scale)
	log.Printf("New header: %v", header)
	fmt.Fprint(w, header)

	ch := make(chan rle, 100)
	go b.encode(ch, b.Voxels, sx, sy, sz, nx, ny, nz)

	var buf bytes.Buffer
	var numWhite int64
	for v := range ch {
		if v.value == 1 {
			numWhite += v.count
		}
		if numMax := v.count / 255; numMax > 0 {
			if _, err := buf.Write(bytes.Repeat([]byte{v.value, 255}, int(numMax))); err != nil {
				return 0, fmt.Errorf("unable to write to buf: %v", err)
			}
		}
		if rem := v.count % 255; rem > 0 {
			if _, err := buf.Write([]byte{v.value, byte(rem)}); err != nil {
				return 0, fmt.Errorf("unable to write buf: %v", err)
			}
		}
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return 0, fmt.Errorf("unable to write to file: %v", err)
	}

	return numWhite, nil
}

// rle represents the run-length encoding for voxel values.
type rle struct {
	value byte
	count int64
}

func (b *BinVOX) encode(ch chan<- rle, lookup VoxelMap, sx, sy, sz, nx, ny, nz int) {
	var value byte
	var count int64
	for xi := sx; xi < sx+nx; xi++ {
		for zi := sz; zi < sz+nz; zi++ {
			for yi := sy; yi < sy+ny; yi++ {
				k := Key{xi, yi, zi}
				c := gl.Black
				if _, ok := lookup[k]; ok {
					c = gl.White
				}
				var newValue byte
				if c.A > 0 && (c.R > 0.5 || c.G > 0.5 || c.B > 0.5) {
					newValue = 1
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
