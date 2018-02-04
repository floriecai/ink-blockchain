/*

This file contains functions related to Path.

*/

package shapelib

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

	return Path { points, filled, xMin, xMax + 1, yMin, yMax + 1 }
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

		for i := 0; i < len(p.Points) - 1; i++ {
			yPrev = p.Points[i].Y

			if p.Points[i+1].Moved {
				if prevVertDir == 0 && yPrev == yStart {
					sub.flipAllRight(p.Points[i].X, yPrev)
				}

				yStartFillCount = 0
				xStart = p.Points[i+1].X
				yStart = p.Points[i+1].Y
				continue
			}

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
						yStartFillCount % 2 == 1) {
						sub.flipAllRight(x, y)
					}

					yPrev = y
				}
			}
		}

		if prevVertDir == 0 && yPrev == yStart {
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
