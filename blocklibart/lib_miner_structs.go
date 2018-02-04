package lib_miner_structs

// Msgs used by both blockartlib and miner

//////////////////////////Request msgs

// other messages will be a []byte

type DrawRequest struct {
	Id int
	SVG string
}

type DeleteRequest struct {
	Id int
	ShapeHash string
}

type SignedRequest struct {
	Id []byte
}
//////////////////////////Response msgs

type RegisterResponse {
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