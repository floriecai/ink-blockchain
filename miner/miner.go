package main

import (
	"net/rpc"

	"../libminer"
)

var MinerInstance *Miner

/*******************************
| Structs for the miners to use internally
| note: shared structs should be put in a different lib
********************************/
type Miner struct {
}

type Lib_Miner_Interface struct {
}

/*******************************
| Miner functions
********************************/
func (m *Miner) ConnectToServer() {

}

/*******************************
| RPC functions
********************************/

func (lmi *Lib_Miner_Interface) OpenCanvas(hello *libminer.RegisterRequest, response *libminer.RegisterResponse) (err error) {
	return nil
}

func (lmi *Lib_Miner_Interface) GetInk(hello *[]byte, response *libminer.InkResponse) (err error) {
	return nil
}

func (lmi *Lib_Miner_Interface) Draw(req *libminer.DrawRequest, response *libminer.DrawResponse) (err error) {
	return nil
}

func (lmi *Lib_Miner_Interface) Delete(req *libminer.DeleteRequest, response *libminer.InkResponse) (err error) {
	return nil
}

/* TODO
func (lmi *Lib_Miner_Interface) GetBlockChain(hello *[]byte, response *[]Block) (err error) {
	return nil
}
*/

func (lmi *Lib_Miner_Interface) GetGenesisBlock(hello *[]byte, response *string) (err error) {
	return nil
}

/*******************************
| Helpers
********************************/

/*******************************
| Main
********************************/
func main() {

	//Setup the singleton miner instance
	MinerInstance = &Miner{}
	MinerInstance.ConnectToServer()

	//Setup the node->miner rpc calls
	lib_miner_int := new(Lib_Miner_Interface)
	server := rpc.NewServer()
	server.Register(lib_miner_int)
}
