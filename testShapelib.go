package main

import (
	"./shapelib"
)

func main() {
//	// RANDOM SHAPE
//	points := make([]shapelib.Point, 10)
//	points[0] = shapelib.Point {0, 60}
//	points[1] = shapelib.Point {150, 80}
//	points[2] = shapelib.Point {160, 300}
//	points[3] = shapelib.Point {450, 150}
//	points[4] = shapelib.Point {600, 250}
//	points[5] = shapelib.Point {590, 30}
//	points[6] = shapelib.Point {300, 150}
//	points[7] = shapelib.Point {150, 30}
//	points[8] = shapelib.Point {0, 30}
//	points[9] = shapelib.Point {0, 60}

	// HEXAGON
	points := make([]shapelib.Point, 7)
	points[0] = shapelib.Point {10, 160}
	points[1] = shapelib.Point {140, 300}
	points[2] = shapelib.Point {350, 300}
	points[3] = shapelib.Point {480, 160}
	points[4] = shapelib.Point {350, 20}
	points[5] = shapelib.Point {140, 20}
	points[6] = shapelib.Point {10, 160}

	path1 := shapelib.Path{points, true, 0, 610, 0, 308}
	sub1 := path1.GetSubArray()

	a := shapelib.NewPixelArray(650, 400)
	a.MergeSubArray(sub1)
	a.Print()
}
