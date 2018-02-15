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
	blockchain, _ := GetLongestPath(p.miner.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)
	*reply = blockchain
	fmt.Println("Connect called by: ", args.Addr.String())

	return nil
}

// This RPC is a no-op. It's used by the peer to ensure that this miner is still alive.
func (p *PeerRpc) Hb(args *Empty, reply *Empty) error {
	fmt.Println("Hb called")
	return nil
}

// Get a shape interface from an operation.
func (m Miner) getShapeFromOp(op blockchain.Operation) (shapelib.Shape, error) {
	pathlist, err := utils.GetParsedSVG(op.SVGString)
	if err == nil {
		// Error is nil, should be parsable into shapelib.Path
		return utils.SVGToPoints(pathlist, int(m.Settings.CanvasSettings.CanvasXMax),
			int(m.Settings.CanvasSettings.CanvasXMax), op.Fill != "transparent",
			op.Stroke != "transparent")
	}

	// TODO: try parsing it as a circle
	//circ, err := utils.GetParsedCirc(op.SVGOp)
	//if err != nil {
	//	fmt.Println("SVG string is neither circle nor path:", op.SVGOp)
	//	return shapelib.Shape(shapelib.Circle{0, 0, 0, false}), err
	//}

	// FIXME: change for circle
	circ := shapelib.NewCircle(0, 0, 0, false, false)
	return circ, fmt.Errorf("Not a path or circle")
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
	defer validateLock.Unlock()

	blocks, _ := GetLongestPath(p.miner.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)
	if args.OpInfo.Op.OpType == blockchain.ADD {
		err = p.miner.checkInkAndConflicts(subarr, inkRequired, args.OpInfo.PubKey, blocks, args.OpInfo.Op.SVGString)
	} else {
		err = p.miner.checkDeletion(args.OpInfo.OpSig, args.OpInfo.PubKey, blocks)
	}

	if err != nil {
		return err
	}

	// Update the solver. There will likely need to be additional logic somewhere here.
	p.opSCh <- args.OpInfo

	// Propagate op to list of connected peers.
	// TODO: figure out a way to optimize this... don't want to revalidate ops and stuff
	args.TTL--
	if args.TTL > 0 {
		p.opCh <- args
	}

	return nil
}

// This RPC is used to send a new block (addshape, deleteshape) to miners.
// Will not return any useful information.
func (p *PeerRpc) PropagateBlock(args PropagateBlockArgs, reply *Empty) error {
	fmt.Println("PropagateBlock called")

	// - Validate the block

	validateLock.Lock()
	defer validateLock.Unlock()

	longest, _ := GetLongestPath(p.miner.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)

	p.miner.ValidateBlock(args.Block, longest)

	// - Add block to block chain.
	InsertBlock(args.Block)

	length := len(longest)
	lastblock := longest[length-1]

	// Propagate block to list of connected peers.
	args.TTL--
	if args.TTL > 0 {
		p.blkCh <- args
	}

	newlongest, _ := GetLongestPath(p.miner.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)
	newlength := len(newlongest)
	newlastblock := newlongest[newlength-1]
	if newlength >= length && newlastblock.Nonce != lastblock.Nonce && newlastblock.MinerPubKey != lastblock.MinerPubKey {
		p.blkSCh <- args.Block
	}

	return nil
}

// This RPC is used for peers to get latest information when they are newly
// initalized. No useful argument.
func (p PeerRpc) GetBlockChain(args Empty, reply *GetBlockChainArgs) error {
	fmt.Println("GetBlockChain called")

	// Return a flattened version of the blockchain from somewhere

	return nil
}

// This will initialize the miner peer listener.
func listenPeerRpc(ln net.Listener, miner *Miner, opCh chan PropagateOpArgs,
	blkCh chan PropagateBlockArgs, opSCh chan blockchain.OperationInfo,
	blkSCh chan blockchain.Block, reqCh chan net.Addr) {
	pRpc := PeerRpc{miner, opCh, blkCh, opSCh, blkSCh, reqCh}

	fmt.Println("listenPeerRpc::listening on: ", ln.Addr().String())

	server := rpc.NewServer()
	server.RegisterName("Peer", &pRpc)

	server.Accept(ln)
}
