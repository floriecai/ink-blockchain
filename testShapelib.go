package main

import (
	"./shapelib"
)

func main() {
	points := make([]shapelib.Point, 10)
	points[0] = shapelib.Point {0, 5}
	points[1] = shapelib.Point {15, 8}
	points[2] = shapelib.Point {30, 30}
	points[3] = shapelib.Point {45, 15}
	points[4] = shapelib.Point {60, 25}
	points[5] = shapelib.Point {45, 1}
	points[6] = shapelib.Point {30, 15}
	points[7] = shapelib.Point {15, 1}
	points[8] = shapelib.Point {0, 1}
	points[9] = shapelib.Point {0, 5}

	path1 := shapelib.Path{points, false, 0, 100, 1, 30}
	sub1 := path1.GetSubArray()

	a := shapelib.NewPixelArray(100, 40)
	a.MergeSubArray(sub1)
	a.Print()

	path2 := shapelib.Path{points, true, 5, 100, 1, 30}
	sub2 := path2.GetSubArray()
	a = shapelib.NewPixelArray(100, 40)
	a.MergeSubArray(sub2)
	a.Print()
}
