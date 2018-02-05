package libminer

import (
	"math/big"
)
// Msgs used by both blockartlib and miner

//////////////////////////Request msgs

type Request struct {
	Msg []byte
	Sign []byte
	R big.Int
	S big.Int
}

type DrawRequest struct {
	Id int
	SVG string
}

type DeleteRequest struct {
	Id int
	ShapeHash string
}

//////////////////////////Response msgs

type RegisterResponse struct {
	Id int
	CanvasXMax uint32
	CanvasYMax uint32
	//GenesisBlockHash string
}

type InkResponse struct {
	InkRemaining int
}

type DrawResponse struct {
	ShapeHash string
	BlockHash string
	InkRemaining int
}

type BlocksResponse struct {
	//blocks []Block
	//TODO: Block struct to be completed
}