package pow

import (
	"crypto/md5"
	"strconv"
	"../blockchain"
	"../libminer"
	"encoding/json"
)

// Return true if hex representation of hash has exactly N trailing zeroes
func Verify(hash string, N int) bool {
	l := len(hash)
	return strings.Count(hash[l-N], "0") == N && strings.Count(hash[l-N-1], "0") == N
}

func Solve(block blockchain.Block, powDiff uint8, start uint64, out chan) string {
	h := md5.New()
	N := int(powDiff)
	hashIn := json.Marshal(block)
	secret := 0
	for {
		h.Write([]byte(hashIn + strconv.Itoa(secret)))
		hash := hex.EncodeToString(h.Sum(nil))
		if Verify(hash, N) {
			block.Nonce = secret
			out <- hash
		} else {
			h.Reset()
			secret += 1
		}
	}
}
