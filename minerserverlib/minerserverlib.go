package minerserverlib

type RServer int

type MinerInfo struct {
	Address net.Addr 
	Key     ecdsa.PublicKey
}
