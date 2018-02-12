# Plan

Purpose: to lay out our objectives for completing the project

-----

## Miner

- [x] Miner Connection to Server
- [x] Miner completes crypto puzzles
- [x] Miner creates a block chain of noops starting from genesis block
- [ ] Miner sends the longest chain to other miners
- [x] Miner accepts rpc calls from art nodes
- [ ] Miner can perform art transactions
- [ ] Miner creates a block chain of noops and art transactions
- [ ] Miner validates art transactions with other miners

## Artlib

- [x] Artlib connects to miner
- [ ] OpenCanvas(minerAddress, private/public Keys) -> canvas, settings, err
- [ ] canvas.CloseCanvas() -> inkRemaining, err
- [ ] canvas.AddShape(validateNum, shapeType, shapeSvgString, fill, stroke) -> shapeHash, blockHash, inkRemaining, err 
- [ ] canvas.GetSvgString(shapeHash) -> svgString, err
- [ ] canvas.GetInk() -> inkRemaining
- [ ] canvas.DeleteShape(validateNum, shapeHash) -> inkRemaining, err
- [ ] canvas.GetShapes(blockHash) -> shapeHashes, err
- [ ] canvas.GetGenesisBlock() -> blockHash, err
- [ ] canvas.GetChildren(blockHash) -> blockHashes, err

## Misc

- [ ] Deploy on Azure
