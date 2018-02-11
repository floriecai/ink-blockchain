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
	"net"
	"net/rpc"
	"os"
	"../blockchain"
	"../shapelib"
	"../utils"
	"sync"
)

/*******************
* TYPE_DEFINITIONS *
*******************/

// Struct for maintaining state of the peerRpc
type peerRpc struct {
	miner *Miner

	// Param blocksPublished is a map used as a set data structure. It
	// stores the blockhash as a string. Any received blockhash that is
	// found to be in this set is assumed to already have been published to
	// peers, and will not be published again. This is in order to avoid
	// broadcast loops.
	blocksPublished map[string]empty
}

// Empty struct. Use for filling required but unused function parameters.
type empty struct{}

type connectArgs struct {
	peer_addr string
}

type propagateOpArgs struct {
	op blockchain.Operation
}

type propagateBlockArgs struct {
	block blockchain.Block
}

type getBlockChainArgs struct {
	blockChain []blockchain.Block
}

/***********************
* FUNCTION_DEFINITIONS *
***********************/

// Adds the connecting peer to the list of maintained peers. The peer
// requesting connect will be added to the maintained peer count. There will
// be a heartbeat procedure for it, and any data propagations will be sent to
// the peer as well.
func (p *peerRpc) Connect(args *connectArgs, reply *empty) error {
	fmt.Println("Connect called")

	// - Add the peer miner to list of connected peers.
	// - Start a heartbeat for the new miner.

	return nil
}

// This RPC is a no-op. It's used by the peer to ensure that this miner is still alive.
func (p *peerRpc) Hb(args *empty, reply *empty) error {
	fmt.Println("Hb called")
	return nil
}

// Get a shapelib.Path from an operation
func (m Miner)getPathFromOp(op blockchain.Operation) (shapelib.Path, error) {
	pathlist, err := utils.GetParsedSVG(op.SVGOp)
	if err != nil {
		fmt.Println("PropagateOp err:", err);
		path := shapelib.NewPath(nil, false)
		return path, err
	}

	// Get the shapelib.Path representation for this svg path
	path := utils.SVGToPoints(pathlist, int(m.Settings.CanvasSettings.CanvasXMax),
		int(m.Settings.CanvasSettings.CanvasXMax), op.Fill != "transparent")

	return path[0], nil
}

// This lock is intended to be used so that only one operation will be in the
// validation procedure at any given point. This is to prevent race conditions
// of multiple, conflicting operations.
var validateOpLock sync.Mutex

// This RPC is used to send an operation (addshape, deleteshape) to miners.
// Will not return any useful information.
func (p *peerRpc) PropagateOp(args *propagateOpArgs, reply *empty) error {
	fmt.Println("PropagateOp called")

	// Get the shapelib.Path representation for this svg path
	path, err := p.miner.getPathFromOp(args.op)
	if err != nil {
		return err
	}

	// Get the subarr for checking conflicting shapes, as well as for computing area
	// in the case that fill is not transparent.
	subarr := path.SubArray()

	var inkRequired int
	if args.op.Fill != "transparent" {
		inkRequired = path.TotalLength()
	} else {
		inkRequired = subarr.PixelsFilled()
	}

	validateOpLock.Lock()
	err = p.miner.checkInkAndConflicts(subarr, inkRequired, args.op.PubKey)
	validateOpLock.Unlock()

	if err != nil {
		return err
	}

	// - Update the solver.
	// - Propagate op to list of connected peers.

	return nil
}

// This RPC is used to send a new block (addshape, deleteshape) to miners.
// Will not return any useful information.
func (p *peerRpc) PropagateBlock(args *propagateBlockArgs, reply *empty) error {
	fmt.Println("PropagateBlock called")

	// - Validate the block
	// - Add block to block chain.
	// - Update the solver
	// - Propagate block to list of connected peers.

	return nil
}

// This RPC is used for peers to get latest information when they are newly
// initalized. No useful argument.
func (p *peerRpc) GetBlockChain(args *empty, reply *getBlockChainArgs) error {
	fmt.Println("GetBlockChain called")

	// Return a flattened version of the blockchain from somewhere

	return nil
}

// This will initialize the miner peer listener.
func listenPeerRpc(addr string, miner *Miner) {
	pRpc := peerRpc{miner, make(map[string]empty)}

	conn, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Peer RPC could not initialize tcp listener")
		os.Exit(1)
	}

	server := rpc.NewServer()
	server.RegisterName("Peer", pRpc)

	server.Accept(conn)
}
