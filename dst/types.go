package dst

import (
	"sync"
)

type Transaction struct {
	SenderAddress     string  `json:"sender_address"`
	RecipientAddress  string  `json:"recipient_address"`
	TypeOfTransaction string  `json:"type_of_transaction"`
	Amount            float64 `json:"amount,omitempty"`
	Message           string  `json:"message,omitempty"`
	Fee               float64 `json:"fee"`
	Nonce             uint32  `json:"nonce"`
	TransactionId     string  `json:"transaction_id"`
	Signature         []byte  `json:"signature"`
}

type Block struct {
	Index        uint32        `json:"index"`
	Transactions []Transaction `json:"transactions"`
	Validator    string        `json:"validator"`
	Hash         string        `json:"hash"`
	PreviousHash string        `json:"previous_hash"`
	Capacity     uint32        `json:"capacity"`
}

type BlockUnHashed struct {
	Index        uint32
	Transactions []Transaction
	Validator    string
	PreviousHash string
	Capacity     uint32
}

type Node struct {
	// Ip net.IP maybe no need for ip
	// Port      uint32 maybe no need for port
	Id         uint32     `json:"id"`
	BootStrap  bool       `json:"bootstrap"`
	Nonce      uint32     `json:"nonce"`
	Stake      float64    `json:"stake"`
	PublicKey  string     `json:"public_key"`
	Balance    float64    `json:"balance"`
	Validator  string     `json:"validator"`
	BlockChain []Block    `json:"blockchain"`
	Mu         sync.Mutex `json:"-"`
}

type NeighboorNode struct {
	Id        uint32  `json:"id"`
	BootStrap bool    `json:"bootstrap"`
	PublicKey string  `json:"public_key"`
	Balance   float64 `json:"balance"`
	Stake     float64 `json:"stake"`
}

// type Neighboors struct {
// 	DSNodes []NeighboorNode `json:"neighboors"`
// 	Mu      sync.Mutex      `json:"-"`
// }

type Neighboors struct {
	DSNodes map[uint32]*NeighboorNode `json:"node"`
	Mu      sync.Mutex                `json:"-"`
}

type NeighboorInformationMessage struct {
	Info       *NeighboorNode
	Peers      map[uint32]*NeighboorNode
	Blockchain []Block
}

type TransactionResponse struct {
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
}

type TransactionRequest struct {
	RecipientAddress string  `json:"recipient_address"`
	Amount           float64 `json:"amount,omitempty"`
	Message          string  `json:"message,omitempty"`
}

type BalanceResponse struct {
	Status    string  `json:"status"`
	Reason    string  `json:"reason,omitempty"`
	Timestamp string  `json:"timestamp"`
	Balance   float64 `json:"balance,omitempty"`
}
