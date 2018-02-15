/*

Purpose of this file is to contain the validation functions needed for the add
and delete operations for shapes in the blockchain.

*/

package main

import (
	"fmt"

	"../blockchain"
	"../libminer"
	"../shapelib"
)

const LOG_VALIDATION = true


func (m Miner) ValidateBlock(block blockchain.Block, chain []blockchain.Block) bool {
	//fmt.Println("ValidateBlock::TODO: Unfinished")

	// check that the block hashes correctly
	// this is checked a lot though, do we need this? TODO
	if VerifyBlock(block){
		validatedops := ValidateOps(block.OpHistory, chain)
		if len(validatedops) == len(block.OpHistory) {
			return true
		}
	}

	return false
}

// Validates a set of operations against the longest block chain
func ValidateOps(ops []blockchain.OperationInfo, chain []blockchain.Block) ([]blockchain.OperationInfo) {
		testblock := new(blockchain.Block)
		testblock.MinerPubKey = ""
		for _, opinfo := range ops {
			testchain := append(chain, *testblock)
			op := opinfo.Op
			shape, err := MinerInstance.getShapeFromOp(op)
			if err != nil {
				continue
			}

			subarr, inkRequired := shape.SubArrayAndCost()
			if opinfo.Op.OpType == blockchain.ADD{
				err = MinerInstance.checkInkAndConflicts(subarr, inkRequired, opinfo.PubKey, testchain, op.SVGString)
			}	else {
				err = MinerInstance.checkDeletion(opinfo.OpSig, opinfo.PubKey, testchain)
			}
			if err != nil {
				continue
			}

			testblock.OpHistory = append(testblock.OpHistory, opinfo)
		}
		return testblock.OpHistory
}

// Checks if there are overlaps and enough ink
func ValidateOperation(op blockchain.Operation, pubKey string) error {
	shape, err := MinerInstance.getShapeFromOp(op)
	if err != nil {
		return err
	}

	subarr, inkRequired := shape.SubArrayAndCost()

	validateLock.Lock()
	defer validateLock.Unlock()

	blocks, _ := GetLongestPath(MinerInstance.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)
	err = MinerInstance.checkInkAndConflicts(subarr, inkRequired, pubKey, blocks, op.SVGString)

	if err != nil {
		return err
	}

	return nil
}

// Function used to determine if an add operation is allowed on the blockchain.
func (m Miner) checkInkAndConflicts(subarr shapelib.PixelSubArray, inkRequired int,
	pubkey string, blocks []blockchain.Block, svgString string) error {
	if LOG_VALIDATION {
		fmt.Println("checkInkAndConflicts called")
	}

	pubkeyInk := uint32(0)
	shapesExisting := make(map[string]*blockchain.OperationInfo)

	// Iterate over all blocks in this structure to form the pixel array
	// formed by all shapes not from this pubkey, and the ink remaining
	// of this pubkey.
	for i := 0; i < len(blocks); i++ {
		block := blocks[i]

		numOps := len(block.OpHistory)
		if block.MinerPubKey == pubkey {
			if numOps > 0 {
				pubkeyInk += m.Settings.InkPerOpBlock
			} else {
				pubkeyInk += m.Settings.InkPerNoOpBlock
			}
		}

		for j := 0; j < numOps; j++ {
			opInfo := block.OpHistory[j]
			op := opInfo.Op

			if opInfo.PubKey == pubkey {
				shape, err := m.getShapeFromOp(op)
				if err != nil {
					fmt.Println("CRITICAL ERROR: BAD SHAPE IN BLOCKCHAIN")
					continue
				}

				_, cost := shape.SubArrayAndCost()

				// Don't fill in the pixels for the same pubkey,
				// but compute the ink required in order to
				// check if pubkey has sufficient ink.
				// Don't bother validating that a DELETE has a
				// corresponding ADD. Assume all are valid.
				if op.OpType == blockchain.ADD {
					pubkeyInk -= uint32(cost)
				} else {
					pubkeyInk += uint32(cost)
				}
			} else {
				if op.OpType == blockchain.ADD {
					shapesExisting[opInfo.OpSig] = &opInfo
				} else {
					delete(shapesExisting, opInfo.OpSig)
				}
			}
		}
	}

	if inkRequired > int(pubkeyInk) {
		fmt.Println("checkInkAndConflicts: insufficient ink")
		return libminer.InsufficientInkError(uint32(inkRequired))
	}

	pixelarr := shapelib.NewPixelArray(int(m.Settings.CanvasSettings.CanvasXMax),
		int(m.Settings.CanvasSettings.CanvasYMax))

	// Merge all shapes existing into the pixel array for validating conflicts
	for _, v := range shapesExisting {
		shape, err := m.getShapeFromOp(v.Op)
		if err != nil {
			fmt.Println("CRITICAL ERROR: BAD SHAPE IN BLOCKCHAIN")
		}

		subarr, _ := shape.SubArrayAndCost()
		pixelarr.MergeSubArray(subarr)
	}

	if pixelarr.HasConflict(subarr) {
		fmt.Println("checkInkAndConflicts: conflict found")
		return libminer.ShapeOverlapError(svgString)
	}

	return nil
}

// Function used to determine if a delete operation is allowed on the blockchain.
func (m Miner) checkDeletion(sHash string, pubkey string, blocks []blockchain.Block) error {
	if LOG_VALIDATION {
		fmt.Println("checkInkAndConflicts called")
	}

	delAllowed := false

	// Iterate over all blocks in this structure to check if the shape has
	// actually been added. If shape is found to be added, cannot break out
	// of loop immediately - need to check if delete was already done also.
	// If a delete was done, can break out of loop and return an error.
	for i := 0; i < len(blocks); i++ {
		block := blocks[i]

		for j := 0; j < len(block.OpHistory); j++ {
			opInfo := block.OpHistory[j]

			if opInfo.PubKey == pubkey && opInfo.OpSig == sHash {
				if opInfo.Op.OpType == blockchain.ADD {
					delAllowed = true
				} else {
					delAllowed = false
					goto breakOuterLoop
				}
			}
		}
	}

breakOuterLoop:

	if !delAllowed {
		return libminer.ShapeOwnerError(sHash)
	}

	return nil
}
