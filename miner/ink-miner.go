package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
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

const (
	TRANSPARENT = "transparent"
)

// Our singleton miner instance
var MinerInstance *Miner

// Primitive representation of active art miners
var ArtNodeList map[int]bool = make(map[int]bool)

// List of peers WE connect TO, not peers that connect to US
var PeerList map[string]*Peer = make(map[string]*Peer)

const (
	// Global TTL of propagate requests
	TTL = 100
	// Maximum threads we will use for problem solving
	MAX_THREADS = 4
)

// Global block chain array
var BlockNodeArray []blockchain.BlockNode

// Global block chain search map
// Key: The hash of a block
// Val: The index of block with such hash in BlockNodeArray
var BlockHashMap map[string]int = make(map[string]int)

// Current Job ID
var CurrJobId int = 0

/*******************************
| Structs for the miners to use internally
| note: shared structs should be put in a different lib
********************************/
type Miner struct {
	PrivKey    *ecdsa.PrivateKey
	Addr       net.Addr
	Settings   minerserver.MinerNetSettings
	InkAmt     int
	LMI        *LibMinerInterface
	MSI        *MinerServerInterface
	BlockChain []blockchain.BlockNode
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
	Client        *rpc.Client
	LastHeartBeat time.Time
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
	}

	err = fmt.Errorf("invalid user")
	return err
}

func (lmi *LibMinerInterface) GetInk(req *libminer.Request, response *libminer.InkResponse) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		response.InkRemaining = uint32(MinerInstance.InkAmt)
		return nil
	}

	err = fmt.Errorf("invalid user")
	return err
}

func (lmi *LibMinerInterface) Draw(req *libminer.Request, response *libminer.DrawResponse) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		var drawReq libminer.DrawRequest
		json.Unmarshal(req.Msg, &drawReq)

		op := blockchain.Operation{
			"", // ShapeHash, FIXME
			"", // OpSig, FIXME
			blockchain.ADD,
			drawReq.SVGString,
			drawReq.Fill,
			drawReq.Stroke,
			""} // PubKey, FIXME

		shape, err := MinerInstance.getShapeFromOp(op)
		if err != nil {
			return err
		}

		subarr, inkRequired := shape.SubArrayAndCost()

		validateLock.Lock()
		defer validateLock.Unlock()

		blocks := GetLongestPath(MinerInstance.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)
		err = MinerInstance.checkInkAndConflicts(subarr, inkRequired, "", blocks) // FIXME: PubKey

		fmt.Println("drawing is currently unimplemented, sorry!")
		return nil
	}
	err = fmt.Errorf("invalid user")
	return err

}

func (lmi *LibMinerInterface) Delete(req *libminer.Request, response *libminer.InkResponse) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		var deleteReq libminer.DeleteRequest
		json.Unmarshal(req.Msg, &deleteReq)

		blockChain := GetLongestPath(MinerInstance.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)
		err = MinerInstance.checkDeletion(deleteReq.ShapeHash, "", blockChain) // FIXME: PubKey

		if err != nil {
			return libminer.ShapeOwnerError(deleteReq.ShapeHash)
		} else {
			// TODO:
			// send to solver
			// send to connection manager
		}
	}

	err = fmt.Errorf("invalid user")
	return err
}

func (lmi *LibMinerInterface) GetGenesisBlock(req *libminer.Request, response *string) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		*response = MinerInstance.Settings.GenesisBlockHash
		return nil
	}
	err = fmt.Errorf("invalid user")
	return err
}

func (lmi *LibMinerInterface) GetChildren(req *libminer.Request, response *libminer.BlocksResponse) (err error) {
	if Verify(req.Msg, req.HashedMsg, req.R, req.S, MinerInstance.PrivKey) {
		//children := GetBlockChildren(req.BlockHash)
		//response.Blocks = children
		return nil
	}
	err = fmt.Errorf("invalid user")
	return err
}

/*******************************
| Blockchain functions
********************************/

// Appends the new block to BlockArray and updates BlockHashMap
func InsertBlock(newBlock blockchain.Block) (err error) {
	hash := GetBlockHash(newBlock)
	if _, ok := BlockHashMap[hash]; !ok && VerifyBlock(newBlock) {
		// Create a new node for newBlock and append it to BlockNodeArray
		newBlockNode := blockchain.BlockNode{Block: newBlock, Children: []int{}}
		BlockNodeArray = append(BlockNodeArray, newBlockNode)

		// Create an entry for newBlock in BlockHashMap
		childIndex := len(BlockNodeArray) - 1
		BlockHashMap[hash] = childIndex

		// Update the entry for newBlock's parent in BlockNodeArray
		parentIndex := BlockHashMap[newBlock.PrevHash]
		parentBlockNode := &BlockNodeArray[parentIndex]
		parentBlockNode.Children = append(parentBlockNode.Children, childIndex)
		//fmt.Println("parent's node with new child:", parentBlockNode)
		return nil
	}
	err = fmt.Errorf("Block hash does not match up with block contents!")
	return err
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

func VerifyBlock(block blockchain.Block) bool {
	hash := GetBlockHash(block)
	if len(block.OpHistory) == 0 {
		return pow.Verify(hash, int(MinerInstance.Settings.PoWDifficultyNoOpBlock))
	} else {
		return pow.Verify(hash, int(MinerInstance.Settings.PoWDifficultyOpBlock))
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
	//fmt.Println("ServerHeartBeat::Sending heartbeat")
	err := msi.Client.Call("RServer.HeartBeat", MinerInstance.PrivKey.PublicKey, &ignored)
	if CheckError(err, "ServerHeartBeat") {
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
			if CheckError(err, "GetPeers:ResolvePeerAddr") {
				continue
			}

			PeerAddr, err := net.ResolveTCPAddr("tcp", addr.String())
			if CheckError(err, "GetPeers:ResolveLocalAddr") {
				continue
			}

			conn, err := net.DialTCP("tcp", LocalAddr, PeerAddr)
			if CheckError(err, "GetPeers:DialTCP") {
				continue
			}

			client := rpc.NewClient(conn)

			args := ConnectArgs{conn.LocalAddr().String()}
			err = client.Call("Peer.Connect", args, &empty)
			if CheckError(err, "GetPeers:Connect") {
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
func ManageConnections(pop chan PropagateOpArgs, pblock chan PropagateBlockArgs) {
	// Send heartbeats at three times the timeout interval to be safe
	interval := time.Duration(MinerInstance.Settings.HeartBeat / 5)
	heartbeat := time.Tick(interval * time.Millisecond)
	for {
		select {
		case <-heartbeat:
			MinerInstance.MSI.ServerHeartBeat()
			PeerHeartBeats()
		case op := <-pop:
			MinerInstance.MSI.ServerHeartBeat()
			PeerPropagateOp(op)
		case block := <-pblock:
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
		if !CheckError(err, "PeerHeartBeats:"+addr) {
			peer.LastHeartBeat = time.Now()
		}
	}
}

// Send a PropagateOp call to each peer
// Assumption: Nothing needs to be done on the miner itself, only send the op onwards
func PeerPropagateOp(op PropagateOpArgs) {
	for addr, peer := range PeerList {
		empty := new(Empty)
		args := PropagateOpArgs{op.Op, op.TTL}
		err := peer.Client.Call("Peer.PropagateOp", args, &empty)
		if !CheckError(err, "PeerPropagateOp:"+addr) {
			peer.LastHeartBeat = time.Now()
		}
	}
}

// Send a PropagateBlock call to each peer
// Assumption: Nothing needs to be done on the miner itself, only send the block onwards
func PeerPropagateBlock(block PropagateBlockArgs) {
	for addr, peer := range PeerList {
		empty := new(Empty)
		args := PropagateBlockArgs{block.Block, block.TTL}
		err := peer.Client.Call("Peer.PropagateBlock", args, &empty)
		if !CheckError(err, "PeerPropagateBlock:"+addr) {
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
| Crypto-Management
********************************/
// The problemsolver handles 4 main functions
// 1. Spins new workers for a new job
// 2. Kills old workers for a new job
// 3. Receive job updates via the given channels
// 4. TODO: Return solution

func ProblemSolver(sop chan blockchain.Operation, sblock chan blockchain.Block) {
	// Channel for receiving the final block w/ nonce from workers
	solved := make(chan blockchain.Block)

	// Channel returned by a job call that can kill the workers for that particular job
	var done chan bool

	for {
		select {
		case op := <-sop:
			// Received an op from somewhere
			// Assuming it is properly validated
			// Add it to the block we were working on
			// reissue job
			fmt.Println("got new op to hash:", op)
			// Kill current job
			close(done)
			close(solved)

			// Make a new channel
			solved = make(chan blockchain.Block)

			// TODO: setup a new OpJob with the given op

		case block := <-sblock:
			// Received a block from somewhere
			// Assume that this block was validated
			// Assume this is the next block to build off of
			// Reissue a job with this blockhash as prevBlock
			fmt.Println("got new block to hash:", block)

			// Kill current job
			close(done)
			close(solved)

			// Make a new channel
			solved = make(chan blockchain.Block)

			// Assume this was block was validated
			// Assume this block has already been inserted 
			done = NoopJob(GetBlockHash(block), solved)

		case sol := <-solved:
			fmt.Println("got a solution: ", sol)

			// Kill current job
			close(done)
			close(solved)
			// Make a new channel
			solved = make(chan blockchain.Block)

			// Insert block into our data structure
			// TODO: Do we insert it here or upstream via a channel?
			InsertBlock(sol)
			//fmt.Println("inserted solution: ", BlockNodeArray)
			// Start a job on the longest block in the chain
			blockchain := GetLongestPath(MinerInstance.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)
			//fmt.Println("state of the longest blockchain", blockchain)
			lastblock := blockchain[len(blockchain)-1]
			done = NoopJob(GetBlockHash(lastblock), solved)

		default:
			if CurrJobId == 0 {
				fmt.Println("Initiating the first job")
				done = NoopJob(MinerInstance.Settings.GenesisBlockHash, solved)
			}
			// Wait for current job to change
		}
	}
}

// Initiate a job with an empty op array and a blockhash
func NoopJob(hash string, solved chan blockchain.Block) chan bool {
	CurrJobId++
	block := blockchain.Block{PrevHash: hash,
		MinerPubKey: pubKeyToString(MinerInstance.PrivKey.PublicKey)}
	done := make(chan bool)
	for i := 0; i <= MAX_THREADS; i++ {
		CurrJobId++
		// Split up the start by the maximum number of threads we allow
		start := math.MaxUint32 / MAX_THREADS * i
		go pow.Solve(block, MinerInstance.Settings.PoWDifficultyNoOpBlock, uint32(start), solved, done)
	}
	return done
}

// Initiate the a job with a predefined op array
func OpJob(hash string, Ops []blockchain.Operation, solved chan blockchain.Block) chan bool {
	CurrJobId++
	block := blockchain.Block{PrevHash: hash,
		OpHistory:   Ops,
		MinerPubKey: pubKeyToString(MinerInstance.PrivKey.PublicKey)}
	done := make(chan bool)
	for i := 0; i <= MAX_THREADS; i++ {
		CurrJobId++
		// Split up the start by the maximum number of threads we allow
		start := math.MaxUint32 / MAX_THREADS * i
		go pow.Solve(block, MinerInstance.Settings.PoWDifficultyOpBlock, uint32(start), solved, done)
	}
	return done
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

// Returns an array of Blocks that are on the longest path
func GetLongestPath(initBlockHash string, blockHashMap map[string]int, blockNodeArray []blockchain.BlockNode) []blockchain.Block {
	blockChain := make([]blockchain.Block, 0)

	initBIndex := blockHashMap[initBlockHash]
	blockChain = append(blockChain, blockNodeArray[initBIndex].Block)

	if len(blockNodeArray[initBIndex].Children) == 0 {
		return blockChain
	}

	var longestPath []blockchain.Block
	maxLen := -1

	for _, childIndex := range blockNodeArray[initBIndex].Children {
		child := blockNodeArray[childIndex]

		childHash := GetBlockHash(child.Block)
		childPath := GetLongestPath(childHash, blockHashMap, blockNodeArray)
		if maxLen < len(childPath) {
			maxLen = len(childPath)
			longestPath = childPath
		}
	}

	blockChain = append(blockChain, longestPath...)
	return blockChain

}

func GetBlockHash(block blockchain.Block) string {
	h := md5.New()
	bytes, _ := json.Marshal(block)
	h.Write(bytes)
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
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

	// 2. Create communication channels between goroutines
	pop := make(chan PropagateOpArgs, 8)
	pblock := make(chan PropagateBlockArgs, 8)
	sop := make(chan blockchain.Operation, 8)
	sblock := make(chan blockchain.Block, 8)

	// 3. Setup Miner-Miner Listener
	go listenPeerRpc(ln, MinerInstance, pop, pblock, sop, sblock)

	// Connect to Server
	MinerInstance.ConnectToServer(serverIP)
	MinerInstance.MSI.Register(addr)

	//Initialize the hash map and the block node array with the genesis block
	BlockHashMap[MinerInstance.Settings.GenesisBlockHash] = 0
	BlockNodeArray = append(BlockNodeArray, blockchain.BlockNode{})

	// 4. Setup Miner Heartbeat Manager
	go ManageConnections(pop, pblock)

	// 5. Setup Problem Solving
	go ProblemSolver(sop, sblock)

	// 6. Setup Client-Miner Listener (this thread)
	OpenLibMinerConn(":0")
}
