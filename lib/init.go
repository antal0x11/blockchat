package lib

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/antal0x11/blockchat/dst"
)

// Creates the genesis block, updates bootstrap nodes balance
// and adds genesis block to the blockchain.
func BootStrapBlockInitialize(node *dst.Node, wallet Wallet) {

	amount, err := strconv.ParseInt(os.Getenv("NEIGHBOORS"), 10, 8)
	if err != nil {
		log.Fatal("# [NODE] Invalid configuration for neighboors\n")
	}

	genesisTransaction := dst.Transaction{
		SenderAddress:     "0",
		RecipientAddress:  node.PublicKey,
		TypeOfTransaction: "coins",
		Amount:            float64(2000 * amount),
		Nonce:             node.Nonce,
	}

	unsignedGenesisTransaction := UnsignedTransaction{
		SenderAddress:     genesisTransaction.SenderAddress,
		RecipientAddress:  genesisTransaction.RecipientAddress,
		TypeOfTransaction: genesisTransaction.TypeOfTransaction,
		Amount:            genesisTransaction.Amount,
		Nonce:             genesisTransaction.Nonce,
	}

	genesisTransactionID, genesisSignature := wallet.SignTransaction(unsignedGenesisTransaction)
	genesisTransaction.TransactionId = genesisTransactionID
	genesisTransaction.Signature = genesisSignature

	genesisBlock := dst.Block{
		Index:        0,
		Validator:    "0",
		PreviousHash: "1",
		Capacity:     1,
	}

	genesisBlock.Transactions = append(genesisBlock.Transactions, genesisTransaction)

	_bHash := dst.BlockUnHashed{
		Index:        genesisBlock.Index,
		Transactions: genesisBlock.Transactions,
		Validator:    genesisBlock.Validator,
		PreviousHash: genesisBlock.PreviousHash,
		Capacity:     genesisBlock.Capacity,
	}

	_bHashBody, err := json.Marshal(_bHash)
	if err != nil {
		log.Fatal("# [NODE] Failed to initialize genesis block.")
	}

	blockHash := sha256.Sum256(_bHashBody)

	genesisBlock.Hash = fmt.Sprintf("%x", blockHash)

	node.Mu.Lock()

	node.Nonce++
	node.BlockChain = append(node.BlockChain, genesisBlock)
	node.Balance = genesisTransaction.Amount

	node.Mu.Unlock()

}
