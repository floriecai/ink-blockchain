package main

import (
	"crypto/ecdsa"
	"crypto/md5"
	"crypto/x509"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/rpc"
	"time"
	"os"
	"../libminer"
	"../minerserverlib"
)

var MinerInstance *Miner

//Primitive representation of active art miners
var ArtNodeList map[int]bool = make(map[int]bool)

/*******************************
| Structs for the miners to use internally
| note: shared structs should be put in a different lib
********************************/
type Miner struct {
	CurrJobId int
	PrivKey ecdsa.PrivateKey
	Settings libminer.MinerNetSettings
	InkAmt int
}

type MinerInfo struct {
	Address net.Addr
	Key ecdsa.PublicKey
}

type LibMinerInterface struct {

}

type MinerMinerInterface struct {

}

type MinerServerInterface struct {
	Client *rpc.Client
}

/*******************************
| Miner functions
********************************/
func (m *Miner) ConnectToServer(ip string) *net.TCPConn {
	LocalAddr, err := net.ResolveTCPAddr("tcp", ":0")
	CheckError(err)
	ServerAddr, err := net.ResolveTCPAddr("tcp", ip)
	CheckError(err)
	Conn, err := net.DialTCP("tcp", LocalAddr, ServerAddr)
	CheckError(err)
	return Conn
}

/*******************************
| RPC functions
********************************/

//Setup an interface that implements rpc calls for the lib
func OpenLibMinerConn(ip string) {
	lib_miner_int := new(LibMinerInterface)
	server := rpc.NewServer()
	server.Register(lib_miner_int)
	tcp, err := net.Listen("tcp", ip)
	CheckError(err)
	server.Accept(tcp)
}

func (lmi *LibMinerInterface) OpenCanvas(req *libminer.Request, response *libminer.RegisterResponse) (err error){
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		//Generate an id in a basic fashion
		for i := 0;;i++ {
			if !ArtNodeList[i] {
				ArtNodeList[i] = true
				response.Id = i
				response.CanvasXMax = MinerInstance.Settings.CanvasSettings.CanvasXMax
				response.CanvasYMax = MinerInstance.Settings.CanvasSettings.CanvasYMax
				break
			}
		}
		return nil
	} else {
		err = fmt.Errorf("invalid user")
		return err
	}
}

func (lmi *LibMinerInterface) GetInk(req *libminer.Request, response *libminer.InkResponse) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		response.InkRemaining = MinerInstance.InkAmt
		return nil
	} else {
		err = fmt.Errorf("invalid user")
		return err
	}
}

func (lmi *LibMinerInterface) Draw(req *libminer.Request, response *libminer.DrawResponse) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		fmt.Println("drawing is currently unimplemented, sorry!")
		return nil
	} else {
		err = fmt.Errorf("invalid user")
		return err
	}
}

func (lmi *LibMinerInterface) Delete(req *libminer.Request, response *libminer.InkResponse) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		response.InkRemaining = MinerInstance.InkAmt
		fmt.Println("deletion is currently unimplemented, sorry!")
		return nil
	} else {
		err = fmt.Errorf("invalid user")
		return err
	}
}

/* TODO
func (lmi *Lib_Miner_Interface) GetBlockChain(hello *libminer.Request, response *[]Block) (err error) {
	return nil
}
*/

func (lmi *LibMinerInterface) GetGenesisBlock(req *libminer.Request, response *string) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		*response = MinerInstance.Settings.GenesisBlockHash
		return nil
	} else {
		err = fmt.Errorf("invalid user")
		return err
	}
}

// Server

func (msi *MinerServerInterface) Register(m MinerInfo, r *MinerNetSettings) {
	reqArgs := libminer.MinerInfo{Address: m., Key: }
	var resp libMiner.MinerNetSettings
	err := msi.Client.Call("RServer.Register", &reqArgs, &resp)
	CheckError(err)
	return nil
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
func CheckError(err error, parent string) {
	if err != nil {
		fmt.Println(parent, ":: found error! ",err)
	}
}

func ExtractKeyPairs(pubKey, privKey string){
	var PublicKey ecdsa.PublicKey
	var PrivateKey ecdsa.PrivateKey

	pubKeyBytesRestored, _ := hex.DecodeString(pubKey)
	privKeyBytesRestored, _ := hex.DecodeString(privKey)

	PublicKey, _ := x509.ParsePKIXPublicKey(pubKeyBytesRestored)
	PrivateKey, _ := x509.ParseECPrivateKey(privKeyBytesRestored)

	if PublicKey == PrivateKey.PublicKey {
		MinerInstance.PrivKey = PrivateKey
		fmt.Println("ExtractKeyPairs:: Key pair verified")
	}
	fmt.Println("ExtractKeyPairs:: Key pair incorrect, please recheck")
}

/*******************************
| Main
********************************/
func main() {
	serverIP, pubKey, privKey := os.Args[1], os.Args[2], os.Args[3]

	// 1. Setup the singleton miner instance
	MinerInstance = new(Miner)

	// Extract key pairs
	ExtractKeyPairs(pubKey, privKey)

	// TODO: Get MinerNetSettings from server

	ServerConn := MinerInstance.ConnectToServer(serverIP)
	defer ServerConn.Close()

	// 2. Setup Miner-Miner Listener

	// 3. Setup Miner Heartbeat Manager
	// Change interval to 1000ms from 10ms

	// 4. Setup Problem Solving

	// 5. Setup Client-Miner Listener (this thread)
	OpenLibMinerConn(":8080")
}
