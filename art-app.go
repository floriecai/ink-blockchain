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
	minerAddr := "127.0.0.1:52552"

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

	fmt.Println("added a line:", shapeHash, blockHash, ink)
	// Add another line.
	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 5 0", "transparent", "blue")
	if checkError(err) != nil {
		return
	}

	fmt.Println("added another line", shapeHash2, blockHash2, ink2)
	// Delete the first line.
	ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	if checkError(err) != nil {
		return
	}

	fmt.Println("deleted a line", ink3)
	// assert ink3 > ink2

	// Close the canvas.
	ink4, err := canvas.CloseCanvas()
	if checkError(err) != nil {
		return
	}

	fmt.Println("closed canvas", ink4)
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}
