package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/antal0x11/blockchat/dst"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NodeHttpService(node *dst.Node, neighboors *dst.Neighboors, wallet *Wallet) {

	fmt.Println("# [NodeHttpService] HttpService is running.")
	go func(node *dst.Node, neighboors *dst.Neighboors, wallet *Wallet) {
		http.HandleFunc("/", nodeInfo(node))
		http.HandleFunc("/transaction", createTransaction(node, neighboors, wallet))
		http.HandleFunc("/api/neighboors", neighboorsHttpService(neighboors))
		http.HandleFunc("/api/view", viewLastBlock(node))
		http.HandleFunc("/api/balance", getBalance(node))

		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			log.Fatal("# [NodeHttpService] HttpService failed.")
		}
	}(node, neighboors, wallet)

}

func nodeInfo(node *dst.Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		node.Mu.Lock()

		_res, err := json.Marshal(node)
		if err != nil {
			log.Fatal("# [NodeHttpService] Failed to serialize node.")
		}

		node.Mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, string(_res[:]))
	}
}

func neighboorsHttpService(neighboors *dst.Neighboors) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		neighboors.Mu.Lock()
		_res, err := json.Marshal(neighboors)
		if err != nil {
			log.Fatal("# [NeighboorsHttpService] Failed to serialize neighboors.")
		}

		neighboors.Mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, string(_res[:]))
	}
}

func createTransaction(node *dst.Node, neighboors *dst.Neighboors, wallet *Wallet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			log.Println("# [HttpCreateTransaction] Failed to handle incoming data request.")
			_rMethod := dst.TransactionResponse{
				Timestamp: time.Now().String(),
				Status:    "fail",
				Reason:    "Invalid Request",
			}
			_rMethodResposne, err := json.Marshal(_rMethod)
			if err != nil {
				log.Fatal("# [HttpCreateTransaction] Failed to marshal response.")
			}
			http.Error(w, string(_rMethodResposne[:]), http.StatusBadRequest)
			return
		}

		connectionURL := os.Getenv("CONNECTION_URL")

		conn, err := amqp.Dial(connectionURL)
		if err != nil {
			log.Fatal("# [HttpCreateTransaction] Couldn't establish a connection with RabbitMQ server.")
		}
		defer conn.Close()

		channel, err := conn.Channel()
		if err != nil {
			log.Fatal("# [HttpCreateTransaction] Couldn't create a channel.")
		}
		defer channel.Close()

		err = channel.ExchangeDeclare(
			"transactions",
			"fanout",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Fatal("# [HttpCreateTransaction] Failed to declare the transactions exchange.")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var dataReceived dst.TransactionRequest

		err = json.NewDecoder(r.Body).Decode(&dataReceived)
		if err != nil {
			log.Println("# [HttpCreateTransaction] Failed to handle incoming data request.")
			_fresponse := dst.TransactionResponse{
				Timestamp: time.Now().String(),
				Status:    "fail",
				Reason:    "Invalid Request",
			}
			_failResponse, err := json.Marshal(_fresponse)
			if err != nil {
				log.Fatal("# [HttpCreateTransaction] Failed to marshal response.")
			}
			http.Error(w, string(_failResponse[:]), http.StatusBadRequest)
			return
		}

		// Map the given node id with the corresponding public key for the recipient address
		_nodeId, err := strconv.ParseInt(dataReceived.RecipientAddress, 10, 32)
		if err != nil {
			log.Fatal("# [HttpCreateTransaction Failed to parse recipient address.]")
		}

		_transaction := dst.Transaction{
			SenderAddress:    node.PublicKey,
			RecipientAddress: neighboors.DSNodes[uint32(_nodeId)].PublicKey,
			Nonce:            node.Nonce,
		}

		if dataReceived.Amount == 0 && dataReceived.Message != "" {

			_transaction.TypeOfTransaction = "message"
			_transaction.Message = dataReceived.Message
			_transaction.Fee = float64(len(dataReceived.Message))

		} else if dataReceived.Amount != 0 && dataReceived.Message == "" {

			_transaction.TypeOfTransaction = "coins"
			_transaction.Amount = dataReceived.Amount
			_transaction.Fee = 0.03 * dataReceived.Amount

		} else {

			log.Println("# [HttpCreateTransaction] Failed handle incoming data request.")
			_fResponse := dst.TransactionResponse{
				Timestamp: time.Now().String(),
				Status:    "fail",
				Reason:    "Invalid Request, transaction rejected",
			}
			_failResponse, err := json.Marshal(_fResponse)
			if err != nil {
				log.Fatal("# [HttpCreateTransaction] Failed to marshal response.")
			}
			http.Error(w, string(_failResponse[:]), http.StatusBadRequest)
			return

		}

		fmt.Println("# [HttpCreateTransaction] Finished adding fee to transaction.")

		_unsignedTransaction := UnsignedTransaction{
			SenderAddress:     _transaction.SenderAddress,
			RecipientAddress:  _transaction.RecipientAddress,
			TypeOfTransaction: _transaction.TypeOfTransaction,
			Amount:            _transaction.Amount,
			Message:           _transaction.Message,
			Nonce:             _transaction.Nonce,
		}

		transactionID, signature := wallet.SignTransaction(_unsignedTransaction)
		_transaction.TransactionId = transactionID
		_transaction.Signature = signature

		fmt.Println("# [HttpCreateTransaction] Signed transaction.")

		node.Mu.Lock()

		node.Nonce++

		node.Mu.Unlock()

		body, err := json.Marshal(_transaction)
		if err != nil {
			log.Fatal("# [HttpCreateTransaction] Failed to create json message from transaction.")
		}
		err = channel.PublishWithContext(ctx,
			"transactions",
			"",
			false,
			false, amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(body),
			})
		if err != nil {
			log.Fatal("# [HttpCreateTransaction] Failed to publish transaction.")
		}
		fmt.Println("# [HttpCreateTransaction] Transaction Sent.")

		transactionResponse := dst.TransactionResponse{
			Timestamp: time.Now().String(),
			Status:    "ok",
		}
		_r, err := json.Marshal(transactionResponse)
		if err != nil {
			log.Fatal("# [HttpCreateTransaction] Failed to send response to client.")
		}
		io.WriteString(w, string(_r[:]))
	}
}

func viewLastBlock(node *dst.Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_r := dst.Block{
			Transactions: node.BlockChain[len(node.BlockChain)-1].Transactions,
			Validator:    node.Validator,
		}

		if _v, err := json.Marshal(_r); err != nil {
			log.Fatal("# [HttpServiceViewLastBlock] Failed to marshal last block before sending it.")
		} else {
			io.WriteString(w, string(_v[:]))
		}
	}
}

func getBalance(node *dst.Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_r := dst.BalanceResponse{
			Status:    "ok",
			Timestamp: time.Now().String(),
			Balance:   node.Balance,
		}

		if _balanceResposne, err := json.Marshal(_r); err != nil {
			log.Fatal("# [HttpServiceGetBalance] Failed to marshal balance response.")
		} else {
			io.WriteString(w, string(_balanceResposne[:]))
		}
	}
}
