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
	SVGString string // svgString passed in by the operation
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
	Children []int
}
