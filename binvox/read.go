package binvox

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
)

var (
	dimRE       = regexp.MustCompile(`^dim (\d+) (\d+) (\d+)\s*$`)
	translateRE = regexp.MustCompile(`^translate (\S+) (\S+) (\S+)\s*$`)
	scaleRE     = regexp.MustCompile(`^scale (\S+)\s*$`)
)

// Read reads a binvox file and returns a BinVOX.
// sx, sy, sz are the starting indices for reading a model.
// nx, ny, nz are the number of voxels to read in each direction (0=all).
func Read(filename string, sx, sy, sz, nx, ny, nz int) (*BinVOX, error) {
	log.Printf("Loading file %q...", filename)
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %v", filename, err)
	}
	defer f.Close()
	binVOX, err := read(f, sx, sy, sz, nx, ny, nz)
	if err != nil {
		return nil, fmt.Errorf("Read(%q): %v", filename, err)
	}
	log.Printf("Done loading %v voxels from file %q.", len(binVOX.Voxels), filename)
	return binVOX, nil
}

func read(r io.Reader, sx, sy, sz, cx, cy, cz int) (*BinVOX, error) {
	if sx < 0 || sy < 0 || sz < 0 || cx < 0 || cy < 0 || cz < 0 {
		return nil, fmt.Errorf("invalid parameters: start=(%v,%v,%v) count=(%v,%v,%v)", sx, sy, sz, cx, cy, cz)
	}
	b := bufio.NewReader(r)

	header, err := b.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read header: %v", err)
	}
	if header != "#binvox 1\n" {
		return nil, fmt.Errorf("not a binvox file: %v", header)
	}

	dim, err := b.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read dimensions: %v", err)
	}
	parts := dimRE.FindStringSubmatch(dim)
	if len(parts) != 4 {
		return nil, fmt.Errorf("unable to parse dimensions: %v", dim)
	}
	nx, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse dimensions: %v", dim)
	}
	ny, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("unable to parse dimensions: %v", dim)
	}
	nz, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("unable to parse dimensions: %v", dim)
	}
	if sx > nx || sy > ny || sz > nz {
		return nil, fmt.Errorf("invalid parameters: start=(%v,%v,%v) model dim=(%v,%v,%v)", sx, sy, sz, nx, ny, nz)
	}
	if cx == 0 || cx > nx {
		cx = nx
	}
	if cy == 0 || cy > ny {
		cy = ny
	}
	if cz == 0 || cz > nz {
		cz = nz
	}

	translate, err := b.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read translate: %v", err)
	}
	parts = translateRE.FindStringSubmatch(translate)
	if len(parts) != 4 {
		return nil, fmt.Errorf("unable to parse translation: %v", translate)
	}
	tx, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse translation: %v", translate)
	}
	ty, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse translation: %v", translate)
	}
	tz, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse translation: %v", translate)
	}

	scaleLine, err := b.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read scale: %v", err)
	}
	parts = scaleRE.FindStringSubmatch(scaleLine)
	if len(parts) != 2 {
		return nil, fmt.Errorf("unable to parse scale: %v", scaleLine)
	}
	scale, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse scale: %v", scaleLine)
	}

	data, err := b.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read data: %v", err)
	}
	if data != "data\n" {
		return nil, fmt.Errorf("could not find data section: %v", data)
	}
	log.Printf("binvox dim=(%v,%v,%v) translate=(%v,%v,%v), uniform scale=%v", nx, ny, nz, tx, ty, tz, scale)

	// Read run-length encoded data.
	var xi, yi, zi int
	voxels := VoxelMap{}
	for {
		value, err := b.ReadByte()
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("error reading value: %v", err)
			}
			break
		}
		if value != 0 && value != 1 {
			return nil, fmt.Errorf("invalud value: %v", value)
		}
		count, err := b.ReadByte()
		if err != nil { // Should not EOF when getting count, so return error
			return nil, fmt.Errorf("error reading count: %v", err)
		}
		if count == 0 {
			return nil, fmt.Errorf("invalid count: %v", count)
		}

		// The y-coordinate runs fastest, then the z-coordinate, then the x-coordinate.
		// Note that we are only saving the white pixels.
		for i := 0; i < int(count); i++ {
			if xi >= nx {
				return nil, fmt.Errorf("run-length encoding overrun: x index=%v, x dim=%v", xi, nx)
			}
			if value == 1 && xi >= sx && xi <= sx+cx && yi >= sy && yi <= sy+cy && zi >= sz && zi <= sz+cz {
				voxels[Key{X: xi, Y: yi, Z: zi}] = White
			}
			yi++
			if yi >= ny {
				yi = 0
				zi++
				if zi >= nz {
					zi = 0
					xi++
				}
			}
		}
	}

	return &BinVOX{
		NX: nx, NY: ny, NZ: nz,
		TX: tx, TY: ty, TZ: tz,
		Scale:  scale,
		Voxels: voxels,
	}, nil
}
