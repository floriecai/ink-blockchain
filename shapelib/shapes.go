/*

This file contains functions related to the shapes (Path and Circle)

*/

package shapelib

import "math"

// Alias the math library's functions because I'm lazy
var pow = math.Pow
var sqrt = math.Sqrt

/* PATH_FUNCTIONS */

// Create a new Path struct from a Point slice.
func NewPath(points []Point, filled bool) Path {
	xMin := points[0].X
	xMax := points[0].X
	yMin := points[0].Y
	yMax := points[0].Y

	for i := 1; i < len(points); i++ {
		x := points[i].X
		y := points[i].Y

		if x < xMin {
			xMin = x
		} else if x > xMax {
			xMax = x
		}

		if y < yMin {
			yMin = y
		} else if y > yMax {
			yMax = y
		}
	}

	return Path { points, filled, xMin, xMax, yMin, yMax }
}

// Generate a sub array for the Path object.
// Will fill based on the Filled field of Path.
func (p Path)GetSubArray() PixelSubArray {
	// Create a new sub array that can fit the Path
	sub := NewPixelSubArray(p.XMin, p.XMax, p.YMin, p.YMax)

	// Fill separately from doing the outline - more accurate
	if p.Filled {
		// Initialize some values.
		prevVertDir := -2
		xStart := p.Points[0].X
		yStart := p.Points[0].Y
		yStartFillCount := 0
		yPrev := 0
		_, firstVertDir := linePointsGen(p.Points[0], p.Points[1])

		for i := 0; i < len(p.Points) - 1; i++ {
			if i+2 < len(p.Points) && p.Points[i+1].Moved {
				if yPrev == yStart && ((yStartFillCount % 2) != 1 || prevVertDir == 0) {
					sub.flipAllRight(p.Points[len(p.Points) - 1].X, yPrev)
				}

				yStartFillCount = 0
				xStart = p.Points[i+1].X
				yStart = p.Points[i+1].Y
				prevVertDir = -2
				_, firstVertDir = linePointsGen(p.Points[i+1], p.Points[i+2])

				continue
			}

			yPrev = p.Points[i].Y

			// Get the iterator for pixels along a line between points.
			nextPoint, vertDir := linePointsGen(p.Points[i], p.Points[i+1])

			// Random crap that seems to work.
			if prevVertDir != vertDir && prevVertDir != 0 {
				sub.flipAllRight(p.Points[i].X, yPrev)

				if yPrev == yStart {
					yStartFillCount++
				}
			}

			prevVertDir = vertDir

			// Fill in the pixels provided by the iterator
			for x, y := nextPoint(); x != -1; x, y = nextPoint() {
				if y != yPrev {
					// Effin magic.
					if p.Filled && vertDir != 0 &&
						(!(x == xStart && y == yStart) ||
						yStartFillCount % 2 == 1 || firstVertDir != 0) {
						sub.flipAllRight(x, y)
					}

					if yStartFillCount % 2 == 1 {
						yStartFillCount ++
					}

					yPrev = y
				}
			}
		}

		if yPrev == yStart && ((yStartFillCount % 2) != 1 || prevVertDir == 0) {
			sub.flipAllRight(p.Points[len(p.Points) - 1].X, yPrev)
		}
	}

	// Do the outline of the shape
	for i := 0; i < len(p.Points) - 1; i++ {
		if p.Points[i+1].Moved {
			continue
		}

		// Get the iterator for pixels along a line between points.
		nextPoint, _ := linePointsGen(p.Points[i], p.Points[i+1])

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

// Compute total length of the path
func (p Path)TotalLength() int {
	sum := float64(0)

	for i := 0; i < len(p.Points) - 1; i++ {
		if p.Points[i+1].Moved {
			continue
		}

		x1 := float64(p.Points[i].X)
		x2 := float64(p.Points[i+1].X)
		y1 := float64(p.Points[i].Y)
		y2 := float64(p.Points[i+1].Y)

		sum += sqrt(pow(x2-x1, 2) + pow(y2-y1, 2))
	}

	return int(sum + 0.5)
}

/* CIRCLE_FUNCTIONS */

// Basic. Here in the case that someone doesn't want to
// manually create a circle struct
func NewCircle(xc, yc, radius int, filled bool) Circle {
	return Circle {Point {xc, yc, false}, radius, filled}
}

// Compute 2pi * r
func (c Circle)Circumference() int {
	return int((math.Pi*float64(c.R)*2.0) + 0.5)
}

// Return a PixelSubArray representing the Circle
func (c Circle)GetSubArray() PixelSubArray {
	sub := NewPixelSubArray(c.C.X-c.R, c.C.X+c.R, c.C.Y-c.R, c.C.Y+c.R)

	// Variables are named xLen and yLen because they are relative to c.C;
	// they are not absolute coordinates.
	xLenPrev := c.R
	rSquared := pow(float64(c.R), 2)

	for yLen := 0; yLen <= c.R; yLen++ {
		xLen := int(sqrt(rSquared - pow(float64(yLen), 2)) + 0.5)

		sub.set(c.C.X + xLen, c.C.Y + yLen)
		sub.set(c.C.X + xLen, c.C.Y - yLen)
		sub.set(c.C.X - xLen, c.C.Y - yLen)
		sub.set(c.C.X - xLen, c.C.Y + yLen)

		if c.Filled {
			xLenFill := xLenPrev - 1
			sub.fillBetween(c.C.X - xLenFill, c.C.X + xLenFill, c.C.Y + yLen)
			sub.fillBetween(c.C.X - xLenFill, c.C.X + xLenFill, c.C.Y - yLen)
		}

		for ; xLenPrev > xLen; xLenPrev-- {
			sub.set(c.C.X + xLenPrev, c.C.Y + yLen)
			sub.set(c.C.X + xLenPrev, c.C.Y - yLen)
			sub.set(c.C.X - xLenPrev, c.C.Y - yLen)
			sub.set(c.C.X - xLenPrev, c.C.Y + yLen)
		}
	}

	return sub
}
