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
	"../libminer"
)

/*******************
* TYPE_DEFINITIONS *
*******************/

// Struct for maintaining state of the PeerRpc
type PeerRpc struct {
	miner *Miner
	
	// Param blocksPublished is a map used as a set data structure. It
	// stores the blockhash as a string. Any received blockhash that is
	// found to be in this set is assumed to already have been published to
	// peers, and will not be published again. This is in order to avoid
	// broadcast loops.
	blocksPublished map[string]Empty
}

// Empty struct. Use for filling required but unused function parameters.
type Empty struct{}

type ConnectArgs struct {
	Peer_addr string
}

type PropagateOpArgs struct {
	op Empty // TODO: proper struct here
}

type PropagateBlockArgs struct {
	block libminer.Block
}

type GetBlockChainArgs struct {
	blockChain []libminer.Block
}

/***********************
* FUNCTION_DEFINITIONS *
***********************/

// Adds the connecting peer to the list of maintained peers. The peer
// requesting connect will be added to the maintained peer count. There will
// be a heartbeat procedure for it, and any data propagations will be sent to
// the peer as well.
func (p PeerRpc) Connect(args *ConnectArgs, reply *Empty) error {
	fmt.Println("Connect called by: ", args.Peer_addr)

	// - Add the peer miner to list of connected peers.
	// - Start a heartbeat for the new miner.


	return nil
}

// This RPC is a no-op. It's used by the peer to ensure that this miner is
// still alive.
func (p PeerRpc) Hb(args *Empty, reply *Empty) error {
	fmt.Println("Hb called")
	return nil
}

// This RPC is used to send an operation (addshape, deleteshape) to miners.
// Will not return any useful information.
func (p PeerRpc) PropagateOp(args *PropagateOpArgs, reply *Empty) error {
	fmt.Println("PropagateOp called")

	// - Validate the operation
	// - Update the solver.
	// - Propagate op to list of connected peers.

	return nil
}

// This RPC is used to send a new block (addshape, deleteshape) to miners.
// Will not return any useful information.
func (p PeerRpc) PropagateBlock(args *PropagateBlockArgs, reply *Empty) error {
	fmt.Println("PropagateBlock called")

	// - Validate the block
	// - Add block to block chain.
	// - Update the solver
	// - Propagate block to list of connected peers.

	return nil
}

// This RPC is used for peers to get latest information when they are newly
// initalized. No useful argument.
func (p PeerRpc) GetBlockChain(args *Empty, reply *GetBlockChainArgs) error {
	fmt.Println("GetBlockChain called")

	// Return a flattened version of the blockchain from somewhere

	return nil
}

// This will initialize the miner peer listener.
func listenPeerRpc(ln net.Listener, miner *Miner) {
	pRpc := PeerRpc{miner, make(map[string]Empty)}

	fmt.Println("listenPeerRpc::listening on: ", ln.Addr().String())

	server := rpc.NewServer()
	server.RegisterName("Peer", pRpc)

	server.Accept(ln)
}
