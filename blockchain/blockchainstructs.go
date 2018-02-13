package blockchain

type OpType int

const (
	ADD OpType = iota
	DELETE
)

type Operation struct {
	ShapeHash string
	OpSig     string
	OpType    OpType
	SVGOp     string
	Fill      string
	Stroke    string
	PubKey    string
}

type Block struct {
	PrevHash    string
	ThisHash    string
	OpHistory   []Operation
	MinerPubKey string
	Nonce       string
}

type BlockNode struct {
	Block    Block
	Children []int
}
