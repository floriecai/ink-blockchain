/*

This file contains functions related to Path.

*/

package shapelib

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
