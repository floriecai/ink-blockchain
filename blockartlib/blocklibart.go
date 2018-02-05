/*

This package specifies the application's interface to the the BlockArt
library (blockartlib) to be used in project 1 of UBC CS 416 2017W2.

*/

package blockartlib

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"strings"
)

const MAX_SVG_LEN = 128

// Represents a type of shape in the BlockArt system.
type ShapeType int

const (
	// Path shape.
	PATH ShapeType = iota

	// Circle shape (extra credit).
	// CIRCLE
)

// Settings for a canvas in BlockArt.
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32
	CanvasYMax uint32
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string

	// The minimum number of ink miners that an ink miner should be
	// connected to. If the ink miner dips below this number, then
	// they have to retrieve more nodes from the server using
	// GetNodes().
	MinNumMinerConnections uint8

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32
	InkPerNoOpBlock uint32

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8
	PoWDifficultyNoOpBlock uint8

	// Canvas settings
	canvasSettings CanvasSettings
}

////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// These type definitions allow the application to explicitly check
// for the kind of error that occurred. Each API call below lists the
// errors that it is allowed to raise.
//
// Also see:
// https://blog.golang.org/error-handling-and-go
// https://blog.golang.org/errors-are-values

// Contains address IP:port that art node cannot connect to.
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("BlockArt: cannot connect to [%s]", string(e))
}

// Contains amount of ink remaining.
type InsufficientInkError uint32

func (e InsufficientInkError) Error() string {
	return fmt.Sprintf("BlockArt: Not enough ink to addShape [%d]", uint32(e))
}

// Contains the offending svg string.
type InvalidShapeSvgStringError string

func (e InvalidShapeSvgStringError) Error() string {
	return fmt.Sprintf("BlockArt: Bad shape svg string [%s]", string(e))
}

// Contains the offending svg string.
type ShapeSvgStringTooLongError string

func (e ShapeSvgStringTooLongError) Error() string {
	return fmt.Sprintf("BlockArt: Shape svg string too long [%s]", string(e))
}

// Contains the bad shape hash string.
type InvalidShapeHashError string

func (e InvalidShapeHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid shape hash [%s]", string(e))
}

// Contains the bad shape hash string.
type ShapeOwnerError string

func (e ShapeOwnerError) Error() string {
	return fmt.Sprintf("BlockArt: Shape owned by someone else [%s]", string(e))
}

// Empty
type OutOfBoundsError struct{}

func (e OutOfBoundsError) Error() string {
	return fmt.Sprintf("BlockArt: Shape is outside the bounds of the canvas")
}

// Contains the hash of the shape that this shape overlaps with.
type ShapeOverlapError string

func (e ShapeOverlapError) Error() string {
	return fmt.Sprintf("BlockArt: Shape overlaps with a previously added shape [%s]", string(e))
}

// Contains the invalid block hash.
type InvalidBlockHashError string

func (e InvalidBlockHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid block hash [%s]", string(e))
}

// </ERROR DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////////////////

// Represents a canvas in the system.
type Canvas interface {
	// Adds a new shape to the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - InsufficientInkError
	// - InvalidShapeSvgStringError
	// - ShapeSvgStringTooLongError
	// - ShapeOverlapError
	// - OutOfBoundsError
	AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error)
	// aDD SHAPE blocks until number of blocks (validateNum) follow current block

	// Returns the encoding of the shape as an svg string.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidShapeHashError
	GetSvgString(shapeHash string) (svgString string, err error)

	// Returns the amount of ink currently available.
	// Can return the following errors:
	// - DisconnectedError
	GetInk() (inkRemaining uint32, err error)

	// Removes a shape from the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - ShapeOwnerError
	DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error)

	// Retrieves hashes contained by a specific block.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetShapes(blockHash string) (shapeHashes []string, err error)

	// Returns the block hash of the genesis block.
	// Can return the following errors:
	// - DisconnectedError
	GetGenesisBlock() (blockHash string, err error)

	// Retrieves the children blocks of the block identified by blockHash.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetChildren(blockHash string) (blockHashes []string, err error)

	// Closes the canvas/connection to the BlockArt network.
	// - DisconnectedError
	CloseCanvas() (inkRemaining uint32, err error)
}

type SVGCommand interface {
	isExpr()
}

type MCommand struct {
	IsRelative bool
	X          int
	Y          int
}

func (c MCommand) isExpr() {}

type LCommand struct {
	IsRelative bool
	X          int
	Y          int
}

func (c LCommand) isExpr() {}

type HCommand struct {
	IsRelative bool
	Y          int
}

func (c HCommand) isExpr() {}

type VCommand struct {
	IsRelative bool
	X          int
}

func (c VCommand) isExpr() {}

type ZCommand struct{}

func (c ZCommand) isExpr() {}

type SVGPath []SVGCommand

// Parses a string into a list of SVGCommands
// Returns an ordered list of SVGCommands that denote an SVGPath
// Possible Errors: InvalidShapeSvgStringError
func getParsedSVG(svgString string) (svgPath SVGPath, err error) {
	if len(svgString) > MAX_SVG_LEN {
		return svgPath, InvalidShapeSvgStringError(svgString)
	}

	tokens := strings.Split(svgString, " ")
	tokenLen := len(tokens)
	i := 0
	for i < tokenLen {
		var svgCommand SVGCommand
		tokenUpper := strings.ToUpper(tokens[i])
		token := tokens[i]

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
				svgCommand = LCommand{X: param1, IsRelative: token == "l"}
			} else {
				svgCommand = MCommand{X: param1, Y: param2, IsRelative: token == "m"}
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
				svgCommand = VCommand{X: param1, IsRelative: token == "v"}
			} else {
				svgCommand = HCommand{Y: param1, IsRelative: token == "h"}
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

// The constructor for a new Canvas object instance. Takes the miner's
// IP:port address string and a public-private key pair (ecdsa private
// key type contains the public key). Returns a Canvas instance that
// can be used for all future interactions with blockartlib.
//
// The returned Canvas instance is a singleton: an application is
// expected to interact with just one Canvas instance at a time.
//
// Can return the following errors:
// - DisconnectedError
func OpenCanvas(minerAddr string, privKey ecdsa.PrivateKey) (canvas Canvas, setting CanvasSettings, err error) {
	// TODO
	// For now return DisconnectedError
	return nil, CanvasSettings{}, DisconnectedError("")
}
