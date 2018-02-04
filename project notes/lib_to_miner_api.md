# BlockArtLib -> Miner API docs

------
Purpose: To agree on the calls and returns supported by our miner's rpc to make the art nodes work

## OpenCanvas

* Calls miner.OpenCanvas()
* gives []byte(“hi”) -> ecdsa.Sign(miner’s_priv_key)
* (on the miner: he will verify with his public key, generate an ID to use for every future rpc)
* gets struct{generated_ID, settings} //do we need genesis block? don't think so

## CloseCanvas

* Calls miner.GetInk()
* gives []byte(id), signed with private key
* gets struct{ink_remaining}

## AddShape

* Calls miner.Draw()
* gives struct{id, svg_string}, signed with private key
* gets struct{shapehash, blockhash, ink_remaining}

## DeleteShape

* calls miner.Delete()
* gives struct{id, private_key, shapehash}, signed with private key
* gets struct{ink_remaining}

## GetInk

* calls miner.GetInk()
* gives []byte(id), signed with private key
* gets struct{ink_remaining}


## GetShapes

* calls miner.GetBlockChain()
* gives []byte(id), signed with private key
* gets array of blocks

## GetChildren

* calls miner.GetBlockChain()
* gives []byte(id), signed with private key
* gets array of blocks

## GetGenesisBlock

* calls miner.GetGenesisBlock()
* gives []byte(id), signed with private key
* gets hash_of_genesis_block string

