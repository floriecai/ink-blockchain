type Operation struct {
	ShapeHash string
	OpSig     string
	SVGOpType string
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
