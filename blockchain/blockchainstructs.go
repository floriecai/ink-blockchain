type Operation struct {
	ShapeHash string
	OpSig     string
	OpType    int    // 0 for Add, 1 for Delete
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
