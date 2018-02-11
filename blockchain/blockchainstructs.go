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
	PubKey    string
}

type Block struct {
	PrevHash    string
	OpHistory   []Operation
	MinerPubKey string
	Nonce       string
}

type BlockNode struct {
	Block    Block
	Children []int
}