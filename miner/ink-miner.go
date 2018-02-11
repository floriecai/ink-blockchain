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
	"../blockchain"
	"../libminer"
	"../minerserver"
	"../pow"
)

// Our singleton miner instance
var MinerInstance *Miner

// Primitive representation of active art miners
var ArtNodeList map[int]bool = make(map[int]bool)

// List of peers WE connect TO, not peers that connect to US
var PeerList map[string]*Peer = make(map[string]*Peer)

// Global TTL of propagate requests
var TTL int = 100

// Global block chain array
var BlockNodeArray []blockchain.BlockNode

// Global block chain search map
// Key: The hash of a block
// Val: The index of block with such hash in BlockNodeArray
var BlockHashMap map[string]int = make(map[string]int)

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

	fmt.Println("ConnectToServer::connecting to server on:", conn.LocalAddr().String())

	client := rpc.NewClient(conn)
	miner_server_int.Client = client
	m.MSI = miner_server_int
}
/*******************************
| Lib->Miner RPC functions
********************************/

// Setup an interface that implements rpc calls for the lib
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

func (lmi *LibMinerInterface) GetGenesisBlock(req *libminer.Request, response *string) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		*response = MinerInstance.Settings.GenesisBlockHash
		return nil
	} else {
		err = fmt.Errorf("invalid user")
		return err
	}
}

func (lmi *LibMinerInterface) GetChildren(req *libminer.Request, response *libminer.BlocksResponse) (err error) {
    if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		children := GetBlockChildren(req.BlockHash)
		response.Blocks = children
        return nil
    } else {
        err = fmt.Errorf("invalid user")
        return err
    }
}

/*******************************
| Blockchain functions
********************************/

// Appends the new block to BlockArray and updates BlockHashMap
func InsertBlock(newBlock blockchain.Block) (err error) {
	// Create a new node for newBlock and append it to BlockNodeArray
	newBlockNode := blockchain.BlockNode{Block: newBlock, Children: []int{}}
	BlockNodeArray = append(BlockNodeArray, newBlockNode)
	// Create an entry for newBlock in BlockHashMap
	childIndex := len(BlockNodeArray) - 1
	childHash := GetBlockHash(newBlock)
	if VerifyBlock(newBlock) {
		BlockHashMap[childHash] = childIndex
		// Update the entry for newBlock's parent in BlockNodeArray
		parentIndex := BlockHashMap[newBlock.PrevHash]
		parentBlock := BlockNodeArray[parentIndex].Block
		parentBlock.Children = append(parentBlock.Children, childIndex)
		return nil
	} else {
		err = fmt.Errorf("block hash does not match up with block contents")
		return err
	}
}

// Do we need this?
// It seems like the only block individually retrieved is the GenesisBlock
func GetBlock(blockHash string) blockchain.Block {
	index := BlockHashMap[blockHash]
	return BlockNodeArray[index].Block
}

func GetBlockChildren(blockHash string) []blockchain.Block {
	var children []blockchain.Block
	parentIndex := BlockHashMap[blockHash]
	for _, childIndex := range BlockNodeArray[parentIndex].Children {
		children = append(children, BlockNodeArray[childIndex].Block)
	}
	return children
}

func GetBlockHash(block blockchain.Block) string {
	h := md5.New()
	hashIn := pow.Stringify(block) + block.Nonce
	h.Write([]byte(hashIn))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifyBlock(block blockchain.Block) bool {
	hash := GetBlockHash(block)
	if len(block.OpHistory) == 0 {
		return pow.Verify(hash, MinerInstance.Settings.PoWDifficultyNoOpBlock)
	} else { 
		return pow.Verify(hash, MinerInstance.Settings.PoWDifficultyOpBlock)
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
	fmt.Println("ServerHeartBeat::Sending heartbeat")
	err := msi.Client.Call("RServer.HeartBeat", MinerInstance.PrivKey.PublicKey, &ignored)
	if CheckError(err, "ServerHeartBeat"){
		//Reconnect to server if timed out
		msi.Register(MinerInstance.Addr)
	}
}

func (msi *MinerServerInterface) GetPeers() {
	var addrSet []net.Addr
	var empty Empty
	msi.Client.Call("RServer.GetNodes", MinerInstance.PrivKey.PublicKey, &addrSet)
	for _, addr := range addrSet {
		if _, ok := PeerList[addr.String()]; !ok {
			fmt.Println("GetPeers::Connecting to address: ", addr.String())
			LocalAddr, err := net.ResolveTCPAddr("tcp", ":0")
			if CheckError(err, "GetPeers:ResolvePeerAddr"){
				continue
			}

			PeerAddr, err := net.ResolveTCPAddr("tcp", addr.String())
			if CheckError(err, "GetPeers:ResolveLocalAddr"){
				continue
			}

			conn, err := net.DialTCP("tcp", LocalAddr, PeerAddr)
			if CheckError(err, "GetPeers:DialTCP"){
				continue
			}

			client := rpc.NewClient(conn)

			args := ConnectArgs{conn.LocalAddr().String()}
			err = client.Call("Peer.Connect", args, &empty)
			if CheckError(err, "GetPeers:Connect"){
				continue
			}

			PeerList[addr.String()] = &Peer{client, time.Now()}
		}
	}
}
/*******************************
| Connection Management
********************************/
// This function has 5 purposes:
// 1. Send the server heartbeat to maintain connectivity
// 2. Send miner heartbeats to maintain connectivity with peers
// 3. Check for stale peers and remove them from the list
// 4. Request new nodes from server and connect to them when peers drop too low
// 5. When a operation or block is sent through the channel, heartbeat will be replaced by Propagate<Type>
// This is the central point of control for the peer connectivity
func ManageConnections(pop chan blockchain.Operation, pblock chan blockchain.Block) {
	// Send heartbeats at three times the timeout interval to be safe
	interval := time.Duration(MinerInstance.Settings.HeartBeat / 3)
	heartbeat := time.Tick(interval * time.Millisecond)
	for {
		select {
		case <- heartbeat:
			MinerInstance.MSI.ServerHeartBeat()
			PeerHeartBeats()
		case op := <- pop:
			MinerInstance.MSI.ServerHeartBeat()
			PeerPropagateOp(op)
		case block := <- pblock:
			MinerInstance.MSI.ServerHeartBeat()
			PeerPropagateBlock(block)
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
		empty := new(Empty)
		err := peer.Client.Call("Peer.Hb", &empty, &empty)
		if !CheckError(err, "PeerHeartBeats:"+addr){
			peer.LastHeartBeat = time.Now()
		}
	}
}

// Send a PropagateOp call to each peer
// Assumption: Nothing needs to be done on the miner itself, only send the op onwards
func PeerPropagateOp(op blockchain.Operation) {
	for addr, peer := range PeerList {
		empty := new(Empty)
		args := PropagateOpArgs{op, TTL}
		err := peer.Client.Call("Peer.PropagateOp", args, &empty)
		if !CheckError(err, "PeerPropagateOp:"+addr){
			peer.LastHeartBeat = time.Now()
		}
	}
}

// Send a PropagateBlock call to each peer
// Assumption: Nothing needs to be done on the miner itself, only send the block onwards
func PeerPropagateBlock(block blockchain.Block) {
	for addr, peer := range PeerList {
		empty := new(Empty)
		args := PropagateBlockArgs{block, TTL}
		err := peer.Client.Call("Peer.PropagateBlock", args, &empty)
		if !CheckError(err, "PeerPropagateBlock:"+addr){
			peer.LastHeartBeat = time.Now()
		}
	}
}
// Look through current active connections and delete them if they are not live
func CheckLiveliness() {
	interval := time.Duration(MinerInstance.Settings.HeartBeat) * time.Millisecond
	for addr, peer := range PeerList {
		if time.Since(peer.LastHeartBeat) > interval {
			fmt.Println("Stale connection: ", addr, " deleting")
			peer.Client.Close()
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
func CheckError(err error, parent string) bool {
	if err != nil {
		fmt.Println(parent, ":: found error! ", err)
		return true
	}
	return false
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
	// Extract key pairs
	ExtractKeyPairs(pubKey, privKey)
	// Listening Address
	ln, _ := net.Listen("tcp", ":0")
	addr := ln.Addr()
	MinerInstance.Addr = addr
	// 2. Setup Miner-Miner Listener
	go listenPeerRpc(ln, MinerInstance)

	// Connect to Server
	MinerInstance.ConnectToServer(serverIP)
	MinerInstance.MSI.Register(addr)

	pop := make(chan blockchain.Operation)
	pblock := make(chan blockchain.Block)
	// 3. Setup Miner Heartbeat Manager
	go ManageConnections(pop, pblock)

	// 4. Setup Problem Solving

	// 5. Setup Client-Miner Listener (this thread)
	OpenLibMinerConn(":0")
}
