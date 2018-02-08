package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
)

// To encode publicKey use:
// publicKeyBytes, _ = x509.MarshalPKIXPublicKey(&private_key.PublicKey)

// Private Key:
// 30770201010420717b415578a72a446ff6d844e72a25f0def82e51206c286ae837fb1ea6478f7aa00a06082a8648ce3d030107a14403420004b8d0de4d1eab31ddf95cb466b4001356acf17f49c1b8dc3cc78cdb7cdb21aee9262b2551fa977e9a6a2b77294d233bdbc38aae74a9bed79b4cf5d0feab35009c

// Public Key:
// 3059301306072a8648ce3d020106082a8648ce3d03010703420004b8d0de4d1eab31ddf95cb466b4001356acf17f49c1b8dc3cc78cdb7cdb21aee9262b2551fa977e9a6a2b77294d233bdbc38aae74a9bed79b4cf5d0feab35009c

func main() {
	p256 := elliptic.P256()
	priv1, _ := ecdsa.GenerateKey(p256, rand.Reader)

	privateKeyBytes, _ := x509.MarshalECPrivateKey(priv1)

	encodedPrivateBytes := hex.EncodeToString(privateKeyBytes)
	fmt.Printf("Private: %s\n", encodedPrivateBytes)

	privateKeyBytesRestored, _ := hex.DecodeString(encodedPrivateBytes)
	priv2, _ := x509.ParseECPrivateKey(privateKeyBytesRestored)

	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&priv1.PublicKey)
	encodedPublicBytes := hex.EncodeToString(publicKeyBytes)

	fmt.Println("Public key is: %s", encodedPublicBytes)
	fmt.Println("Private key is: %s", encodedPrivateBytes)
	data := []byte("data")
	// Signing by priv1
	r, s, _ := ecdsa.Sign(rand.Reader, priv1, data)

	// Verifying against priv2 (restored from priv1)
	if !ecdsa.Verify(&priv2.PublicKey, data, r, s) {
		fmt.Printf("Error")
		return
	}

	fmt.Printf("Key was restored from string successfully")
}
