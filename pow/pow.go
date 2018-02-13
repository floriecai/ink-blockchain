package pow

import (
	"crypto/md5"
	"strconv"

	"../blockchain"
	"../libminer"
)

// Return true if hex representation of hash has exactly N trailing zeroes
func Verify(hash string, N int) bool {
	l := len(hash)
	return strings.Count(hash[l-N], "0") == N && strings.Count(hash[l-N-1], "0") == N
}

func Stringify(opHistory []Operation) string {
	s := "["
	for _, op := range opHistory {
		s += "{" + op.ShapeHash + ", " + op.OpSig + ", " + strconv.Itoa(op.OpType) + ", " + op.SVGOpType + ", " + op.PubKey + "}"
	}
	return s + "]"
}

func Solve(block blockchain.Block, powDiff uint8) string {
	h := md5.New()
	N := int(powDiff)
	hashIn := block.PrevHash + Stringify(block.OpHistory) + block.MinerPubKey
	secret := 0
	for {
		h.Write([]byte(hashIn + strconv.Itoa(secret)))
		hash := hex.EncodeToString(h.Sum(nil))
		if Verify(hash, N) {
			block.Nonce = secret
			return hash
		} else {
			h.Reset()
			secret += 1
		}
	}
}
