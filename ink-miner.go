
package main

import "./miner"
import "ioutil"
import "os"
import "strings"

func main(){
	serverIP := os.Args[1]
	
	// Grab pubKey and privKey from key-pairs.txt
	keyBytes, err := ioutil.ReadFile("./key-pairs.txt")
	_ = CheckError(err, "main:ioutil.ReadFile")
	keyString := string(keyBytes[:])
	privKey = strings.Split(keyString, "\n")[0]
	pubKey = strings.Split(keyString, "\n")[1]

	miner.Mine(serverIP, pubKey, privKey)
}