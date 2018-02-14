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


	/*******************
	TODO: Delay evaluation of pixel array
		  until the entire block chain is built
		  to account for deletes
	*******************/

// Function used to determine if an add operation is allowed on the blockchain.
func (m Miner) checkInkAndConflicts(subarr shapelib.PixelSubArray, inkRequired int,
	pubkey string, blocks []blockchain.Block, svgString string) error {
	if LOG_VALIDATION {
		fmt.Println("checkInkAndConflicts called")
	}

	// Pixel array for checking shape conflicts
	pixelarr := shapelib.NewPixelArray(int(m.Settings.CanvasSettings.CanvasXMax),
		int(m.Settings.CanvasSettings.CanvasYMax))

	pubkeyInk := uint32(0)

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
			path, err := m.getShapeFromOp(op)
			if err != nil {
				fmt.Println("CRITICAL ERROR, BAD OP IN BLOCKCHAIN")
				continue
			}

			subarr, cost := path.SubArrayAndCost()

			if opInfo.PubKey != pubkey {
				pixelarr.MergeSubArray(subarr)
			} else {
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
			}
		}
	}

	// TODO: check ink of the public key for any operations currently in
	// progress; a pubkey may have more than one op in progress of being
	// put into the blockchain at a given time.

	if inkRequired > int(pubkeyInk) {
		fmt.Println("checkInkAndConflicts: insufficient ink")
		return libminer.InsufficientInkError(uint32(inkRequired))
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
