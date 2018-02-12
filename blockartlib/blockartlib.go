/*

This package specifies the application's interface to the the BlockArt
library (blockartlib) to be used in project 1 of UBC CS 416 2017W2.

*/

package blockartlib

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"
	"os"

	"../libminer"
	"../utils"
)

const (
	TRANSPARENT = "transparent"
)

// Represents a type of shape in the BlockArt system.
type ShapeType int

const (
	// Path shape.
	PATH ShapeType = iota

	// Circle shape (extra credit).
	// CIRCLE
)

// Settings for a canvas in BlockArt.
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32
	CanvasYMax uint32
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string

	// The minimum number of ink miners that an ink miner should be
	// connected to. If the ink miner dips below this number, then
	// they have to retrieve more nodes from the server using
	// GetNodes().
	MinNumMinerConnections uint8

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32
	InkPerNoOpBlock uint32

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8
	PoWDifficultyNoOpBlock uint8

	// Canvas settings
	CanvasSettings CanvasSettings
}

////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// These type definitions allow the application to explicitly check
// for the kind of error that occurred. Each API call below lists the
// errors that it is allowed to raise.
//
// Also see:
// https://blog.golang.org/error-handling-and-go
// https://blog.golang.org/errors-are-values

// Contains address IP:port that art node cannot connect to.
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("BlockArt: cannot connect to [%s]", string(e))
}

// Contains amount of ink remaining.
type InsufficientInkError uint32

func (e InsufficientInkError) Error() string {
	return fmt.Sprintf("BlockArt: Not enough ink to addShape [%d]", uint32(e))
}

// Contains the offending svg string.
type InvalidShapeSvgStringError string

func (e InvalidShapeSvgStringError) Error() string {
	return fmt.Sprintf("BlockArt: Bad shape svg string [%s]", string(e))
}

// Contains the offending svg string.
type ShapeSvgStringTooLongError string

func (e ShapeSvgStringTooLongError) Error() string {
	return fmt.Sprintf("BlockArt: Shape svg string too long [%s]", string(e))
}

// Contains the bad shape hash string.
type InvalidShapeHashError string

func (e InvalidShapeHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid shape hash [%s]", string(e))
}

// Contains the bad shape hash string.
type ShapeOwnerError string

func (e ShapeOwnerError) Error() string {
	return fmt.Sprintf("BlockArt: Shape owned by someone else [%s]", string(e))
}

// Empty
type OutOfBoundsError struct{}

func (e OutOfBoundsError) Error() string {
	return fmt.Sprintf("BlockArt: Shape is outside the bounds of the canvas")
}

// Contains the hash of the shape that this shape overlaps with.
type ShapeOverlapError string

func (e ShapeOverlapError) Error() string {
	return fmt.Sprintf("BlockArt: Shape overlaps with a previously added shape [%s]", string(e))
}

// Contains the invalid block hash.
type InvalidBlockHashError string

func (e InvalidBlockHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid block hash [%s]", string(e))
}

// </ERROR DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////////////////

// Represents a canvas in the system.
type Canvas interface {
	// Adds a new shape to the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - InsufficientInkError
	// - InvalidShapeSvgStringError
	// - ShapeSvgStringTooLongError
	// - ShapeOverlapError
	// - OutOfBoundsError
	AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error)
	// aDD SHAPE blocks until number of blocks (validateNum) follow current block

	// Returns the encoding of the shape as an svg string.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidShapeHashError
	GetSvgString(shapeHash string) (svgString string, err error)

	// Returns the amount of ink currently available.
	// Can return the following errors:
	// - DisconnectedError
	GetInk() (inkRemaining uint32, err error)

	// Removes a shape from the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - ShapeOwnerError
	DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error)

	// Retrieves hashes contained by a specific block.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetShapes(blockHash string) (shapeHashes []string, err error)

	// Returns the block hash of the genesis block.
	// Can return the following errors:
	// - DisconnectedError
	GetGenesisBlock() (blockHash string, err error)

	// Retrieves the children blocks of the block identified by blockHash.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetChildren(blockHash string) (blockHashes []string, err error)

	// Closes the canvas/connection to the BlockArt network.
	// - DisconnectedError
	CloseCanvas() (inkRemaining uint32, err error)
}

type CanvasT struct {
	Id       int
	Settings CanvasSettings
	Miner    *rpc.Client
	PrivKey  ecdsa.PrivateKey
}

// Adds a new shape to the canvas.
// Can return the following errors:
// - DisconnectedError
// - InsufficientInkError
// - InvalidShapeSvgStringError
// - ShapeSvgStringTooLongError
// - ShapeOverlapError
// - OutOfBoundsError
func (canvas CanvasT) AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error) {
	drawRequest := libminer.DrawRequest{
		Id:          canvas.Id,
		ValidateNum: validateNum,
		SVGString:   shapeSvgString,
		Fill:        fill,
		Stroke:      stroke}
	msg, _ := json.Marshal(drawRequest)
	req := getRPCRequest(msg, &canvas.PrivKey)

	var reply libminer.DrawResponse

	err = canvas.Miner.Call("LibMinerInterface.Draw", &req, &reply)

	if err != nil {
		fmt.Println("Error on calling Miner.Draw")
		return "", "", 0, err
	}

	return reply.ShapeHash, reply.BlockHash, reply.InkRemaining, err
}

// aDD SHAPE blocks until number of blocks (validateNum) follow current block
// Returns the encoding of the shape as an svg string.
// Can return the following errors:
// - DisconnectedError
// - InvalidShapeHashError
func (canvas CanvasT) GetSvgString(shapeHash string) (svgString string, err error) {
	if canvas.Miner == nil {
		return "", DisconnectedError(canvas.Id)
	}

	msg, _ := json.Marshal(libminer.GenericRequest{Id: canvas.Id})
	req := getRPCRequest(msg, &canvas.PrivKey)
	var resp libminer.BlocksResponse

	err = canvas.Miner.Call("LibMinerInterface.GetBlockChain", &req, &resp)

	if err != nil {
		checkError(err)
		return "", err
	}

	// TODO fcai
	// for i, block := range resp.Blocks {
	// 	for j, ops := range block.Ops {
	// 		if ops.Hash == shapeHash {
	// 			return ops.svgString, nil
	// 		}
	// 	}
	// }

	return "", InvalidShapeHashError(shapeHash)
}

// Returns the amount of ink currently available.
// Can return the following errors:
// - DisconnectedError
func (canvas CanvasT) GetInk() (inkRemaining uint32, err error) {
	if canvas.Miner == nil {
		return 0, DisconnectedError(canvas.Id)
	}

	msg, _ := json.Marshal(libminer.GenericRequest{Id: canvas.Id})
	req := getRPCRequest(msg, &canvas.PrivKey)
	var resp libminer.InkResponse

	err = canvas.Miner.Call("LibMinerInterface.GetInk", &req, &resp)

	if err != nil {
		checkError(err)
		return 0, err
	}

	return resp.InkRemaining, err
}

// Removes a shape from the canvas.
// Can return the following errors:
// - DisconnectedError
// - ShapeOwnerError
func (canvas CanvasT) DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error) {
	if canvas.Miner == nil {
		return 0, DisconnectedError(string(canvas.Id))
	}

	deleteArgs := libminer.DeleteRequest{Id: canvas.Id, ShapeHash: shapeHash, ValidateNum: validateNum}
	msg, _ := json.Marshal(deleteArgs)
	req := getRPCRequest(msg, &canvas.PrivKey)

	var resp libminer.InkResponse

	err = canvas.Miner.Call("LibMinerInterface.Delete", &req, &resp)

	if err != nil {
		log.Println("Error in Miner.Delete")
		checkError(err)
		return 0, err
	}

	return resp.InkRemaining, err
}

// Retrieves hashes contained by a specific block.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (canvas CanvasT) GetShapes(blockHash string) (shapeHashes []string, err error) {
	if canvas.Miner == nil {
		return shapeHashes, DisconnectedError(string(canvas.Id))
	}

	msg, _ := json.Marshal(libminer.GenericRequest{Id: canvas.Id})
	req := getRPCRequest(msg, &canvas.PrivKey)
	var resp libminer.BlocksResponse

	err = canvas.Miner.Call("LibMinerInterface.GetBlockChain", &req, &resp)

	if err != nil {
		log.Println("Error in Miner.GetBlockChain in GetShapes")
		checkError(err)
		return shapeHashes, err
	}

	// TODO fcai
	// for i, block := resp.Blocks {
	// 	if block.Hash == blockHash {
	// 		return block.Ops, nil
	// 	}
	// }

	return shapeHashes, InvalidBlockHashError(blockHash)
}

// Returns the block hash of the genesis block.
// Can return the following errors:
// - DisconnectedError
func (canvas CanvasT) GetGenesisBlock() (blockHash string, err error) {
	if canvas.Miner == nil {
		return "", DisconnectedError(string(canvas.Id))
	}

	msg, _ := json.Marshal(libminer.GenericRequest{Id: canvas.Id})
	req := getRPCRequest(msg, &canvas.PrivKey)

	var resp libminer.BlocksResponse

	err = canvas.Miner.Call("LibMinerInterface.GetGenesisBlock", &req, &resp)
	if err != nil {
		checkError(err)
		return "", err
	}

	// TODO fcai, get blockhash from resp
	// return resp.Block[0].Hash, nil
	return "", err
}

// Retrieves the children blocks of the block identified by blockHash.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (canvas CanvasT) GetChildren(blockHash string) (blockHashes []string, err error) {
	if canvas.Miner == nil {
		return blockHashes, DisconnectedError(string(canvas.Id))
	}

	msg, _ := json.Marshal(libminer.GenericRequest{Id: canvas.Id})
	req := getRPCRequest(msg, &canvas.PrivKey)

	var resp libminer.BlocksResponse

	err = canvas.Miner.Call("LibMinerInterface.GetBlockChain", &req, &resp)

	// TODO fcai - get the children of the blockchain
	// for i, block := range resp.Blocks {
	// 	if block.Hash == blockHash {
	// 		// TODO get the children
	// 	}
	// }

	return blockHashes, err
}

// Closes the canvas/connection to the BlockArt network.
// - DisconnectedError
func (canvas CanvasT) CloseCanvas() (inkRemaining uint32, err error) {
	if canvas.Miner == nil {
		return 0, DisconnectedError(string(canvas.Id))
	}

	msg, _ := json.Marshal(libminer.GenericRequest{Id: canvas.Id})
	req := getRPCRequest(msg, &canvas.PrivKey)
	var resp libminer.InkResponse

	canvas.Miner.Call("LibMinerInterface.GetInk", &req, &resp)
	return inkRemaining, err
}

// The constructor for a new Canvas object instance. Takes the miner's
// IP:port address string and a public-private key pair (ecdsa private
// key type contains the public key). Returns a Canvas instance that
// can be used for all future interactions with blockartlib.
//
// The returned Canvas instance is a singleton: an application is
// expected to interact with just one Canvas instance at a time.
//
// Can return the following errors:
// - DisconnectedError
func OpenCanvas(minerAddr string, privKey ecdsa.PrivateKey) (canvas Canvas, setting CanvasSettings, err error) {
	var canvasT CanvasT
	miner, err := rpc.Dial("tcp", minerAddr)

	if err != nil {
		return canvasT, CanvasSettings{}, DisconnectedError(minerAddr)
	}

	msg := []byte("Hi")
	req := getRPCRequest(msg, &privKey)
	var resp libminer.RegisterResponse

	err = miner.Call("LibMinerInterface.OpenCanvas", &req, &resp)

	if err != nil {
		return canvas, setting, err
	}

	canvasT = CanvasT{
		Miner:    miner,
		Id:       resp.Id,
		PrivKey:  privKey,
		Settings: CanvasSettings{CanvasXMax: resp.CanvasXMax, CanvasYMax: resp.CanvasYMax}}

	return canvasT, canvasT.Settings, nil
}

func checkError(err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}

func getRPCRequest(msg []byte, privKey *ecdsa.PrivateKey) libminer.Request {
	hashedMsg := utils.ComputeHash(msg)
	r, s, _ := ecdsa.Sign(rand.Reader, privKey, hashedMsg)
	req := libminer.Request{R: *r, S: *s, HashedMsg: hashedMsg, Msg: msg}

	return req
}
