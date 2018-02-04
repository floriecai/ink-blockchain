/*

This package is intended to be used to verify that shapes are not conflicting
with each other.

TODO:
 - Circle implementation. Honestly probably easier than Path...
 - Create a function that creates a Path object based on just a
   Point slice

   Search-index:
      TYPE_DEFINITIONS
      FUNCTION_DEFINITIONS
        PIXEL_ARRAY_FUNCTIONS
        PIXEL_SUB_ARRAY_FUNCTIONS
        PATH_FUNCTIONS
*/

package shapelib

import (
	"fmt"
)

/*******************
* TYPE_DEFINITIONS *
*******************/

// Array of pixels. (Max - 1) is the last index that is accessible.
type PixelArray [][]byte

// SubArray that starts at a relative position rather than (0,0)
// xStart should be on a byte boundary, ie. % 8 == 0.
type PixelSubArray struct {
	bytes [][]byte
	xStartByte int
	yStart int
}

// Represents the data of a Path SVG item.
// Any closed shape must have the last point in the
// array be equal to the first point. So a quadrilateral
// should have len(Points) == 5 and  Points[0] is equal
// to Points[4].
type Path struct {
	Points []Point
	Filled bool
	// The below 4 values should create a rectangle that
	// can fit the entire path within it.
	XMin int
	XMax int
	YMin int
	YMax int
}

// Point. Represents a point or pixel on a discrete 2D array.
// All points should be in the 1st quadrant (x >= 0, y >= 0)
type Point struct {
	X int
	Y int
}

// Circle. Not much more to say really.
type Circle struct {
	C Point
	R int
	Filled bool
}

// Used for computing shit for the Path object.
type slopeType int
const (
	POSRIGHT slopeType = iota
	NEGRIGHT
	POSLEFT
	NEGLEFT
	INFUP
	INFDOWN
)

/***********************
* FUNCTION_DEFINITIONS *
************************/

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

/* PIXEL_ARRAY_FUNCTIONS */

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
	// there should be some bitwise & that != 0.
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

// Prints the bits in the array.
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

/* PIXEL_SUB_ARRAY_FUNCTIONS */

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

// Prints the bits in the array. There is no on the screen
// for where the sub-array is meant to be located
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

/* PATH_FUNCTIONS */

// Get slope type, slope, and intercept for a a pair of points
func getLineParams(p1, p2 Point) (sT slopeType, slope, intercept float64) {
	if p1.X == p2.X {
		// Check for infinite slope.
		if p2.Y > p1.Y {
			sT = INFUP
		} else {
			sT = INFDOWN
		}

		slope, intercept = 0, 0
	} else {
		// 4 classifications of non infinite slope based
		// on the relative positions of p1 and p2
		slope, intercept = getSlopeIntercept(p1, p2)
		if p1.X < p2.X {
			if slope >= 0 {
				fmt.Println("POSRIGHT slope")
				sT = POSRIGHT
			} else {
				fmt.Println("NEGRIGHT slope")
				sT = NEGRIGHT
			}
		} else {
			if slope >= 0 {
				fmt.Println("POSLEFT slope")
				sT = POSLEFT
			} else {
				fmt.Println("NEGLEFT slope")
				sT = NEGLEFT
			}
		}
	}

	return sT, slope, intercept
}

// Generates an iterator for a line.  What a mess.
func linePointsGen(p1, p2 Point) (gen func () (x, y int), vertDirection int) {
	// Set up math
	slopeT, slope, intercept := getLineParams(p1, p2)

	x := float64(p1.X)
	xPrev := int(x)
	y := p1.Y
	yThresh := 0

	// Every slope type has a different iterator, since the change the
	// x and y values in different combinations, as well as do different
	// comparisons on the values.
	switch slopeT {
	case POSRIGHT:
		if slope == 0 {
			vertDirection = 0
		} else {
			vertDirection = 1
		}

		return func() (int, int) {
			if y < yThresh {
				if y > p2.Y {
					return -1, -1
				}

				y++
				return xPrev, y
			} else {
				if int(x) > p2.X {
					return -1, -1
				}

				yThresh = int(slope * x + intercept + 0.5)
				xPrev = int(x)
				x++

				if (y != yThresh) {
					y++
				}

				return xPrev, y
			}
		}, vertDirection
	case NEGRIGHT:
		vertDirection = -1
		yThresh = int(slope * x + intercept + 0.5)

		return func () (int, int) {
			if y > yThresh {
				if y < p2.Y {
					return -1, -1
				}

				y--
				return xPrev, y
			} else {
				if int(x) > p2.X {
					return -1, -1
				}

				yThresh = int(slope * x + intercept + 0.5)
				xPrev = int(x)
				x++

				if (y != yThresh) {
					y--
				}

				return xPrev, y
			}
		}, vertDirection
	case POSLEFT:
		if slope == 0 {
			vertDirection = 0
		} else {
			vertDirection = -1
			fmt.Println("POSLEFT, slope:", slope)
		}

		yThresh = int(slope * x + intercept + 0.5)

		return func() (int, int) {
			if y > yThresh {
				if y < p2.Y {
					return -1, -1
				}

				y--
				return xPrev, y
			} else {
				if int(x) < p2.X {
					return -1, -1
				}

				yThresh = int(slope * x + intercept + 0.5)
				xPrev = int(x)
				x--

				if (y != yThresh) {
					y--
				}

				return xPrev, y
			}
		}, vertDirection
	case NEGLEFT:
		vertDirection = 1
		fmt.Println("NEGLEFT, slope:", slope)

		return func () (int, int) {
			if y < yThresh {
				if y > p2.Y {
					return -1, -1
				}

				y++
				return xPrev, y
			} else {
				if int(x) < p2.X {
					return -1, -1
				}

				yThresh = int(slope * x + intercept + 0.5)
				xPrev = int(x)
				x--

				if (y != yThresh) {
					y++
				}

				return xPrev, y
			}
		}, vertDirection
	case INFUP:
		vertDirection = 1

		return func () (int, int) {
			if (y > p2.Y) {
				return -1, -1
			}

			yPrev := y
			y++
			return int(x), yPrev
		}, vertDirection
	case INFDOWN:
		vertDirection = -1

		return func () (int, int) {
			if (y < p2.Y) {
				return -1, -1
			}

			yPrev := y
			y--
			return int(x), yPrev
		}, vertDirection
	}

	return nil, -1
}

// Generate a sub array for the Path object.
// Will fill based on the Filled field of Path.
func (p Path)GetSubArray() PixelSubArray {
	// Create a new sub array that can fit the Path
	sub := NewPixelSubArray(p.XMin, p.XMax, p.YMin, p.YMax)

	// Initialize some values. Need to get start since filling the very
	// last point twice needs to be avoided.
	prevVertDir := -2
	xStart := p.Points[0].X
	yStart := p.Points[0].Y

	// Fill separately from doing the outline - more accurate
	if p.Filled {
		for i := 0; i < len(p.Points) - 1; i++ {
			// Get the iterator for pixels along a line between points.
			nextPoint, vertDir := linePointsGen(p.Points[i], p.Points[i+1])

			// Random crap that seems to work.
			yPrev := p.Points[i].Y
			if prevVertDir != vertDir && prevVertDir != 0 {
				sub.flipAllRight(p.Points[i].X, yPrev)
			}

			prevVertDir = vertDir

			// Fill in the pixels provided by the iterator
			for x, y := nextPoint(); x != -1; x, y = nextPoint() {
				if y != yPrev {
					if p.Filled && vertDir != 0 &&
					!(x == xStart && y == yStart) {
						sub.flipAllRight(x, y)
					}
					yPrev = y
				}
			}
		}
	}

	// Do the outline of the shape
	for i := 0; i < len(p.Points) - 1; i++ {
		// Get the iterator for pixels along a line between points.
		nextPoint, vertDir := linePointsGen(p.Points[i], p.Points[i+1])

		prevVertDir = vertDir
		yPrev := p.Points[i].Y

		// Fill in the pixels provided by the iterator
		for x, y := nextPoint(); x != -1; x, y = nextPoint() {
			if y != yPrev {
				// This set is done to make sure that the pixels
				// are continuous; such as in a 45 degree line.
				sub.set(x, yPrev)
				yPrev = y
			}

			sub.set(x, y)
		}
	}

	return sub
}
