package dst

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Transaction struct {
	SenderAddress     string
	RecipientAddress  string
	TypeOfTransaction string
	Amount            float64
	Message           string
	Nonce             uint32
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
	Id uint32
	// Ip net.IP maybe no need for ip
	// Port      uint32 maybe no need for port
	BootStrap bool
	Nonce     uint32
	Stake     uint32
	PublicKey string
	Balance   uint32
	Validator string
	Mu        sync.Mutex
}

type TransactionJSON struct {
	SenderAddress     string  `json:"sender_address"`
	RecipientAddress  string  `json:"recipient_address"`
	TypeOfTransaction string  `json:"type_of_transaction"`
	Amount            float64 `json:"amount,omitempty"`
	Message           string  `json:"message,omitempty"`
	Nonce             uint32  `json:"nonce"`
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
