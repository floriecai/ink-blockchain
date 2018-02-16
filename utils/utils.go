package utils

import (
	"crypto/ecdsa"
	"crypto/md5"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"../blockchain"
	"../libminer"
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

// Given a blockchain.OperationInfo, returns the corresponding html svg element
// i.e. <path d="M 0 0 H 10 10 v 20 Z" fill="transparent" stroke="red">
func GetHTMLSVGString(op blockchain.Operation) string {
	var fill, stroke string
	if op.OpType == blockchain.DELETE {
		fill = "white"
		stroke = "white"
	} else {
		fill = op.Fill
		stroke = op.Stroke
	}
	return fmt.Sprintf("<path d=\"%s\" fill=\"%s\" stroke=\"%s\">", op.SVGString, fill, stroke)
}

// Parses a string into a list of SVGCommands
// Returns an ordered list of SVGCommands that denote an SVGPath
// Possible Errors:
// - InvalidShapeSvgStringError
// - ShapeSvgStringTooLongError
func GetParsedSVG(svgString string) (svgPath SVGPath, err error) {
	if len(svgString) > MAX_SVG_LEN {
		return svgPath, libminer.ShapeSvgStringTooLongError(svgString)
	}

	tokens := strings.Split(svgString, " ")
	tokenLen := len(tokens)

	i := 0

	svgPath = make(SVGPath, 0)
	for i < tokenLen {
		var svgCommand SVGCommand
		tokenUpper := strings.ToUpper(tokens[i])
		token := tokens[i]

		// Must start with M command
		if i == 0 && tokenUpper != "M" {
			return svgPath, libminer.InvalidShapeSvgStringError(svgString)
		}

		var param1, param2 int
		if tokenUpper == "L" || tokenUpper == "M" {
			if i+2 < tokenLen {
				param1, err = strconv.Atoi(tokens[i+1])
				if err != nil {
					return svgPath, libminer.InvalidShapeSvgStringError(svgString)
				}
				param2, err = strconv.Atoi(tokens[i+2])
				if err != nil {
					return svgPath, libminer.InvalidShapeSvgStringError(svgString)
				}
			}

			if tokenUpper == "L" {
				svgCommand = LCommand{X: param1, Y: param2, IsAbsolute: token != "l"}
				svgPath = append(svgPath, svgCommand)
			} else {
				// M commands are always absolute for our purposes because we can't
				// know the cursor position to get relative coordinates for m
				svgCommand = MCommand{X: param1, Y: param2, IsAbsolute: false}
				svgPath = append(svgPath, svgCommand)
			}

			i += 3
		} else if tokenUpper == "V" || tokenUpper == "H" {
			if i+1 < tokenLen {
				param1, err = strconv.Atoi(tokens[i+1])
				if err != nil {
					return svgPath, libminer.InvalidShapeSvgStringError(svgString)
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
			log.Println("Command does not exist")
			return svgPath, err
		}
	}

	return svgPath, nil
}

// Returns a list of of Points
// Possible Errors:
// - OutOfBoundsError
// - InvalidShapeSvgStringError
func SVGToPoints(svgPath SVGPath, canvasX int, canvasY int, filled bool, strokeFilled bool) (path shapelib.Path, err error) {
	fmt.Println("svgPaths")

	maxX := -1
	maxY := -1
	minX := canvasX + 1
	minY := canvasY + 1

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
		if point.X > canvasX || point.Y > canvasY {
			return path, libminer.OutOfBoundsError{}
		}

		minX = int(math.Min(float64(minX), float64(point.X)))
		minY = int(math.Min(float64(minY), float64(point.Y)))

		maxX = int(math.Max(float64(minX), float64(point.X)))
		maxY = int(math.Max(float64(minY), float64(point.Y)))

		points = append(points, point)
	}

	// Check if any other point other than the first has a "move". If so,
	// the fill check is different.
	var moved = false
	for i := 1; i < len(points); i++ {
		if points[i].Moved {
			moved = true
		}
	}

	// If it is filled, path must represent closed shapes
	if filled {
		if !moved {
			// Normal first-last point check
			lastPoint := points[len(points)-1]
			firstPoint := points[0]

			if firstPoint.X != lastPoint.X || firstPoint.Y != lastPoint.Y {
				return path, libminer.InvalidShapeSvgStringError("")
			}
		} else {
			// Weird stuff here. Need to check each individual shape
			// created between "moves".
			startPoint := points[0]
			prevPoint := points[0]

			for i := 0; i < len(points); i++ {
				if points[i].Moved {
					if startPoint.X != prevPoint.X ||
						startPoint.Y != prevPoint.Y {
						return path, libminer.InvalidShapeSvgStringError("")
					}

					startPoint = points[i]
				}

				prevPoint = points[i]
			}

			// Check the last moved section
			if startPoint.X != prevPoint.X ||
				startPoint.Y != prevPoint.Y {
				return path, libminer.InvalidShapeSvgStringError("")
			}
		}
	}

	path = shapelib.Path{
		XMax:         maxX,
		YMax:         maxY,
		XMin:         minX,
		YMin:         minY,
		Points:       points,
		Filled:       filled,
		StrokeFilled: strokeFilled}

	fmt.Printf("\n\n")
	fmt.Printf("Paths is: %#v", path)
	fmt.Printf("\n\n")
	return path, nil
}

// Return a shapelib.Circle struct from a blockchain operation struct.
// Errors returned:
//    libminer.InvalidShapeSvgStringError
//    libminer.OutOfBoundsError
func GetParsedCirc(op blockchain.Operation, canvasX int, canvasY int) (shapelib.Circle, error) {
	var circ shapelib.Circle

	if op.Fill == "transparent" && op.Stroke == "transparent" {
		return circ, libminer.InvalidShapeSvgStringError(op.SVGString)
	}

	re := regexp.MustCompile(`circle x:(\d+) y:(\d+) r:(\d+)`)
	match := re.FindStringSubmatch(op.SVGString)

	if match == nil {
		return circ, libminer.InvalidShapeSvgStringError(op.SVGString)
	}

	x, err := strconv.ParseUint(match[1], 10, 64)
	if err != nil {
		return circ, libminer.InvalidShapeSvgStringError(op.SVGString)
	}

	y, err := strconv.ParseUint(match[2], 10, 64)
	if err != nil {
		return circ, libminer.InvalidShapeSvgStringError(op.SVGString)
	}

	r, err := strconv.ParseUint(match[3], 10, 64)
	if err != nil {
		return circ, libminer.InvalidShapeSvgStringError(op.SVGString)
	}

	if x+r > uint64(canvasX) || y+r > uint64(canvasY) {
		return circ, libminer.OutOfBoundsError{}
	}

	circ = shapelib.NewCircle(
		int(x),
		int(y),
		int(r),
		op.Fill != "transparent",
		op.Stroke != "transparent")

	return circ, nil
}

func ComputeHash(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

func GetPublicKeyString(pubKey ecdsa.PublicKey) string {
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&pubKey)
	return hex.EncodeToString(publicKeyBytes)
}
