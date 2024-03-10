package lib

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
)

type Wallet struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

type UnsignedTransaction struct {
	SenderAddress     string
	RecipientAddress  string
	TypeOfTransaction string
	Amount            float64
	Message           string
	Nonce             uint32
}

// Creates a new wallet, a pair of public/private key.
func GenerateWallet() (wallet Wallet) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal("[Node] Failed to Generate Private Key.\n", err)
	}

	publicKey := &privateKey.PublicKey

	return Wallet{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}

func (_w Wallet) SignTransaction(_t UnsignedTransaction) (transactionID string, signature string) {

	_ut, err := json.Marshal(_t)
	if err != nil {
		log.Fatal(" [Node] Failed to create the []byte of transaction.\n")
	}
	hash := sha256.Sum256(_ut)

	sig, err := rsa.SignPKCS1v15(rand.Reader, _w.PrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		log.Fatal(" [Node] Failed to create signature.\n")
	}

	return fmt.Sprintf("%x", hash), fmt.Sprintf("%x", sig)
}

func (_w Wallet) PublicKeyToString() string {

	publickeyBytes := x509.MarshalPKCS1PublicKey(_w.PublicKey)
	publicKeyPEM := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publickeyBytes,
	}
	var publicKeyBuffer bytes.Buffer
	err := pem.Encode(&publicKeyBuffer, &publicKeyPEM)
	if err != nil {
		log.Fatal("[Node] Failed to channel in buffer Public Key.\n", err)
	}

	return publicKeyBuffer.String()
}

func (_w Wallet) PrivateKeyToString() string {

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(_w.PrivateKey)
	privateKeyPEM := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	var privateKeyBuffer bytes.Buffer
	err := pem.Encode(&privateKeyBuffer, &privateKeyPEM)
	if err != nil {
		log.Fatal("[Node] Failed to channel in buffer Private Key.\n", err)
	}

	return privateKeyBuffer.String()
}
