package binvox

import (
	gl "github.com/fogleman/fauxgl"
)

const dd = 0.5

var (
	// x_1y_1z_1 - do not use
	x_1y_1z0 = gl.Vertex{Position: gl.V(-dd, -dd, 0)}
	x_1y_1z1 = gl.Vertex{Position: gl.V(-dd, -dd, dd)}
	x_1y0z_1 = gl.Vertex{Position: gl.V(-dd, 0, -dd)}
	x_1y0z0  = gl.Vertex{Position: gl.V(-dd, 0, 0)}
	x_1y0z1  = gl.Vertex{Position: gl.V(-dd, 0, dd)}
	// x_1y1z_1  - do not use
	x_1y1z0 = gl.Vertex{Position: gl.V(-dd, dd, 0)}
	// x_1y1z1 - do not use
	x0y_1z_1 = gl.Vertex{Position: gl.V(0, -dd, -dd)}
	x0y_1z0  = gl.Vertex{Position: gl.V(0, -dd, 0)}
	x0y_1z1  = gl.Vertex{Position: gl.V(0, -dd, dd)}
	x0y0z_1  = gl.Vertex{Position: gl.V(0, 0, -dd)}
	x0y0z0   = gl.Vertex{Position: gl.V(0, 0, 0)} // origin
	x0y0z1   = gl.Vertex{Position: gl.V(0, 0, dd)}
	x0y1z_1  = gl.Vertex{Position: gl.V(0, dd, -dd)}
	x0y1z0   = gl.Vertex{Position: gl.V(0, dd, 0)}
	x0y1z1   = gl.Vertex{Position: gl.V(0, dd, dd)}
	// x1y_1z_1 - do not use
	x1y_1z0 = gl.Vertex{Position: gl.V(dd, -dd, 0)}
	x1y_1z1 = gl.Vertex{Position: gl.V(dd, -dd, dd)}
	x1y0z_1 = gl.Vertex{Position: gl.V(dd, 0, -dd)}
	x1y0z0  = gl.Vertex{Position: gl.V(dd, 0, 0)}
	x1y0z1  = gl.Vertex{Position: gl.V(dd, 0, dd)}
	// x1y1z_1 - do not use
	x1y1z0 = gl.Vertex{Position: gl.V(dd, dd, 0)}
	// x1y1z1 - do not use

	singleFace = []*gl.Triangle{ // oriented as a front face
		{V1: x_1y0z_1, V2: x0y0z_1, V3: x0y0z0},
		{V1: x_1y0z_1, V2: x0y0z0, V3: x_1y0z0},

		{V1: x0y0z_1, V2: x1y0z_1, V3: x1y0z0},
		{V1: x0y0z_1, V2: x1y0z0, V3: x0y0z0},

		{V1: x_1y0z0, V2: x0y0z0, V3: x0y0z1},
		{V1: x_1y0z0, V2: x0y0z1, V3: x_1y0z1},

		{V1: x0y0z0, V2: x1y0z0, V3: x1y0z1},
		{V1: x0y0z0, V2: x1y0z1, V3: x0y0z1},
	}

	singleCorner = []*gl.Triangle{ // oriented as back lower left corner
		{V1: x_1y0z_1, V2: x0y0z_1, V3: x0y0z0},
		{V1: x_1y0z_1, V2: x0y0z0, V3: x_1y0z0},
		{V1: x0y0z0, V2: x0y1z0, V3: x_1y1z0},
		{V1: x0y0z0, V2: x_1y1z0, V3: x_1y0z0},
		{V1: x0y0z0, V2: x0y0z_1, V3: x0y1z_1},
		{V1: x0y0z0, V2: x0y1z_1, V3: x0y1z0},
	}

	twoAdjacentCorners = []*gl.Triangle{ // oriented as back lower corners
		{V1: x_1y0z0, V2: x0y0z0, V3: x0y1z0},
		{V1: x_1y0z0, V2: x0y1z0, V3: x_1y1z0},

		{V1: x0y0z0, V2: x1y0z0, V3: x1y1z0},
		{V1: x0y0z0, V2: x1y1z0, V3: x0y1z0},

		{V1: x_1y0z0, V2: x_1y0z_1, V3: x0y0z_1},
		{V1: x_1y0z0, V2: x0y0z_1, V3: x0y0z0},

		{V1: x0y0z0, V2: x0y0z_1, V3: x1y0z_1},
		{V1: x0y0z0, V2: x1y0z_1, V3: x1y0z0},
	}

	twoOppositeLevelCorners = []*gl.Triangle{ // oriented as back lower left to front lower right corners
		{V1: x_1y1z0, V2: x_1y0z0, V3: x0y_1z0},
		{V1: x_1y1z0, V2: x0y_1z0, V3: x1y_1z0},
		{V1: x_1y1z0, V2: x1y_1z0, V3: x1y0z0},
		{V1: x_1y1z0, V2: x1y0z0, V3: x0y1z0},

		{V1: x_1y0z0, V2: x_1y0z_1, V3: x0y_1z_1},
		{V1: x_1y0z0, V2: x0y_1z_1, V3: x0y_1z0},

		{V1: x1y0z0, V2: x1y0z_1, V3: x0y1z_1},
		{V1: x1y0z0, V2: x0y1z_1, V3: x0y1z0},
	}

	twoOppositeDiagonalCorners = []*gl.Triangle{ // oriented as back lower left to front upper right corners
		{V1: x_1y1z0, V2: x_1y0z0, V3: x0y_1z1},
		{V1: x_1y1z0, V2: x0y_1z1, V3: x0y0z1},

		{V1: x_1y0z0, V2: x_1y0z_1, V3: x0y_1z0},
		{V1: x_1y0z0, V2: x0y_1z0, V3: x0y_1z1},

		{V1: x_1y0z_1, V2: x0y0z_1, V3: x1y_1z0},
		{V1: x_1y0z_1, V2: x1y_1z0, V3: x0y_1z0},

		{V1: x0y0z_1, V2: x0y1z_1, V3: x1y0z0},
		{V1: x0y0z_1, V2: x1y0z0, V3: x1y_1z0},

		{V1: x0y1z_1, V2: x0y1z0, V3: x1y0z1},
		{V1: x0y1z_1, V2: x1y0z1, V3: x1y0z0},

		{V1: x0y1z0, V2: x_1y1z0, V3: x0y0z1},
		{V1: x0y1z0, V2: x0y0z1, V3: x1y0z1},
	}

	threeAdjacentCorners = []*gl.Triangle{ // g0 | g1 | g2
		{V1: x_1y0z0, V2: x1y0z0, V3: x1y1z0},
		{V1: x_1y0z0, V2: x1y1z0, V3: x_1y1z0},
		{V1: x_1y0z_1, V2: x0y0z_1, V3: x0y0z0},
		{V1: x_1y0z_1, V2: x0y0z0, V3: x_1y0z0},
		{V1: x0y0z0, V2: x0y0z_1, V3: x0y_1z_1},
		{V1: x0y0z0, V2: x0y_1z_1, V3: x0y_1z0},
		{V1: x0y0z0, V2: x0y_1z0, V3: x1y_1z0},
		{V1: x0y0z0, V2: x1y_1z0, V3: x1y0z0},
	}

	halfCorner = []*gl.Triangle{ // g0 | g1 | g2 | g5
		{V1: x_1y0z_1, V2: x0y0z_1, V3: x0y0z0},
		{V1: x_1y0z_1, V2: x0y0z0, V3: x_1y0z0},
		{V1: x0y0z0, V2: x0y1z0, V3: x_1y1z0},
		{V1: x0y0z0, V2: x_1y1z0, V3: x_1y0z0},
	}

	// unusual cases

	g0357 = []*gl.Triangle{ // g0357_003.blend
		{V1: x_1y0z0, V2: x_1y0z1, V3: x0y0z1},
		{V1: x_1y0z0, V2: x0y0z1, V3: x0y1z1},
		{V1: x_1y0z0, V2: x0y1z1, V3: x_1y1z0},
		{V1: x0y0z_1, V2: x0y1z_1, V3: x1y1z0},
		{V1: x0y0z_1, V2: x1y1z0, V3: x1y0z0},
		{V1: x0y0z0, V2: x0y0z_1, V3: x1y0z0},
		{V1: x0y0z0, V2: x1y0z0, V3: x1y_1z0},
		{V1: x0y0z0, V2: x1y_1z0, V3: x0y_1z0},
		{V1: x0y0z0, V2: x0y_1z0, V3: x0y_1z_1},
		{V1: x0y0z0, V2: x0y_1z_1, V3: x0y0z_1},
	}

	g234 = []*gl.Triangle{ // g234_001.blend
		{V1: x0y0z_1, V2: x_1y0z_1, V3: x_1y1z0},
		{V1: x0y0z_1, V2: x_1y1z0, V3: x0y1z0},
		{V1: x0y0z_1, V2: x0y1z0, V3: x0y0z0},
		{V1: x0y0z_1, V2: x0y0z0, V3: x1y0z_1},
		{V1: x0y0z0, V2: x1y0z0, V3: x1y0z_1},
		{V1: x0y0z0, V2: x1y_1z0, V3: x1y0z0},
		{V1: x0y0z0, V2: x0y_1z0, V3: x1y_1z0},
		{V1: x0y0z0, V2: x0y0z1, V3: x0y_1z0},
		{V1: x0y0z0, V2: x0y1z1, V3: x0y0z1},
		{V1: x0y0z0, V2: x0y1z0, V3: x0y1z1},
		{V1: x0y_1z0, V2: x0y0z1, V3: x_1y0z1},
		{V1: x0y_1z0, V2: x_1y0z1, V3: x_1y_1z0},
	}

	g0156 = []*gl.Triangle{ // g0156_001.blend
		{V1: x_1y0z_1, V2: x1y0z_1, V3: x1y0z0},
		{V1: x_1y0z_1, V2: x1y0z0, V3: x_1y0z0},
		{V1: x0y1z0, V2: x0y_1z0, V3: x0y_1z1},
		{V1: x0y1z0, V2: x0y_1z1, V3: x0y1z1},
		{V1: x0y0z0, V2: x0y1z0, V3: x_1y1z0},
		{V1: x0y0z0, V2: x_1y1z0, V3: x_1y0z0},
		{V1: x0y0z0, V2: x1y0z0, V3: x1y_1z0},
		{V1: x0y0z0, V2: x1y_1z0, V3: x0y_1z0},
	}

	g2457 = []*gl.Triangle{ // g2457_001.blend
		{V1: x1y_1z0, V2: x1y0z1, V3: x0y_1z1},
		{V1: x0y_1z1, V2: x1y0z1, V3: x0y0z1},
		{V1: x0y0z_1, V2: x0y_1z_1, V3: x_1y_1z0},
		{V1: x0y0z_1, V2: x_1y_1z0, V3: x_1y1z0},
		{V1: x0y0z_1, V2: x_1y1z0, V3: x1y1z0},
		{V1: x0y0z_1, V2: x1y1z0, V3: x1y0z_1},
	}

	g12467 = []*gl.Triangle{ // g12467_001.blend
		{V1: x1y0z0, V2: x1y1z0, V3: x0y1z1},
		{V1: x1y0z0, V2: x0y1z1, V3: x0y0z1},
		{V1: x1y0z0, V2: x0y0z1, V3: x1y0z1},

		{V1: x0y0z_1, V2: x0y_1z_1, V3: x0y_1z0},
		{V1: x0y0z_1, V2: x0y_1z0, V3: x0y0z0},
		{V1: x0y0z_1, V2: x0y0z0, V3: x_1y0z0},
		{V1: x0y0z_1, V2: x_1y0z0, V3: x_1y1z0},
		{V1: x0y0z_1, V2: x_1y1z0, V3: x0y1z_1},

		{V1: x0y0z0, V2: x0y_1z0, V3: x_1y_1z0},
		{V1: x0y0z0, V2: x_1y_1z0, V3: x_1y0z0},
	}

	g024567 = []*gl.Triangle{ // g024567_001.blend
		{V1: x_1y0z0, V2: x_1y0z_1, V3: x0y_1z_1},
		{V1: x_1y0z0, V2: x0y_1z_1, V3: x0y_1z0},
		{V1: x_1y0z0, V2: x0y_1z0, V3: x_1y_1z0},

		{V1: x1y0z0, V2: x1y0z_1, V3: x0y1z_1},
		{V1: x1y0z0, V2: x0y1z_1, V3: x0y1z0},
		{V1: x1y0z0, V2: x0y1z0, V3: x1y1z0},
	}

	g13456 = []*gl.Triangle{ // g13456_001.blend
		{V1: x_1y_1z0, V2: x0y_1z1, V3: x0y0z1},
		{V1: x_1y_1z0, V2: x0y0z1, V3: x_1y0z1},

		{V1: x0y_1z_1, V2: x1y0z_1, V3: x1y0z0},
		{V1: x0y_1z_1, V2: x1y0z0, V3: x1y_1z0},

		{V1: x_1y0z_1, V2: x_1y1z0, V3: x0y1z0},
		{V1: x_1y0z_1, V2: x0y1z0, V3: x0y1z_1},
	}

	g0167 = []*gl.Triangle{
		{V1: x1y1z0, V2: x_1y1z0, V3: x_1y0z1},
		{V1: x1y1z0, V2: x_1y0z1, V3: x1y0z1},

		{V1: x_1y0z_1, V2: x1y0z_1, V3: x1y_1z0},
		{V1: x_1y0z_1, V2: x1y_1z0, V3: x_1y_1z0},
	}
)

// {V1: xyz, V2: xyz, V3: xyz},
