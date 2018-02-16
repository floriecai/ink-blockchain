/*

A trivial application to illustrate how the blockartlib library can be
used from an application in project 1 for UBC CS 416 2017W2.

Usage:
go run art-app.go
*/

package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import "./blockartlib"

import (
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"
	"./utils"
)

func main() {
	minerAddr := "[::]:63344"
	privKeyString := "3081a402010104304514498ab89a1680021e77227f8efd7dd55dd3ed98b2474593fc54f65b1f333aa7ac96a6ed3b087a182cbb572911a423a00706052b81040022a16403620004ab05604657137b476df7769af5e968e3124fe6ca1dc95d7caf68d1c6e8b21bde8536d687b121548f330fbc1dea01616cd62d973a4ddde46eed4e4cefa172d749d614e19d00608081b240343b7c3e8b6576bbb09ab12cc51fb8f6eed9ed8fd0c0"
	privateKeyBytes, _ := hex.DecodeString(privKeyString)
	privKey, _ := x509.ParseECPrivateKey(privateKeyBytes)
	// TODO: use crypto/ecdsa to read pub/priv keys from a file argument.

	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(minerAddr, *privKey)
	checkError(err)

	fmt.Println(canvas)
	fmt.Println(settings)

	validateNum := uint8(2)

	// Add a line.
	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 0 5", "transparent", "red")
	checkError(err)

	fmt.Println("%s, %s, %d", shapeHash, blockHash, ink)
	// Add another line.
	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 5 0", "transparent", "blue")
	checkError(err)

	fmt.Println("%s, %s, %d", shapeHash2, blockHash2, ink2)
	// Delete the first line.
	ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	checkError(err)
	fmt.Println("%d", ink3)
	// assert ink3 > ink2

	// Close the canvas.
	ink4, err := canvas.CloseCanvas()
	checkError(err)

	fmt.Println("%d", ink4)

	generateHTML(canvas)
}

// If error is non-nil, print it out.
func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error ", err.Error())
	}
}

// Recursively get the longest blockchain
func getLongestBlockchain(currBlockHash string, canvas blockartlib.Canvas) []string {
	// Add current block hash to longest chain
	longestBlockchain := []string{}
	longestBlockchain = append(longestBlockchain, currBlockHash)

	// Iterate through children of current block if any exist,
	// Adding the longest of them all to the longest blockchain
	children, err := canvas.GetChildren(currBlockHash)
	checkError(err)

	longestChildBlockchain := []string{}
	for child := range children {
		fmt.Println(child)
		childBlockchain := getLongestBlockchain(child, canvas)
		if len(childBlockchain) > len(longestChildBlockchain) {
			longestChildBlockchain = childBlockchain
		}
	}

	return append(longestBlockchain, longestChildBlockchain...)
}

// Generate an HTML file, filled exclusively with 
// HTML SVG strings from the longest blockchain in canvas
func generateHTML(canvas blockartlib.Canvas) {
	// Create a blank HTML file
	HTML, err := os.Create("./art-app-1.html")
	checkError(err)
	defer HTML.Close()

	// Append starting HTML tags
	pre := []byte("<html>\n<body>\n")
	HTML.Write(pre)

	// Get the longest blockchain
	// Start with the genesis block and recursively add to chain
	genesisHash, err := canvas.GetGenesisBlock()
	checkError(err)
	fmt.Println(genesisHash)
	blockchain := getLongestBlockchain(genesisHash, canvas)

	// Add the HTML SVG string of each opeartion in the blockchain
	for blockHash := range blockchain {
		for shapeHash, _ := range canvas.GetShapes(blockHash) {
			HTMLSVGString := utils.GetHTMLSVGString(shapeHash)
			HTML.Write([]byte(HTMLSVGString + "\n"))
		}
	}

	// Append ending HTML tags
	suf := []byte("</body>\n</html>\n")
	HTML.Write(suf)
}
