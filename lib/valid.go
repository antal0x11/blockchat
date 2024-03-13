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

func ValidateTransaction(t *dst.Transaction, neighboors *dst.Neighboors, node *dst.Node, _mapNodeId map[string]uint32) bool {

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

		// Updating the state of the balance in the neighboor map
		idx := _mapNodeId[t.SenderAddress]
		if remainingBalance := neighboors.DSNodes[idx].Balance - t.Fee - t.Amount; remainingBalance > 0 {
			neighboors.DSNodes[idx].Balance = remainingBalance
		} else {
			neighboors.Mu.Unlock()
			return false
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
