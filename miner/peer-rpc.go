/*

This file contains the following:
1. RPC definitions for miner peer to peer communication
2. Structs for request and reply for the above
3. Function to initialize the miner peer listener

Peer RPC calls:
  Connect(args *connectArgs, reply *empty)
  Hb(args *empty, reply *empty)
  PropagateOp(args *propagateOpArgs, reply *empty)
  PropagateBlock(args *propagateBlockArgs, reply *empty)
  GetBlockChain(args *empty, reply *getBlockChainArgs)

*/

package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"sync"

	"../blockchain"
	"../shapelib"
	"../utils"
)

/*******************
* TYPE_DEFINITIONS *
*******************/

// Struct for maintaining state of the PeerRpc
type PeerRpc struct {
	miner  *Miner
	opCh   chan PropagateOpArgs
	blkCh  chan PropagateBlockArgs
	opSCh  chan blockchain.OperationInfo
	blkSCh chan blockchain.Block
	reqCh  chan net.Addr
	blks   map[string]Empty
}

// Empty struct. Use for filling required but unused function parameters.
type Empty struct{}

type ConnectArgs struct {
	Addr net.Addr
}

type PropagateOpArgs struct {
	OpInfo blockchain.OperationInfo
	TTL    int
}

type PropagateBlockArgs struct {
	Block blockchain.Block
	TTL   int
}

type GetBlockChainArgs struct {
	blockChain []blockchain.Block
}

/***********************
* FUNCTION_DEFINITIONS *
***********************/

// Adds the connecting peer to the list of maintained peers. The peer
// requesting connect will be added to the maintained peer count. There will
// be a heartbeat procedure for it, and any data propagations will be sent to
// the peer as well.
func (p *PeerRpc) Connect(args ConnectArgs, reply *[]blockchain.Block) error {

	// - Send through request channel to Connection Manager to connect next time
	p.reqCh <- args.Addr
	blockchain := make([]blockchain.Block, 0)
	for i, node := range BlockNodeArray {
		if i != 0 {
			blockchain = append(blockchain, node.Block)
		}
	}
	*reply = blockchain
	fmt.Println("Connect called by: ", args.Addr.String())

	return nil
}

// This RPC is a no-op. It's used by the peer to ensure that this miner is still alive.
func (p *PeerRpc) Hb(args *Empty, reply *Empty) error {
	//fmt.Println("Hb called")
	return nil
}

// Get a shape interface from an operation.
func (m Miner) getShapeFromOp(op blockchain.Operation) (shapelib.Shape, error) {
	pathlist, err := utils.GetParsedSVG(op.SVGString)
	if err == nil {
		// Error is nil, should be parsable into shapelib.Path
		return utils.SVGToPoints(pathlist,
			int(m.Settings.CanvasSettings.CanvasXMax),
			int(m.Settings.CanvasSettings.CanvasXMax),
			op.Fill != "transparent",
			op.Stroke != "transparent")
	}

	// Try parsing it as a circle
	circ, err := utils.GetParsedCirc(op,
		int(m.Settings.CanvasSettings.CanvasXMax),
		int(m.Settings.CanvasSettings.CanvasXMax))
	if err != nil {
		fmt.Println("SVG string is neither circle nor path:", op.SVGString)
		return circ, err
	}

	// FIXME: change for circle
	return circ, nil
}

// Get a shapelib.Path from an operation
func (m Miner) getPathFromOp(op blockchain.Operation) (shapelib.Path, error) {
	pathlist, err := utils.GetParsedSVG(op.SVGString)
	if err != nil {
		fmt.Println("PropagateOp err:", err)
		path := shapelib.NewPath(nil, false, false)
		return path, err
	}

	// Get the shapelib.Path representation for this svg path
	return utils.SVGToPoints(pathlist, int(m.Settings.CanvasSettings.CanvasXMax),
		int(m.Settings.CanvasSettings.CanvasXMax), op.Fill != "transparent",
		op.Stroke != "transparent")
}

// This lock is intended to be used so that only one op or block will be in the
// validation procedure at any given point. This is to prevent race conditions
// of multiple, conflicting operations.
var validateLock sync.Mutex

// This RPC is used to send an operation (addshape, deleteshape) to miners.
// Will not return any useful information.
func (p *PeerRpc) PropagateOp(args PropagateOpArgs, reply *Empty) error {
	fmt.Println("PropagateOp called")

	// TODO: Validate the shapehash using the public key

	// Get the shapelib.Shape representation for this svg
	shape, err := p.miner.getShapeFromOp(args.OpInfo.Op)
	if err != nil {
		return err
	}

	subarr, inkRequired := shape.SubArrayAndCost()

	validateLock.Lock()
	log.Println("Got ValidateLock in Pop")
	defer log.Println("Release ValidateLock in Pop")
	defer validateLock.Unlock()

	blocks, _ := GetLongestPath(p.miner.Settings.GenesisBlockHash)
	if args.OpInfo.Op.OpType == blockchain.ADD {
		err = p.miner.checkInkAndConflicts(subarr, inkRequired, args.OpInfo.PubKey, blocks, args.OpInfo.Op.SVGString, args.OpInfo.OpSig)
	} else {
		fmt.Println("Checking deletion")
		err = p.miner.checkDeletion(args.OpInfo.AddSig, args.OpInfo.PubKey, blocks)
		if err != nil {
			fmt.Println("DELETE WAS BAD!!!")
		}
	}

	if err != nil {
		return err
	}

	// Update the solver. There will likely need to be additional logic somewhere here.
	p.opSCh <- args.OpInfo

	// Propagate op to list of connected peers.
	args.TTL--
	if args.TTL > 0 {
		p.opCh <- args
	}

	return nil
}

var msgLock sync.Mutex

// This RPC is used to send a new block (addshape, deleteshape) to miners.
// Will not return any useful information.
func (p *PeerRpc) PropagateBlock(args PropagateBlockArgs, reply *Empty) error {
	//fmt.Println("PropagateBlock called")
	msgLock.Lock()
	blkHash := GetBlockHash(args.Block)
	if _, exists := p.blks[blkHash]; exists {
		//fmt.Println("Ignoring already received blockhash")
		msgLock.Unlock()
		return nil
	} else {
		p.blks[blkHash] = Empty{}
		msgLock.Unlock()
	}

	validateLock.Lock()
	defer validateLock.Unlock()

	// Find the path that the block should be on, no guarantee it is the longest
	path := GetPath(args.Block.PrevHash)

	// Validate the block, if the block is not valid just drop it
	if p.miner.ValidateBlock(args.Block, path) {
		// Propagate block to list of connected peers. Too lazy to get rid of TTL;
		// it's not used any more for PropgateBlock though.
		if args.TTL > 0 {
			//fmt.Println("Propgation:", args.TTL)
			p.blkCh <- args
		}

		// Snapshot the current longest path
		longest, length := GetLongestPath(p.miner.Settings.GenesisBlockHash)
		lastblock := longest[length-1]

		// - Add block to block chain.
		InsertBlock(args.Block)

		// Check if the longest path changed
		newlongest, newlength := GetLongestPath(p.miner.Settings.GenesisBlockHash)
		newlastblock := newlongest[newlength-1]

		// If the longest path changed we should build off of it so send it to problem solver
		if newlength >= length && newlastblock.Nonce != lastblock.Nonce && newlastblock.MinerPubKey != lastblock.MinerPubKey {
			p.blkSCh <- args.Block
		}
	}

	return nil
}

// This RPC is used for peers to get latest information when they are newly
// initalized. No useful argument.
func (p *PeerRpc) GetBlockChain(args Empty, reply *GetBlockChainArgs) error {
	fmt.Println("GetBlockChain called")

	blockchain := make([]blockchain.Block, 0)
	for i, node := range BlockNodeArray {
		if i != 0 {
			blockchain = append(blockchain, node.Block)
		}
	}
	*reply = GetBlockChainArgs{blockchain}

	return nil
}

// This will initialize the miner peer listener.
func listenPeerRpc(ln net.Listener, miner *Miner, opCh chan PropagateOpArgs,
	blkCh chan PropagateBlockArgs, opSCh chan blockchain.OperationInfo,
	blkSCh chan blockchain.Block, reqCh chan net.Addr) {
	pRpc := PeerRpc{miner, opCh, blkCh, opSCh, blkSCh, reqCh, make(map[string]Empty)}

	fmt.Println("listenPeerRpc::listening on: ", ln.Addr().String())

	server := rpc.NewServer()
	server.RegisterName("Peer", &pRpc)

	server.Accept(ln)
}
