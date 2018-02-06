package utils

import (
	"crypto/md5"
	"fmt"
	"../shapelib"
	"strconv"
	"strings"
)

const MAX_SVG_LEN = 128

type SVGCommand interface {
	GetX() int
	GetY() int
	IsRelative() bool
}

type MCommand struct {
	IsAbsolute bool
	X          int
	Y          int
}

func (c MCommand) GetX() int { return c.X }
func (c MCommand) GetY() int { return c.Y }

// We don't have relative m commands because we can't get the cursor position without the art node passing it
func (c MCommand) IsRelative() bool { return false }

type LCommand struct {
	IsAbsolute bool
	X          int
	Y          int
}

func (c LCommand) GetX() int        { return c.X }
func (c LCommand) GetY() int        { return c.Y }
func (c LCommand) IsRelative() bool { return !c.IsAbsolute }

type HCommand struct {
	IsAbsolute bool
	Y          int
}

func (c HCommand) GetX() int        { return -1 }
func (c HCommand) GetY() int        { return c.Y }
func (c HCommand) IsRelative() bool { return !c.IsAbsolute }

type VCommand struct {
	IsAbsolute bool
	X          int
}

func (c VCommand) GetX() int        { return c.X }
func (c VCommand) GetY() int        { return -1 }
func (c VCommand) IsRelative() bool { return !c.IsAbsolute }

type ZCommand struct{}

func (c ZCommand) GetX() int        { return -1 }
func (c ZCommand) GetY() int        { return -1 }
func (c ZCommand) IsRelative() bool { return false }

type SVGPath []SVGCommand

// Contains the offending svg string.
type InvalidShapeSvgStringError string

func (e InvalidShapeSvgStringError) Error() string {
	return fmt.Sprintf("BlockArt: Bad shape svg string [%s]", string(e))
}

// Parses a string into a list of SVGCommands
// Returns an ordered list of SVGCommands that denote an SVGPath
// Possible Errors: InvalidShapeSvgStringError
func GetParsedSVG(svgString string) (svgPath SVGPath, err error) {
	if len(svgString) > MAX_SVG_LEN {
		return svgPath, InvalidShapeSvgStringError(svgString)
	}

	// strings.Replace(svgString, ",")
	tokens := strings.Split(svgString, " ")
	tokenLen := len(tokens)
	i := 0
	for i < tokenLen {
		var svgCommand SVGCommand
		tokenUpper := strings.ToUpper(tokens[i])
		token := tokens[i]

		// Must start with M command
		if i == 0 && tokenUpper != "M" {
			return svgPath, InvalidShapeSvgStringError(svgString)
		}

		var param1, param2 int
		if tokenUpper == "L" || tokenUpper == "M" {
			if i+2 < tokenLen {
				param1, err = strconv.Atoi(tokens[i+1])
				if err != nil {
					return svgPath, InvalidShapeSvgStringError(svgString)
				}
				param2, err = strconv.Atoi(tokens[i+2])
				if err != nil {
					return svgPath, InvalidShapeSvgStringError(svgString)
				}
			}

			if tokenUpper == "L" {
				svgCommand = LCommand{X: param1, IsAbsolute: token != "l"}
			} else {
				// M commands are always absolute for our purposes because we can't
				// know the cursor position to get relative coordinates for m
				svgCommand = MCommand{X: param1, Y: param2, IsAbsolute: false}
			}
			i += 3
		} else if tokenUpper == "V" || tokenUpper == "H" {
			if i+1 < tokenLen {
				param1, err = strconv.Atoi(tokens[i+1])
				if err != nil {
					return svgPath, InvalidShapeSvgStringError(svgString)
				}
			}

			if token == "V" {
				svgCommand = VCommand{X: param1, IsAbsolute: token != "v"}
			} else {
				svgCommand = HCommand{Y: param1, IsAbsolute: token != "h"}
			}
			svgPath = append(svgPath, svgCommand)
			i += 2
		} else if tokenUpper == "Z" {
			svgPath = append(svgPath, ZCommand{})
			i++
		} else {
			return svgPath, err
		}
	}

	return svgPath, nil
}

func SVGToPoints(commands SVGPath, canvasX int, canvasY int, filled bool) (path shapelib.Path) {
	//maxX := -1
	//maxY := -1
	//minX := canvasX + 1
	//minY := canvasY + 1

	var points []shapelib.Point

	for i, command := range commands {
		var point shapelib.Point
		switch command.(type) {
		case MCommand:
			point.Moved = true
		default:
			point.Moved = false
		}

		// minX := math.Min(minX, )
		// Commands other than M can be relative
		if command.IsRelative() {
			switch command.(type) {
			case LCommand:
				point.X = points[i-1].X + command.GetX()
				point.Y = points[i-1].Y + command.GetY()
			case VCommand:
				point.X = points[i-1].X + command.GetX()
				point.Y = points[i-1].Y
			case HCommand:
				point.X = points[i-1].X
				point.Y = points[i-1].Y + command.GetY()
			default:
				fmt.Println("Error in svgToPoints: Command isn't relative")
			}
		} else {
			switch command.(type) {
			case ZCommand:
				point.X = points[0].X
				point.Y = points[0].Y
			default:
				point.X = command.GetX()
				point.Y = command.GetY()
			}
		}

		points[i] = point
	}

	path.Points = points
	path.Filled = filled
	return path
}

func ComputeHash(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}
