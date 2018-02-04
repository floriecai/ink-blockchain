package main

import (
	"./shapelib"
)

func main() {
	// RANDOM SHAPE
//	points := make([]shapelib.Point, 10)
//	points[0] = shapelib.Point {0, 60, false }
//	points[1] = shapelib.Point {150, 80, false }
//	points[2] = shapelib.Point {160, 300, false }
//	points[3] = shapelib.Point {450, 150, false }
//	points[4] = shapelib.Point {600, 250, false }
//	points[5] = shapelib.Point {590, 30, false }
//	points[6] = shapelib.Point {300, 150, false }
//	points[7] = shapelib.Point {150, 30, false }
//	points[8] = shapelib.Point {0, 30, false }
//	points[9] = shapelib.Point {0, 60, false }

//	// HEXAGON
//	points := make([]shapelib.Point, 7 )
//	points[0] = shapelib.Point {350, 300, false }
//	points[1] = shapelib.Point {480, 160, false }
//	points[2] = shapelib.Point {350, 0, false }
//	points[3] = shapelib.Point {140, 0, false }
//	points[4] = shapelib.Point {10, 160, false }
//	points[5] = shapelib.Point {140, 300, false }
//	points[6] = shapelib.Point {350, 300, false }

	// MOVE DOUBLE RECTANGLE
	points := make([]shapelib.Point, 10)
	points[0] = shapelib.Point {0, 0, false }
	points[1] = shapelib.Point {0, 300, false }
	points[2] = shapelib.Point {600, 300, false }
	points[3] = shapelib.Point {600, 0, false }
	points[4] = shapelib.Point {0, 0, false }
	points[5] = shapelib.Point {550, 275, true }
	points[6] = shapelib.Point {50, 275, false }
	points[7] = shapelib.Point {50, 25, false }
	points[8] = shapelib.Point {550, 25, false }
	points[9] = shapelib.Point {550, 275, false }

	path1 := shapelib.NewPath(points, true)
	sub1 := path1.GetSubArray()
	sub1.Print()

	a := shapelib.NewPixelArray(650, 400)
	a.MergeSubArray(sub1)
}
