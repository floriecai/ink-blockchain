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

// Array of Point structs in some order
type Path struct {
	Points []Point
	Filled bool
	XMin int
	Max int
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
	ySize := yEnd - yStart

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
	for y := sub.yStart; y < len(sub.bytes); y++ {
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
	// Do some basic validations for overflow
	if (sub.xStartByte + len(sub.bytes[0])) > len((*a)[0])  {
		fmt.Println("Sub array is past the x boundary")
		return
	}


	if (sub.yStart + len(sub.bytes)) > len(a) {
		fmt.Println("Sub array is past the y boundary")
		return
	}

	xLastByte := sub.xStartByte + sub.xSizeByte

	for y := sub.yStart; y < sub.ySize; y++ {
		ySub := y - sub.yStart

		for x := sub.xStartByte; x < xLastByte; x++ {
			xSub := x - sub.xStartByte

			a.bytes[y][x] &= sub.bytes[ySub][xSub]
		}
	}
}

// Print a bit array cuz y da fook not.
func (a PixelArray)Print() {
	for y := 0; y < a.yMax; y++ {
		for x := 0; x < a.xMaxByte; x++ {
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

