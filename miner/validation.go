package main

import (
	"fmt"
	"../shapelib"
	"../blockchain"
)

const LOG_VALIDATION = true

// Function used to determine if an operation is allowed on the blockchain.
func (m Miner) checkInkAndConflicts(subarr shapelib.PixelSubArray, inkRequired int,
		pubkey string) error {
	if LOG_VALIDATION {
		fmt.Println("checkInkAndConflicts called")
	}

	// FIXME: this should be the miner's block data structure.
	// TODO: Need to figure out exactly what to check. There could be multiple
	// longest paths. It could be that there is a conflict on one and
	// not the other. Need to think about this one carefully.
	blocks := make([]blockchain.Block, 0)

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
			path, err := m.getPathFromOp(op)
			if err != nil {
				fmt.Println("CRITICAL ERROR, BAD OP IN BLOCKCHAIN");
				continue
			}

			subarr := path.SubArray()

			if block.OpHistory[j].PubKey != pubkey {
				pixelarr.MergeSubArray(subarr)
			} else {
				// Don't fill in the pixels for the same pubkey,
				// but compute the ink required in order to
				// check if pubkey has sufficient ink.
				var cost int

				if op.Fill == "transparent" {
					cost = path.LineCost()
				} else {
					cost = subarr.PixelsFilled()
				}

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
