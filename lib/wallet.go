package lib

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"

	"github.com/antal0x11/blockchat/dst"
)

// Creates a new wallet, a pair of public/private key.
func GenerateWallet() (wallet dst.Wallet) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal("[Node] Failed to Generate Private Key.\n", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	var privateKeyBuffer bytes.Buffer
	err = pem.Encode(&privateKeyBuffer, &privateKeyPEM)
	if err != nil {
		log.Fatal("[Node] Failed to channel in buffer Private Key.\n", err)
	}

	publicKey := privateKey.PublicKey
	publickeyBytes := x509.MarshalPKCS1PublicKey(&publicKey)
	publicKeyPEM := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publickeyBytes,
	}
	var publicKeyBuffer bytes.Buffer
	err = pem.Encode(&publicKeyBuffer, &publicKeyPEM)
	if err != nil {
		log.Fatal("[Node] Failed to channel in buffer Public Key.\n", err)
	}

	return dst.Wallet{
		PublicKey:  publicKeyBuffer.String(),
		PrivateKey: privateKeyBuffer.String(),
	}
}
