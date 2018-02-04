/*

This package is intended to be used to verify that shapes are not conflicting
with each other.

*/

/*
Fill notes:
- Special case for horizontal lines only: NO-OP for horizontal fill
All other lines: & current position and ^ everything to the right.

Line notes:
- Set the bit in both the floor (floor + 1) of the x-coordinate if the
previous x-coordnate is not the same as the current. This guarantees
that the pixels are continuous, and intersecting lines will always have
a conflict.
*/

package shapelib

import (
	"fmt"
)

// TYPE DEFINITIONS

// Array of pixels. (Max - 1) is the last index that is accessible.
type PixelArray [][]byte

// SubArray that starts at a relative position rather than (0,0)
// xStart should be on a byte boundary, ie. % 8 == 0.
type PixelSubArray struct {
	bytes [][]byte
	xStartByte int
	yStart int
}

// Array of Point structs in some order.
// Any closed shape must have the last point in the
// array be equal to the first point. So a quadrilateral
// should have len(Points) == 5 and  Points[0] is equal
// to Points[4].
type Path struct {
	Points []Point
	Filled bool
	XMin int
	XMax int
	YMin int
	YMax int
}

// Point
type Point struct {
	X int
	Y int
}

// Circle
type Circle struct {
	C Point
	R int
}

/* Function definitions */

// Gets the maximum byte index for a PixelArray's columns.
// Computes the minimum number of bytes needed to contain
// nBits bits.
func maxByte(nBits int) int {
	rem := 0
	if (nBits % 8) != 0 {
		rem = 1
	}

	return (nBits / 8) + rem
}

// Get the slope and y-intercept of a line formed by two points
func getSlopeIntercept(p1 Point, p2 Point) (slope float64, intercept float64) {
	slope = (float64(p2.Y) - float64(p1.Y)) / (float64(p2.X) - float64(p1.X))
	intercept = float64(p1.Y) - slope * float64(p1.X)

	return slope, intercept
}

/* Functions related to pixel arrays */

// Returns a new pixel array that is fully zeroed.
func NewPixelArray(xMax int, yMax int) PixelArray {
	// create rows
	a := make([][]byte, yMax)

	// Initialize the number of bytes required in each row
	xSz := maxByte(xMax)

	for y:= 0; y < yMax; y++ {
		// create compressed columns (one bit per pixel)
		a[y] = make([]byte, xSz)

		// zero fill
		for x := 0; x < xSz; x++ {
			a[y][x] = 0;
		}
	}

	return a
}

// Returns a new pixel sub array.
func NewPixelSubArray(xStart int, xEnd int, yStart int, yEnd int) PixelSubArray {
	// Set up the values for the sub array struct
	xStartByte := xStart / 8
	xSizeByte := maxByte(xEnd) - xStartByte
	ySize := yEnd - yStart + 1

	a := make([][]byte, ySize)

	for y := 0; y < ySize; y++ {
		a[y] = make([]byte, xSizeByte)

		for x := 0; x < xSizeByte; x++ {
			a[y][x] = 0;
		}
	}

	return PixelSubArray { a, xStartByte, yStart }
}

// Checks if there is a conflict between the PixelArray and
// a PixelSubArray.
func (a PixelArray)HasConflict(sub PixelSubArray) bool {
	xLastByte := sub.xStartByte + len(sub.bytes[0])
	yLast := sub.yStart + len(sub.bytes)

	// Do some basic validations for overflow
	if xLastByte > len(a[0])  {
		fmt.Println("Sub array is past the x boundary")
		return true
	}


	if (sub.yStart + len(sub.bytes[0])) > len(a) {
		fmt.Println("Sub array is past the y boundary")
		return true
	}

	// Compare the bytes using bitwise &. If there is a conflict,
	// there should be some byte that has issues.
	for y := sub.yStart; y < yLast; y++ {
		ySub := y - sub.yStart

		for x := sub.xStartByte; x < xLastByte; x++ {
			xSub := x - sub.xStartByte

			if (a[y][x] & sub.bytes[ySub][xSub]) != 0 {
				fmt.Println("Conflict at (x y):", x, y)
				return true
			}
		}
	}

	return false
}

// Applies all of the filled bits in the sub-array to
// the pixel array
func (a *PixelArray)MergeSubArray(sub PixelSubArray) {
	yLast := sub.yStart + len(sub.bytes)

	// Do some basic validations for overflow
	if (sub.xStartByte + len(sub.bytes[0])) > len((*a)[0])  {
		fmt.Println("Sub array is past the x boundary")
		return
	}


	if (sub.yStart + len(sub.bytes)) > len(*a) {
		fmt.Println("Sub array is past the y boundary")
		return
	}

	xLastByte := sub.xStartByte + len(sub.bytes[0])

	for y := sub.yStart; y < yLast; y++ {
		ySub := y - sub.yStart

		for x := sub.xStartByte; x < xLastByte; x++ {
			xSub := x - sub.xStartByte

			(*a)[y][x] |= sub.bytes[ySub][xSub]
		}
	}
}

// Print a bit array cuz y da fook not.
func (a PixelArray)Print() {
	for y := len(a) - 1; y >= 0; y-- {
		fmt.Printf("%d\t", y)
		for x := 0; x < len(a[0]); x++ {
			fmt.Printf("%b%b%b%b%b%b%b%b",
			(a[y][x]) & 1,
			(a[y][x] >> 1) & 1,
			(a[y][x] >> 2) & 1,
			(a[y][x] >> 3) & 1,
			(a[y][x] >> 4) & 1,
			(a[y][x] >> 5) & 1,
			(a[y][x] >> 6) & 1,
			(a[y][x] >> 7) & 1)
		}

		fmt.Printf("\n")
	}
}

// Set the bit on the given co-ordinate
func (a *PixelSubArray)set(x, y int) {
	xByte := x/8 - a.xStartByte
	xBit := uint(x%8)
	yRow := y - a.yStart

	a.bytes[yRow][xByte] |= (1 << xBit)
}

// Flip all of the bits in the sub array to the right
// of the provided coordinate
func (a *PixelSubArray)flipAllRight(x, y int) {
	xBit := uint(x%8)
	xByte := x/8 - a.xStartByte
	yRow := y - a.yStart

	for i := xBit; i < 8; i++ {
		a.bytes[yRow][xByte] ^= (1 << i)
	}

	for i := xByte + 1; i < len(a.bytes[0]); i++ {
		a.bytes[yRow][i] ^= 0xFF
	}
}

func (a PixelSubArray)Print() {
	for y := len(a.bytes) - 1; y >= 0; y-- {
		for x := 0; x < len(a.bytes[0]); x++ {
			fmt.Printf("%b%b%b%b%b%b%b%b",
			(a.bytes[y][x]) & 1,
			(a.bytes[y][x] >> 1) & 1,
			(a.bytes[y][x] >> 2) & 1,
			(a.bytes[y][x] >> 3) & 1,
			(a.bytes[y][x] >> 4) & 1,
			(a.bytes[y][x] >> 5) & 1,
			(a.bytes[y][x] >> 6) & 1,
			(a.bytes[y][x] >> 7) & 1)
		}

		fmt.Printf("\n")
	}
}

/* Functions related to Shapes */

// Generates an iterator for a line.
func linePointsGen(p1, p2 Point) (gen func () (x, y int),
xStart, yStart, direction int) {
	type SlopeType int
	const (
		POS SlopeType = iota
		NEG
		INF
	)

	var slopeType SlopeType
	var slope, intercept float64
	var xEnd, yEnd int

	// Set up math
	if p1.X == p2.X {
		xStart = p1.X
		xEnd = p1.X
		slopeType = INF

		if p2.Y > p1.Y {
			yStart = p1.Y
			yEnd = p2.Y
		} else {
			yStart = p2.Y
			yEnd = p1.Y
		}
	} else {
		if p1.X < p2.X {
			slope, intercept = getSlopeIntercept(p1, p2)

			xStart = p1.X
			xEnd = p2.X
			yStart = p1.Y
			yEnd = p2.Y
		} else {
			slope, intercept = getSlopeIntercept(p2, p1)

			xStart = p2.X
			xEnd = p1.X
			yStart = p2.Y
			yEnd = p1.Y
		}

		if slope >= 0 {
			fmt.Println("POS slope")
			slopeType = POS
		} else {
			fmt.Println("NEG slope")
			slopeType = NEG
		}
	}

	x := float64(xStart)
	y := yStart
	yThresh := 0

	switch slopeType {
	case POS:
		if slope == 0 {
			direction = -1
		} else {
			direction = 1
		}

		return func() (int, int) {
			if (int(x) > xEnd || y > yEnd) {
				fmt.Println("x,y,xend,yend:",x,y,xEnd,yEnd)
				return -1, -1
			} else if y < yThresh {
				y++
				return int(x), y
			} else {
				yThresh = int(slope * x + intercept + 0.5)
				xPrev := int(x)
				x++

				if (y != yThresh) {
					y++
				}

				return xPrev, y
			}
		}, xStart, yStart, direction
	case NEG:
		direction = 0
		yThresh = int(slope * x + intercept + 0.5)

		return func () (int, int) {
			if (int(x) > xEnd || y < yEnd) {
				fmt.Println("x,y,xend,yend:",x,y,xEnd,yEnd)
				return -1, -1
			}

			if y > yThresh {
				y--
				return int(x), y
			} else {
				yThresh = int(slope * x + intercept + 0.5)
				xPrev := int(x)
				x++

				if (y != yThresh) {
					y--
				}

				return xPrev, y
			}
		}, xStart, yStart, direction
	case INF:
		direction = 1

		return func () (int, int) {
			if (int(x) > xEnd || y > yEnd) {
				return -1, -1
			}

			yPrev := y
			y++
			return int(x), yPrev
		}, xStart, yStart, direction
	}

	return nil, -1, -1, -1
}

// Generate a sub array for a shape
func (p Path)GetSubArray() PixelSubArray {
	// Create a new sub array that can fit the Path
	sub := NewPixelSubArray(p.XMin, p.XMax, p.YMin, p.YMax)
	prevDirection := -2

	firstX := p.Points[0].X
	firstY := p.Points[0].Y
	lastYFilled := -1

	for i := 0; i < len(p.Points) - 1; i++ {
		nextPoint, x, prevY, direction := linePointsGen(p.Points[i],
		p.Points[i+1])

		var y int

		prevDirection = direction

		fmt.Println("Next point")
		for x, y = nextPoint(); x != -1; x, y = nextPoint() {
			fmt.Println("Pixel (x,y):", x, y)

			if y != prevY {
				if p.Filled {
					if (direction != prevDirection || y != lastYFilled) && direction != -1 {
						if !(x == firstX && x == firstY) {
						sub.flipAllRight(x, y)
					}
					}
				}

				sub.set(x, prevY)

				prevY = y
			}

			sub.set(x, y)
		}
	}


	return sub
}
