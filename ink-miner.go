
package main

import "./miner"
import "io/ioutil"
import "os"
import "strings"

func main(){
	serverIP := os.Args[1]
	
	// Grab pubKey and privKey from key-pairs.txt
	keyBytes, _ := ioutil.ReadFile("./key-pairs.txt")
	keyString := string(keyBytes[:])
	privKey := strings.Split(keyString, "\n")[0]
	pubKey := strings.Split(keyString, "\n")[1]

	miner.Mine(serverIP, pubKey, privKey)
}
