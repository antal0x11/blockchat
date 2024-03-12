package lib

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/antal0x11/blockchat/dst"
)

func ValidateTransaction(t *dst.Transaction, neighboors *dst.Neighboors) bool {

	// TODO check balance of sender and modify it

	_b, _ := pem.Decode([]byte(t.SenderAddress))
	if _b == nil {
		log.Fatal("# [ValidateTransaction] Failed to decode Public Key.")
	}

	pubKey, err := x509.ParsePKCS1PublicKey(_b.Bytes)
	if err != nil {
		log.Fatal("# [ValidateTransaction] Failed to parse sender address to rsa public key.")
	}

	hash, err := hex.DecodeString(t.TransactionId)
	if err != nil {
		log.Fatal("# [ValidateTransaction] Failed to decode transaction ID to [32]bytes.")
	}

	var _bHash [32]byte
	copy(_bHash[:], hash)

	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, _bHash[:], t.Signature)

	if err == nil {
		fmt.Println("# [ValidateTransaction] Transaction Signature is valid.")
	} else {
		fmt.Println("# [ValidateTransaction] Transaction Signature is not valid.")
	}

	return err == nil
}
