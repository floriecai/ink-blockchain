package utils

import (
	"crypto/md5"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"../shapelib"
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
// Returns a list of SVGPath. Each is a subpath, that begins with the M command
// Returns an ordered list of SVGCommands that denote an SVGPath
// Possible Errors: InvalidShapeSvgStringError
func GetParsedSVG(svgString string) (svgPath []SVGPath, err error) {
	if len(svgString) > MAX_SVG_LEN {
		return svgPath, InvalidShapeSvgStringError(svgString)
	}

	tokens := strings.Split(svgString, " ")
	tokenLen := len(tokens)

	// i iterates through all tokens
	// j iterates through each path
	// k iterates through the command of each path
	i, j := 0, -1

	svgPath = make([]SVGPath, 0)
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
				svgPath[j] = append(svgPath[j], svgCommand)
			} else {
				currPath := SVGPath(make([]SVGCommand, 0))
				// M commands are always absolute for our purposes because we can't
				// know the cursor position to get relative coordinates for m
				svgCommand = MCommand{X: param1, Y: param2, IsAbsolute: false}
				currPath = append(currPath, svgCommand)

				svgPath = append(svgPath, currPath)
				j++
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
			svgPath[j] = append(svgPath[j], svgCommand)
			i += 2
		} else if tokenUpper == "Z" {
			svgPath[j] = append(svgPath[j], ZCommand{})
			i++
		} else {
			log.Println("Command does not exist")
			return svgPath, err
		}
	}

	return svgPath, nil
}

// Takes a list of SVGPaths (Each path starts with the M command, could be a list of lines)
// Returns a list of Paths, each path has a list of points
func SVGToPoints(svgPaths []SVGPath, canvasX int, canvasY int, filled bool) (paths []shapelib.Path) {
	fmt.Println("svgPaths")

	maxX := -1
	maxY := -1
	minX := canvasX + 1
	minY := canvasY + 1

	paths = make([]shapelib.Path, len(svgPaths))
	for _, svgPath := range svgPaths {
		// var points []shapelib.Point
		points := make([]shapelib.Point, 0)
		// Path consists of reference
		for i, command := range svgPath {
			var point shapelib.Point
			switch command.(type) {
			case MCommand:
				point.Moved = true
			default:
				point.Moved = false
			}

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

			minX = int(math.Min(float64(minX), float64(point.X)))
			minY = int(math.Min(float64(minY), float64(point.Y)))

			maxX = int(math.Max(float64(minX), float64(point.X)))
			maxY = int(math.Max(float64(minY), float64(point.Y)))

			points = append(points, point)
		}

		path := shapelib.Path{
			XMax:   maxX,
			YMax:   maxY,
			XMin:   minX,
			YMin:   minY,
			Points: points,
			Filled: filled}

		paths = append(paths, path)
	}

	fmt.Printf("\n\n")
	fmt.Printf("Paths is: %#v", paths)
	fmt.Printf("\n\n")
	return paths
}

func ComputeHash(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}
