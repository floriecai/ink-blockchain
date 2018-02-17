
package main

import "./miner"
import "os"

func main(){
	serverIP, pubKey, privKey := os.Args[1], os.Args[2], os.Args[3]
	miner.Mine(serverIP, pubKey, privKey)
}