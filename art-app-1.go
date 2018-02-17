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
)

func main() {
	// Open files to read from
	ipPort, err := os.Open("./ip-ports")
	checkError(err)
	keyPairs, err := os.Open("./key-pairs")
	checkError(err)

	// Read ip-port and privKey from files
	ipScanner := bufio.NewScanner(ipPort)
	keyScanner := bufio.NewScanner(keyPairs)

	minerAddr := ipScanner.ReadBytes([]byte("\n"))
	privKeyBytes := keyScanner.ReadBytes([]byte("\n"))
	privKey, _ := x509.ParseECPrivateKey(privateKeyBytes)

	// Once finished, delete files
	ipPort.Close()
	keyPairs.Close()
	_ := os.Remove("./ip-ports")
	_ := os.Remove("./key-pairs")

	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(minerAddr, *privKey)
	checkError(err)

	fmt.Println(canvas)
	fmt.Println(settings)

	validateNum := uint8(3)

	// First corner: (0,0)
	// Draw horizontal lines
	shapeHash1, blockHash1, ink1, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 5 0", "transparent", "red")
	checkError(err)
	fmt.Println("%s, %s, %d", shapeHash1, blockHash1, ink1)

	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 5 L 5 0", "transparent", "red")
	checkError(err)
	fmt.Println("%s, %s, %d", shapeHash2, blockHash2, ink2)

	// Draw vertical lines
	shapeHash3, blockHash3, ink3, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 0 5", "transparent", "red")
	checkError(err)
	fmt.Println("%s, %s, %d", shapeHash3, blockHash3, ink3)

	shapeHash4, blockHash4, ink4, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 5 0 L 0 5", "transparent", "red")
	checkError(err)
	fmt.Println("%s, %s, %d", shapeHash3, blockHash3, ink3)

	// Second corner: (CanvasXMax,CanvasYMax)
	// Draw horizontal lines
	svg1 := fmt.Sprintf("M %d %d L -5 0", settings.CanvasXMax, settings.CanvasYMax)
	shapeHash5, blockHash5, ink5, err := canvas.AddShape(validateNum, blockartlib.PATH, svg1, "transparent", "blue")
	checkError(err)
	fmt.Println("%s, %s, %d", shapeHash5, blockHash5, ink5)


	svg2 := fmt.Sprintf("M %d %d L -5 0", settings.CanvasXMax, settings.CanvasYMax - 5)
	shapeHash6, blockHash6, ink6, err := canvas.AddShape(validateNum, blockartlib.PATH, svg2, "transparent", "blue")
	checkError(err)
	fmt.Println("%s, %s, %d", shapeHash6, blockHash6, ink6)

	// Draw vertical lines
	svg3 := fmt.Sprintf("M %d %d L -5 0", settings.CanvasXMax, settings.CanvasYMax)
	shapeHash7, blockHash7, ink7, err := canvas.AddShape(validateNum, blockartlib.PATH, svg3, "transparent", "blue")
	checkError(err)
	fmt.Println("%s, %s, %d", shapeHash7, blockHash7, ink7)

	svg4 := fmt.Sprintf("M %d %d L -5 0", settings.CanvasXMax, settings.CanvasYMax - 5)
	shapeHash8, blockHash8, ink8, err := canvas.AddShape(validateNum, blockartlib.PATH, svg4, "transparent", "blue")
	checkError(err)
	fmt.Println("%s, %s, %d", shapeHash8, blockHash8, ink8)

	// Jan's SVG
	svg5 := "M 850 850 L850 50 50 50 50 850 850 850 M 750 750 150 750 150 150 750 150 750 750"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg5, "darkgrey", "darkgrey")
	svg6 := "M 885 885 L885 15 15 15 15 885 885 885 M 860 860 40 860 40 40 860 40 860 860"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg6, "#232323", "darkgrey")
	svg7 := "M 550 674 L700 449 550 224 350 224 200 449 350 674 550 674 M540 654 L678 449 540 244 360 244 222 449 360 654 540 654"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg7, "#666666", "darkgrey")
	svg8 := "circle x:449 y:449 r:175"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.CIRCLE, svg8, "transparent", "black")
	svg9 := "circle x:449 y:449 r:170"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.CIRCLE, svg9, "transparent", "black")
	svg10 := "circle x:449 y:449 r:165"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.CIRCLE, svg10, "transparent", "black")
	svg11 := "circle x:449 y:449 r:55"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.CIRCLE, svg11, "transparent", "black")
	svg12 := "circle x:449 y:449 r:50"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.CIRCLE, svg12, "#555555", "black")
	svg13 := "circle x:519 y:519 r:25"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.CIRCLE, svg13, "#999999", "#999999")
	svg14 := "circle x:519 y:379 r:25"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.CIRCLE, svg14, "#999999", "#999999")
	svg15 := "circle x:379 y:519 r:25"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.CIRCLE, svg15, "#999999", "#999999")
	svg16 := "circle x:379 y:379 r:25"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.CIRCLE, svg16, "#999999", "#999999")
	svg17 := "M725 725 L725 650 650 725 Z"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg17, "#222222", "#222222")
	svg18 := "M175 725 L175 650 250 725 Z"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg18, "#222222", "#222222")
	svg19 := "M175 175 L175 250 250 175 Z"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg19, "#222222", "#222222")
	svg20 := "M725 175 L650 175 725 250 Z"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg20, "#222222", "#222222")
	svg21 := "M449 600 L439 540 449 520 459 540 Z"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg21, "#222222", "#222222")
	svg22 := "M449 300 L439 360 449 380 459 360 Z"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg22, "#222222", "#222222")
	svg23 := "M300 449 L360 439 380 449 360 459 Z"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg23, "#222222", "#222222")
	svg24 := "M600 449 L540 439 520 449 540 459 Z"
	_, _, _, err := canvas.AddShape(validateNum, blockartlib.PATH, svg24, "#222222", "#222222")

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
func generateHTML(canvas blockartlib.Canvas) {
	// Create a blank HTML file
	HTML, err := os.Create("./art-app-1.html")
	checkError(err)
	defer HTML.Close()

	// Append starting HTML tags
	pre := []byte("<!DOCTYPE html>\n<html>\n<head>\n\t<title>HTML SVG Output</title>\n</head>\n")
	body := []byte("<body>\n\t<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"900\" height=\"900\" version=\"1.1\">\n")
	HTML.Write(pre)
	HTML.Write(Body)

	// Get the longest blockchain
	// Start with the genesis block and recursively add to chain
	gHash, err := canvas.GetGenesisBlock()
	checkError(err)
	blockchain := getLongestBlockchain(gHash, canvas)

	// Add the HTML SVG string of each opeartion in the blockchain
	for _, bHash := range blockchain {
		sHashes, err := canvas.GetShapes(bHash)
		checkError(err)
		for _, sHash := range sHashes {
			HTMLSVGString, err := canvas.GetSvgString(sHash)
			// Expect to see an InvalidShapeHashError
			// as the first line was deleted, but art-node can
			// never tell strictly by shapeHash
			if err == nil || err.(blockartlib.InvalidShapeHashError) {
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
