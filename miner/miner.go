package main

import (
	"net/rpc"
	"../lib_miner_structs"
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
func (m *Miner) ConnectToServer(){

}

/*******************************
| RPC functions
********************************/

func (lmi *Lib_Miner_Interface) OpenCanvas(hello *[]byte, response *lib_miner_structs.RegisterResponse) (err error){
	return nil
}

func (lmi *Lib_Miner_Interface) GetInk(hello *[]byte, response *lib_miner_structs.InkResponse)(err error) {
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