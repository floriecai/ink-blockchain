package libminer

import (
	"fmt"
	"math/big"

	"../blockchain"
)

// Msgs used by both blockartlib and miner

//////////////////////////Request msgs

type Request struct {
	Msg       []byte
	HashedMsg []byte
	R         big.Int
	S         big.Int
}

type DrawRequest struct {
	Id          int
	ValidateNum uint8
	SVGString   string
	Fill        string
	Stroke      string
}

type DeleteRequest struct {
	Id          int
	ValidateNum uint8
	ShapeHash   string
}

type GenericRequest struct {
	Id int
}

type RegisterRequest struct {
	R   big.Int
	S   big.Int
	Msg []byte
}

type BlockRequest struct {
	Id        int
	BlockHash string
}

//////////////////////////Response msgs
type RegisterResponse struct {
	Id         int
	CanvasXMax uint32
	CanvasYMax uint32
}

type InkResponse struct {
	InkRemaining uint32
}

type DrawResponse struct {
	ShapeHash    string
	BlockHash    string
	InkRemaining uint32
}

type BlocksResponse struct {
	Blocks []blockchain.Block
}

/*********** ERRORS ************/
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

/*********** ERRORS ************/
