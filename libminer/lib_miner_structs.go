package libminer

import (
	"math/big"
	"../blockchain"
	"../shapelib"
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
	SVG         []shapelib.Path
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
	//GenesisBlockHash string
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
