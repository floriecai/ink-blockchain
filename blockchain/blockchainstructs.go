package blockchain

type OpType int

const (
	ADD OpType = iota
	DELETE
)

type ShapeHash struct {
	OpNum uint64 // Unique ID for each shapehash
}

type Operation struct {
	OpType    OpType
	SVGString string // svg "path" that was passed in e.g. M 0 0 H 10 V 20 Z
	Fill      string
	Stroke    string
	OpNum     uint64 // Unique id for operations
}

type OperationInfo struct {
	OpSig  string // The shapehash that we will return
	PubKey string
	Op     Operation
}

type Block struct {
	PrevHash    string
	OpHistory   []OperationInfo
	MinerPubKey string
	Nonce       uint32
}

type BlockNode struct {
	Block    Block
	Children []int // The indices of the children in the BlockNodeArray
}
