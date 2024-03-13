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

func ValidateTransaction(t *dst.Transaction, neighboors *dst.Neighboors, node *dst.Node) bool {

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

		fmt.Println("# [ValidateTransaction] Updating neighboors state.")

		neighboors.Mu.Lock()

		// have to use a map instead of a slice with structs for faster update
		for _n := range neighboors.DSNodes {
			if neighboors.DSNodes[_n].PublicKey == t.SenderAddress {
				remainingBalance := neighboors.DSNodes[_n].Balance - t.Fee - t.Amount
				if remainingBalance > 0 {
					neighboors.DSNodes[_n].Balance = remainingBalance
				} else {
					neighboors.Mu.Unlock()
					return false
				}
			}
		}

		neighboors.Mu.Unlock()

		node.Mu.Lock()

		if node.PublicKey == t.SenderAddress {
			node.Balance = node.Balance - t.Fee - t.Amount
		}

		node.Mu.Unlock()

		fmt.Println("# [ValidateTransaction] Neighboors are updated")
	} else {
		fmt.Println("# [ValidateTransaction] Transaction Signature is not valid.")
	}

	return err == nil
}
