# BlockArtLib -> Miner API docs

------
Purpose: To agree on the calls and returns supported by our miner's rpc to make the art nodes work

## OpenCanvas

* Calls miner.OpenCanvas()
* gives private_key
* gets struct{settings, genesis block}

## CloseCanvas

* Calls miner.GetInk()
* gives private key
* gets struct{ink_remaining}

## AddShape

* Calls miner.Draw()
* gives struct{private_key, svg_string}
* gets struct{shapehash, blockhash, ink_remaining}

## DeleteShape

* calls miner.Delete()
* gives struct{private_key, shapehash}
* gets ?

## TODO: FINISH

