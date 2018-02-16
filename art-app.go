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

	// 	<svg>
	// <path d="M 480 40 L 430 120 L 480 150 L 520 120 H 520 L 480 40" fill="red" stroke="red"></path>
	// <path d="M 420 130 L 350 230 L 480 300 V 160 L 420 130" fill="transparent" stroke="red"> </path>
	// <path d="M 490 160 L 530 140 L 610 240 L 490 300 Z" fill="blue" stroke="blue"></path>
	// <path d="M 761 78 L 741 58 H 711 L 691 78 V 98 L 711 118 H 721 L 758 117 L 770 140 V 160 L 750 180 H 710 L 690 160" fill="transparent" stroke="green"></path>
	// <path d="M 700 40 L 720 200" fill="transparent" stroke="green"></path>
	// <path d="M 720 40 L 740 200" fill="transparent" stroke="green"></path>
	// <path d="M 280 140 L 560 50" fill="transparent" stroke="red"></path>
	// <path d="M 280 140 L 560 50" fill="transpraent" stroke="purple"><path>
	// </svg>
	minerAddr := "127.0.0.1:50417"

	privKeyString := "3081a402010104306c9de9cce82755eca357beed8f1c9e9df8594ce575127fe10486cbb6bb87d3d5e3a0e2fb4d6fab7fcbce5a564f313bf2a00706052b81040022a164036200043bdd1a0e32123cf670d74ee918ef4c42a334190dfafcf93ca66955561ff85d49727076dd57705b9f904961292b352fda712c1b546ea3362c3fa63e147add351321c17189ad8b3ada63b0979905b67ca57726193ff939af38ef3aa407424ac55f"
	privateKeyBytes, _ := hex.DecodeString(privKeyString)
	privKey, _ := x509.ParseECPrivateKey(privateKeyBytes)
	// TODO: use crypto/ecdsa to read pub/priv keys from a file argument.

	// Open a canvas.
	canvas, settings, err := blockartlib.OpenCanvas(minerAddr, *privKey)
	if checkError(err) != nil {
		//return
	}

	fmt.Println(canvas)
	fmt.Println(settings)

	validateNum := uint8(2)

	// Add a line.
	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 0 5", "transparent", "red")
	if checkError(err) != nil {
		return
	}

	fmt.Println("%s, %s, %d", shapeHash, blockHash, ink)
	// Add another line.
	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 5 0", "transparent", "blue")
	if checkError(err) != nil {
		return
	}

	fmt.Println("%s, %s, %d", shapeHash2, blockHash2, ink2)
	// Delete the first line.
	ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	if checkError(err) != nil {
		return
	}

	fmt.Println("%d", ink3)
	// assert ink3 > ink2

	// Close the canvas.
	ink4, err := canvas.CloseCanvas()
	if checkError(err) != nil {
		return
	}

	fmt.Println("%d", ink4)
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}
