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

// Array of pixels. (Max - 1) is the last index that is accessible.
type PixelArray struct {
	bytes [][]byte
	xMaxByte int
	yMax int
}

// SubArray that starts at a relative position rather than (0,0)
// xStart should be on a byte boundary, ie. % 8 == 0.
type PixelSubArray struct {
	bytes [][]byte
	xStartByte int
	xSizeByte int
	yStart int
	ySize int
}

// Array of Point structs in some order
type Path struct {
	points []Point
	filled bool
}

// Point
type Point struct {
	absX int
	absY int
}

// Circle
type Circle struct {
	c Point
	r int
}

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

// Checks if there is a conflict between the PixelArray and
// a PixelSubArray. Use for checking conflicts.
func (a *PixelArray)HasConflict(sub PixelSubArray) bool {
	// Do some basic validations for overflow
	if (sub.xStartByte + sub.xSizeByte) > a.xMaxByte  {
		fmt.Println("Sub array is past the x boundary")
		return true
	}


	if (sub.yStart + sub.ySize) > a.yMax {
		fmt.Println("Sub array is past the y boundary")
		return true
	}

	xLastByte := sub.xStartByte + sub.xSizeByte

	// Compare the bytes using bitwise &. If there is a conflict,
	// there should be some byte that has issues.
	for y := sub.yStart; y < sub.ySize; y++ {
		ySub := y - sub.yStart

		for x := sub.xStartByte; x < xLastByte; x++ {
			xSub := x - sub.xStartByte

			if (a.bytes[y][x] & sub.bytes[ySub][xSub]) != 0 {
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
	if (sub.xStartByte + sub.xSizeByte) > a.xMaxByte  {
		fmt.Println("Sub array is past the x boundary")
		return
	}


	if (sub.yStart + sub.ySize) > a.yMax {
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

// Returns a new pixel array that is fully zeroed.
func NewPixelArray(xMax int, yMax int) PixelArray {
	// create rows
	a := make([][]byte, yMax)

	// Initialize the number of bytes required in each row
	xSz := maxByte(xMax)
	fmt.Println(xSz)

	for y:= 0; y < yMax; y++ {
		// create compressed columns (one bit per pixel)
		a[y] = make([]byte, xSz)

		// zero fill
		for x := 0; x < xSz; x++ {
			a[y][x] = 0;
		}
	}

	return PixelArray { a, xSz, yMax }
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

	return PixelSubArray { a, xStartByte, xSizeByte, yStart, ySize }
}

// Print a bit array cuz y da fook not.
func PrintArray(a PixelArray) {
	fmt.Printf("%d\n", a.xMaxByte)

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
