/*

This art-app is purely used to generate the HTML file

*/

package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import "./blockartlib"

import (
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	// Read file content and cast to string
	ipPortBytes, err := ioutil.ReadFile("./ip-ports.txt")
	checkError(err)
	ipPortString := string(ipPortBytes[:])

	keyPairsBytes, err := ioutil.ReadFile("./key-pairs.txt")
	checkError(err)
	keyPairsString := string(keyPairsBytes[:])

	// Parse ip-port and privKey from content string
	minerAddr := strings.Split(ipPortString, "\n")[0]
	privKeyString := strings.Split(keyPairsString, "\n")[0]
	privKeyBytes, err := hex.DecodeString(privKeyString)
	checkError(err)
	privKey, err := x509.ParseECPrivateKey(privKeyBytes)
	checkError(err)

	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(minerAddr, *privKey)
	fmt.Println("OpenCanvas")
	checkError(err)

	fmt.Println(canvas)
	fmt.Println(settings)

	generateHTML(canvas)

	// Close the canvas.
	ink1, err := canvas.CloseCanvas()
	fmt.Println("CloseCanvas")
	checkError(err)

	fmt.Println("%d", ink1)
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
	fmt.Println("GetChildren")
	checkError(err)

	longestChildBlockchain := []string{}
	for _, child := range children {
		childBlockchain := getLongestBlockchain(child, canvas)
		if len(childBlockchain) > len(longestChildBlockchain) {
			longestChildBlockchain = childBlockchain
		}
	}

	return append(longestBlockchain, longestChildBlockchain...)
}

// Generate an HTML file, filled exclusively with 
// HTML SVG strings from the longest blockchain in canvas
func generateHTML(canvas blockartlib.Canvas, settings blockartlib.CanvasSettings) {
	// Create a blank HTML file
	HTML, err := os.Create("./art-app.html")
	checkError(err)
	defer HTML.Close()

	// Append starting HTML tags
	pre := []byte("<!DOCTYPE html>\n<html>\n<head>\n\t<title>HTML SVG Output</title>\n</head>\n")
	bodyString := fmt.Sprintf("<body>\n\t<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" version=\"1.1\">\n", settings.CanvasXMax, settings.CanvasYMax)
	body := []byte(bodyString)
	HTML.Write(pre)
	HTML.Write(body)

	// Get the longest blockchain
	// Start with the genesis block and recursively add to chain
	gHash, err := canvas.GetGenesisBlock()
	fmt.Println("GetGenesisBlock")
	checkError(err)
	blockchain := getLongestBlockchain(gHash, canvas)

	// Add the HTML SVG string of each opeartion in the blockchain
	for _, bHash := range blockchain {
		sHashes, err := canvas.GetShapes(bHash)
		fmt.Println("GetShapes")
		checkError(err)
		for _, sHash := range sHashes {
			HTMLSVGString, err := canvas.GetSvgString(sHash)
			// Expect to see an InvalidShapeHashError
			// as the first line was deleted, but art-node can
			// never tell strictly by shapeHash
			if err == nil {
				HTML.Write([]byte("\t\t" + HTMLSVGString + "\n"))
			} else {
				break
			}
		}
	}

	// Append ending HTML tags
	suf := []byte("\t</svg>\n</body>\n</html>\n")
	HTML.Write(suf)
}
