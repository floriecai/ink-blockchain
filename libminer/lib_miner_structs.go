package libminer

import (
	"math/big"
	"proj1/shapelib"
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
	SVG         shapelib.Path
}

type DeleteRequest struct {
	Id          int
	ValidateNum int
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

//////////////////////////Response msgs
type RegisterResponse struct {
	Id         int
	CanvasXMax uint32
	CanvasYMax uint32
	//GenesisBlockHash string
}

type InkResponse struct {
	InkRemaining int
}

type DrawResponse struct {
	ShapeHash    string
	BlockHash    string
	InkRemaining int
}

type BlocksResponse struct {
	//blocks []Block
	//TODO: Block struct to be completed
}
