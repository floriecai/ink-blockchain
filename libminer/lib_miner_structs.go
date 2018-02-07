package libminer

import (
	"math/big"
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

type Block struct {
	PrevHash    string
	ReqRecord   Request[]
	MinerPubKey ecdsa.PublicKey
	nonce       uint32
}

type BlocksResponse struct {
	Blocks []Block
}

////////////////////////Settings 

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
	CanvasSettings CanvasSettings
}
