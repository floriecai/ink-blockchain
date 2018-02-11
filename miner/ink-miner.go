package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/rpc"
	"os"
	"time"
	"../libminer"
	"../minerserver"
)

var MinerInstance *Miner
//Primitive representation of active art miners
var ArtNodeList map[int]bool = make(map[int]bool)

var PeerList map[string]*Peer = make(map[string]*Peer)
/*******************************
| Structs for the miners to use internally
| note: shared structs should be put in a different lib
********************************/
type Miner struct {
	CurrJobId int
	PrivKey   *ecdsa.PrivateKey
	Addr      net.Addr
	Settings  minerserver.MinerNetSettings
	InkAmt    int
	LMI       *LibMinerInterface
	MSI       *MinerServerInterface
}

type MinerInfo struct {
	Address net.Addr
	Key     ecdsa.PublicKey
}

type LibMinerInterface struct {
}

type MinerServerInterface struct {
	Client *rpc.Client
}

type Peer struct {
	Client 			*rpc.Client
	LastHeartBeat 	time.Time
}
/*******************************
| Miner functions
********************************/
func (m *Miner) ConnectToServer(ip string) {
	miner_server_int := new(MinerServerInterface)

	LocalAddr, err := net.ResolveTCPAddr("tcp", ":0")
	CheckError(err, "ConnectToServer:ResolveLocalAddr")

	ServerAddr, err := net.ResolveTCPAddr("tcp", ip)
	CheckError(err, "ConnectToServer:ResolveServerAddr")

	conn, err := net.DialTCP("tcp", LocalAddr, ServerAddr)
	CheckError(err, "ConnectToServer:DialTCP")

	client := rpc.NewClient(conn)
	miner_server_int.Client = client
	m.MSI = miner_server_int
}
/*******************************
| Lib->Miner RPC functions
********************************/

//Setup an interface that implements rpc calls for the lib
func OpenLibMinerConn(ip string) {
	lib_miner_int := new(LibMinerInterface)

	server := rpc.NewServer()
	server.Register(lib_miner_int)

	tcp, err := net.Listen("tcp", ip)
	CheckError(err, "OpenLibMinerConn:Listen")

	MinerInstance.LMI = lib_miner_int

	server.Accept(tcp)
}

func (lmi *LibMinerInterface) OpenCanvas(req *libminer.Request, response *libminer.RegisterResponse) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		//Generate an id in a basic fashion
		for i := 0; ; i++ {
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
		response.InkRemaining = uint32(MinerInstance.InkAmt)
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
		response.InkRemaining = uint32(MinerInstance.InkAmt)
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

/*******************************
| Server Management functions
********************************/

func (msi *MinerServerInterface) Register(minerAddr net.Addr) {
	reqArgs := minerserver.MinerInfo{Address: minerAddr, Key: MinerInstance.PrivKey.PublicKey}
	var resp minerserver.MinerNetSettings
	err := msi.Client.Call("RServer.Register", reqArgs, &resp)
	CheckError(err, "Register:Client.Call")
	MinerInstance.Settings = resp
}

func (msi *MinerServerInterface) ServerHeartBeat() {
	var ignored bool
	err := msi.Client.Call("RServer.HeartBeat", MinerInstance.PrivKey.PublicKey, &ignored)
	CheckError(err, "ServerHeartBeat")
}

func (msi *MinerServerInterface) GetPeers() {
	var addrSet []net.Addr
	var empty empty
	msi.Client.Call("RServer.GetNodes", MinerInstance.PrivKey.PublicKey, &addrSet)
	for _, addr := range addrSet {
		fmt.Println("Calling address: ", addr.String(), "\n")
		LocalAddr, err := net.ResolveTCPAddr("tcp", ":0")
		CheckError(err, "GetPeers:ResolvePeerAddr")

		PeerAddr, err := net.ResolveTCPAddr("tcp", addr.String())
		CheckError(err, "GetPeers:ResolveLocalAddr")

		conn, err := net.DialTCP("tcp", LocalAddr, PeerAddr)
		CheckError(err, "GetPeers:DialTCP")

		client := rpc.NewClient(conn)

		err = client.Call("peerRPC.Connect", LocalAddr.String(), &empty)
		CheckError(err, "GetPeers:Connect")

		PeerList[addr.String()] = &Peer{client, time.Now()}
	}
}
/*******************************
| Connection Management
********************************/
// This function has 4 purposes:
// 1. Send the server heartbeat to maintain connectivity
// 2. Send miner heartbeats to maintain connectivity with peers
// 3. Check for stale peers and remove them from the list
// 4. Request new nodes from server and connect to them when peers drop too low
// This is the central point of control for the peer connectivity
func ManageConnections() {
	// Send heartbeats at four times the timeout interval to be safe
	interval := time.Duration(MinerInstance.Settings.HeartBeat / 4)
	heartbeat := time.Tick(interval * time.Millisecond)
	for {
		select {
		case <- heartbeat:
			MinerInstance.MSI.ServerHeartBeat()
			PeerHeartBeats()
		default:
			CheckLiveliness()
			if len(PeerList) < int(MinerInstance.Settings.MinNumMinerConnections) {
				MinerInstance.MSI.GetPeers()
			}
		}
	}
}

// Send a heartbeat call to each peer
func PeerHeartBeats() {
	for addr, peer := range PeerList {
		empty := new(empty)
		err := peer.Client.Call("Peer.Hb", &empty, &empty)
		CheckError(err, "PeerHeartBeats:"+addr)
	}
}
// Look through current active connections and delete them if they are not live
func CheckLiveliness() {
	interval := time.Duration(MinerInstance.Settings.HeartBeat)
	for addr, peer := range PeerList {
		if time.Since(peer.LastHeartBeat) > interval {
			fmt.Println("Stale connection: ", addr, " deleting")
			delete(PeerList, addr)
		}
	}
}
/*******************************
| Helpers
********************************/
func Verify(msg []byte, sign []byte, R, S big.Int, privKey *ecdsa.PrivateKey) bool {
	h := md5.New()
	h.Write(msg)
	hash := hex.EncodeToString(h.Sum(nil))
	if hash == hex.EncodeToString(sign) && ecdsa.Verify(&privKey.PublicKey, sign, &R, &S) {
		return true
	} else {
		fmt.Println("invalid access\n")
		return false
	}
}
func CheckError(err error, parent string) {
	if err != nil {
		fmt.Println(parent, ":: found error! ", err)
	}
}

func ExtractKeyPairs(pubKey, privKey string) {
	var PublicKey *ecdsa.PublicKey
	var PrivateKey *ecdsa.PrivateKey

	pubKeyBytesRestored, _ := hex.DecodeString(pubKey)
	privKeyBytesRestored, _ := hex.DecodeString(privKey)

	pub, err := x509.ParsePKIXPublicKey(pubKeyBytesRestored)
	CheckError(err, "ExtractKeyPairs:ParsePKIXPublicKey")
	PublicKey = pub.(*ecdsa.PublicKey)

	PrivateKey, err = x509.ParseECPrivateKey(privKeyBytesRestored)
	CheckError(err, "ExtractKeyPairs:ParseECPrivateKey")

	r, s, _ := ecdsa.Sign(rand.Reader, PrivateKey, []byte("data"))

	if !ecdsa.Verify(PublicKey, []byte("data"), r, s) {
		fmt.Println("ExtractKeyPairs:: Key pair incorrect, please recheck")
	}
	MinerInstance.PrivKey = PrivateKey
	fmt.Println("ExtractKeyPairs:: Key pair verified")
}

func pubKeyToString(key ecdsa.PublicKey) string {
	return string(elliptic.Marshal(key.Curve, key.X, key.Y))
}
/*******************************
| Main
********************************/
func main() {
	gob.Register(&net.TCPAddr{})
	gob.Register(&elliptic.CurveParams{})
	serverIP, pubKey, privKey := os.Args[1], os.Args[2], os.Args[3]

	// 1. Setup the singleton miner instance
	MinerInstance = new(Miner)

	// TODO - Undo the hardcoding after we're done testing
	ln, _ := net.Listen("tcp", ":8080")
	addr := ln.Addr()

	MinerInstance.Addr = addr
	// Extract key pairs
	ExtractKeyPairs(pubKey, privKey)

	// Connect to Server
	MinerInstance.ConnectToServer(serverIP)
	MinerInstance.MSI.Register(addr)

	// 2. Setup Miner-Miner Listener

	// 3. Setup Miner Heartbeat Manager
	// Change interval to 1000ms from 10ms
	go ManageConnections()

	// 4. Setup Problem Solving

	// 5. Setup Client-Miner Listener (this thread)
	OpenLibMinerConn(":8090")
}
