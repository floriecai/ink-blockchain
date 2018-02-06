package main

import (
	"fmt"
	"net"
	"net/rpc"
	"../libminer"
	"os"
	"crypto/ecdsa"
	"crypto/md5"
	"math/big"
	"encoding/hex"
	"encoding/json"
)

var MinerInstance *Miner

/*******************************
| Structs for the miners to use internally
| note: shared structs should be put in a different lib
********************************/
type Miner struct {
	CurrJobId int
	PrivKey ecdsa.PrivateKey
	PubKey ecdsa.PublicKey
	GenesisHash string
}

type Lib_Miner_Interface struct {
}

type Miner_Miner_Interface struct {

}
/*******************************
| Miner functions
********************************/
func (m *Miner) ConnectToServer(ip string){

}

/*******************************
| RPC functions
********************************/

//Setup an interface that implements rpc calls for the lib
func OpenLibMinerConn(ip string) {
	lib_miner_int := new(Lib_Miner_Interface)
	server := rpc.NewServer()
	server.Register(lib_miner_int)
	tcp, err := net.Listen("tcp", ip)
	CheckError(err)
	server.Accept(tcp)
}

func (lmi *Lib_Miner_Interface) OpenCanvas(req *libminer.Request, response *libminer.RegisterResponse) (err error){
	return nil
}

func (lmi *Lib_Miner_Interface) GetInk(req *libminer.Request, response *libminer.InkResponse) (err error) {
	return nil
}

func (lmi *Lib_Miner_Interface) Draw(req *libminer.Request, response *libminer.DrawResponse) (err error) {
	return nil
}

func (lmi *Lib_Miner_Interface) Delete(req *libminer.Request, response *libminer.InkResponse) (err error) {
	return nil
}

/* TODO
func (lmi *Lib_Miner_Interface) GetBlockChain(hello *libminer.Request, response *[]Block) (err error) {
	return nil
}
*/

func (lmi *Lib_Miner_Interface) GetGenesisBlock(req *libminer.Request, response *string) (err error) {
	if Verify(req.Msg, req.Sign, req.R, req.S, MinerInstance.PrivKey) {
		*response = MinerInstance.GenesisHash
		return nil
	} else {
		return nil
	}
}

/*******************************
| Helpers
********************************/
func Verify(msg []byte, sign []byte, R, S big.Int, privKey ecdsa.PrivateKey) bool{
	h := md5.New()
	h.Write(msg)
	hash := hex.EncodeToString(h.Sum(nil))
	if hash == hex.EncodeToString(sign) && ecdsa.Verify(&privKey.PublicKey, sign, &R, &S){
		return true
	} else {
		fmt.Println("invalid access\n")
		return false
	}
}
func CheckError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
/*******************************
| Main
********************************/
func main() {
	server_ip, pubKey, privKey := os.Args[1], os.Args[2], os.Args[3]

	// 1. Setup the singleton miner instance
	MinerInstance = new(Miner)

	// extract key pairs TODO: verify this is correct
	var PublicKey ecdsa.PublicKey
	var PrivateKey ecdsa.PrivateKey
	json.Unmarshal([]byte(pubKey), &PublicKey)
	json.Unmarshal([]byte(privKey), &PrivateKey)
	MinerInstance.PubKey = PublicKey
	MinerInstance.PrivKey = PrivateKey

	MinerInstance.ConnectToServer(server_ip)

	// 2. Setup Miner-Miner Listener

	// 3. Setup Miner Heartbeat Manager

	// 4. Setup Problem Solving

	// 5. Setup Client-Miner Listener (this thread)
	OpenLibMinerConn(":8080")
}
