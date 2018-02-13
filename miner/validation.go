package main

import (
	"fmt"
	"../shapelib"
	"../blockchain"
)

const LOG_VALIDATION = true

// Function used to determine if an add operation is allowed on the blockchain.
func (m Miner) checkInkAndConflicts(subarr shapelib.PixelSubArray, inkRequired int,
		pubkey string) error {
	if LOG_VALIDATION {
		fmt.Println("checkInkAndConflicts called")
	}

	// TODO: Need to figure out exactly what to check. There could be multiple
	// longest paths. It could be that there is a conflict on one and
	// not the other. Need to think about this one carefully.
	blocks := GetLongestPath(m.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)

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
			op := block.OpHistory[j]
			path, err := m.getShapeFromOp(op)
			if err != nil {
				fmt.Println("CRITICAL ERROR, BAD OP IN BLOCKCHAIN");
				continue
			}

			subarr, cost := path.SubArrayAndCost()

			if op.PubKey != pubkey {
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
		return fmt.Errorf("insufficient ink")
	}

	if pixelarr.HasConflict(subarr) {
		fmt.Println("checkInkAndConflicts: conflict found")
		return fmt.Errorf("conflict found")
	}

	return nil
}

// Function used to determine if a delete operation is allowed on the blockchain.
func (m Miner) checkDeletion(sHash string, pubkey string) error {
	if LOG_VALIDATION {
		fmt.Println("checkInkAndConflicts called")
	}

	// TODO: Need to figure out exactly what to check. There could be multiple
	// longest paths. It could be that there is a deletion allowed on one but
	// not the other. Need to think about this one carefully.
	blocks := GetLongestPath(m.Settings.GenesisBlockHash, BlockHashMap, BlockNodeArray)

	delAllowed := false

	// Iterate over all blocks in this structure to check if the shape has
	// actually been added. If shape is found to be added, cannot break out
	// of loop immediately - need to check if delete was already done also.
	// If a delete was done, can break out of loop and return an error.
	for i := 0; i < len(blocks); i++ {
		block := blocks[i]

		for j := 0; j < len(block.OpHistory); j++ {
			op := block.OpHistory[j]

			if op.PubKey == pubkey && op.ShapeHash == sHash {
				if op.OpType == blockchain.ADD {
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
		return fmt.Errorf("Delete operation not allowed")
	}

	return nil
}
