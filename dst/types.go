package dst

import (
	"encoding/json"
	"fmt"
	"net"
)

type Transaction struct {
	SenderAddress     string
	RecipientAddress  string
	TypeOfTransaction string
	Amount            float64
	Message           string
	Nonce             int32
	TransactionId     string
	Signature         string
}

type Block struct {
	Index        uint32
	Transactions []Transaction
	Validator    string
	Hash         string
	PreviousHash string
	Capacity     uint32
}

type Node struct {
	Id        uint32
	Ip        net.IP
	Port      uint32
	BootStrap bool
	Stake     uint32
	PublicKey string
	Balance   uint32
}

type Wallet struct {
	PublicKey  string
	PrivateKey string
}

type TransactionJSON struct {
	SenderAddress     string  `json:"sender_address"`
	RecipientAddress  string  `json:"recipient_address"`
	TypeOfTransaction string  `json:"type_of_transaction"`
	Amount            float64 `json:"amount,omitempty"`
	Message           string  `json:"message,omitempty"`
	Nonce             int32   `json:"nonce"`
	TransactionId     string  `json:"transaction_id"`
	Signature         string  `json:"signature"`
}

type BlockJSON struct {
	Index        uint32            `json:"index"`
	Transactions []TransactionJSON `json:"transactions"`
	Validator    string            `json:"validator"`
	Hash         string            `json:"hash"`
	PreviousHash string            `json:"previous_hash"`
	Capacity     uint32            `json:"capacity"`
}

func (_t Transaction) MarshalJSON() ([]byte, error) {

	_m, err := json.Marshal(TransactionJSON(_t))
	if err != nil {
		fmt.Println("# Failed to make json from trasaction.", err)
	}
	return _m, err
}

func (_b Block) MarshalJSON() ([]byte, error) {

	var transactions []TransactionJSON
	for _, _t := range _b.Transactions {
		_tmp := TransactionJSON(_t)
		transactions = append(transactions, _tmp)
	}

	_m := BlockJSON{
		Index:        _b.Index,
		Transactions: transactions,
		Validator:    _b.Validator,
		Hash:         _b.Hash,
		PreviousHash: _b.PreviousHash,
		Capacity:     _b.Capacity,
	}

	return json.Marshal(_m)
}
