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
	minerAddr := "127.0.0.1:8080"
	privKeyString := "30770201010420717b415578a72a446ff6d844e72a25f0def82e51206c286ae837fb1ea6478f7aa00a06082a8648ce3d030107a14403420004b8d0de4d1eab31ddf95cb466b4001356acf17f49c1b8dc3cc78cdb7cdb21aee9262b2551fa977e9a6a2b77294d233bdbc38aae74a9bed79b4cf5d0feab35009c"
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
