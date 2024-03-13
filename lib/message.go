package lib

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/antal0x11/blockchat/dst"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TransactionConsumer(node *dst.Node, neighboors *dst.Neighboors, wallet Wallet, _mapNodeId map[string]uint32) {
	connectionURL := os.Getenv("CONNECTION_URL")

	conn, err := amqp.Dial(connectionURL)
	LogError(err, "# [TransactionExchangeConsumer] Couldn't establish a connection with RabbitMQ server.")
	defer conn.Close()

	channel, err := conn.Channel()
	LogError(err, "# [TransactionExchangeConsumer] Couldn't create a channel.")
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
	LogError(err, "# [TransactionExchangeConsumer] Failed to declare the transactions exchange.")

	assignedQueue, err := channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	LogError(err, "# [TransactionExchangeConsumer] Failed to declare queue.")

	err = channel.QueueBind(
		assignedQueue.Name,
		"",
		"transactions",
		false,
		nil,
	)
	LogError(err, "# [TransactionExchangeConsumer] Failed to bind a queue.")

	_transaction, err := channel.Consume(
		assignedQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	LogError(err, "# [TransactionExchangeConsumer] Failed to consume transaction from channel.")

	wait := make(chan int)
	limit, err := strconv.ParseInt(os.Getenv("BLOCK_CAPACITY"), 10, 8)

	if err != nil {
		log.Fatal("# [TransactionExchangeConsumer] Can't load capacity configuration.")
	}
	var block []dst.Transaction
	go func() {
		for _t := range _transaction {

			fmt.Println("# [TransactionExchangeConsumer] Received a transaction.")
			var data *dst.Transaction
			err := json.Unmarshal(_t.Body, &data)
			if err != nil {
				log.Fatal("# [TransactionExchangeConsumer] Failed to create Transaction Object.")
			}

			_isValid := ValidateTransaction(data, neighboors, node, _mapNodeId)
			if _isValid {
				block = append(block, *data)

				if len(block) == int(limit) {

					fmt.Printf("# [TransactionExchangeConsumer] Reached Max Block Capacity of %d transactions.\n", len(block))
					fmt.Println("# [[TransactionExchangeConsumer] Starting PoS to select validator.")

					selectedPoSValidator := MineBlock(&node.BlockChain[len(node.BlockChain)-1].Hash, neighboors)

					node.Mu.Lock()

					node.Validator = selectedPoSValidator

					node.Mu.Unlock()

					fmt.Println("# [TransactionExchangeConsumer] PoS completed.")

					if selectedPoSValidator == node.PublicKey {
						fmt.Println("# [TransactionExchangeConsumer] I am block validator.")
						fmt.Println("# [TransactionExchangeConsumer] Sending block to block Publisher.")
						_b := dst.Block{
							Index:        uint32(len(node.BlockChain)),
							Transactions: block,
							Validator:    node.Validator,
							PreviousHash: node.BlockChain[len(node.BlockChain)-1].PreviousHash,
							Capacity:     uint32(len(block)),
						}

						BlockPublisher(_b, node)
					} else {
						fmt.Println("# [TransactionExchangeConsumer] I am not block validator.")
					}

					block = nil
				}
			}
		}
	}()
	<-wait
}

func BlockConsumer(node *dst.Node, neighboors *dst.Neighboors, _mapNodeId map[string]uint32) {

	connectionURL := os.Getenv("CONNECTION_URL")

	conn, err := amqp.Dial(connectionURL)
	LogError(err, "# [BlockExchangeConsumer] Couldn't establish a connection with RabbitMQ server.")
	defer conn.Close()

	channel, err := conn.Channel()
	LogError(err, "# [BlockExchangeConsumer] Couldn't create a channel.")
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"blocks",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	LogError(err, "# [BlockExchangeConsumer] Failed to declare the blocks exchange.")

	assignedQueue, err := channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	LogError(err, "# [BlockExchangeConsumer] Failed to declare queue.")

	err = channel.QueueBind(
		assignedQueue.Name,
		"",
		"blocks",
		false,
		nil,
	)
	LogError(err, "# [BlockExchangeConsumer] Failed to bind a queue.")

	_blocks, err := channel.Consume(
		assignedQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	LogError(err, "# [BlockExchangeConsumer] Failed to consume block from channel.")

	loop := make(chan int)
	go func() {
		for _b := range _blocks {

			fmt.Println("# [BlockExchangeConsumer] Received a block.")
			var data *dst.Block
			err := json.Unmarshal(_b.Body, &data)
			if err != nil {
				log.Fatal("# [BlockExchangeConsumer] Failed to create Block Object.")
			}
			if node.Validator != data.Validator && node.BlockChain[len(node.BlockChain)-1].Hash == data.Hash && data.Index != 0 {
				fmt.Printf("# [BlockExchangeConsumer] Block received with index:%d is not valid.\n", data.Index)

				// TODO discard invalid block
				// invert the balance state

			} else {
				fmt.Printf("# [BlockExchangeConsumer] Block received with index:%d is valid.\n", data.Index)

				node.Mu.Lock()

				node.BlockChain = append(node.BlockChain, *data)

				node.Mu.Unlock()

				fmt.Printf("# [BlockExchangeConsumer] Block with index:%d is pushed to Blockchain.\n", data.Index)

				neighboors.Mu.Lock()
				node.Mu.Lock()

				for _, _transactionsInValidBlock := range node.BlockChain[len(node.BlockChain)-1].Transactions {

					if node.Validator == node.PublicKey {
						node.Balance += _transactionsInValidBlock.Fee
					}

					if node.PublicKey == _transactionsInValidBlock.RecipientAddress {
						node.Balance += (_transactionsInValidBlock.Amount)
					}

					idx := _mapNodeId[_transactionsInValidBlock.SenderAddress]
					neighboors.DSNodes[idx].Balance -= _transactionsInValidBlock.Amount

					idx = _mapNodeId[_transactionsInValidBlock.RecipientAddress]
					neighboors.DSNodes[idx].Balance += _transactionsInValidBlock.Amount

					if node.Validator == neighboors.DSNodes[idx].PublicKey {
						neighboors.DSNodes[idx].Balance += _transactionsInValidBlock.Fee
					}
				}

				neighboors.Mu.Unlock()
				node.Mu.Unlock()
			}
		}
	}()
	<-loop
}

func BlockPublisher(_blockToPublish dst.Block, _node *dst.Node) {

	connectionURL := os.Getenv("CONNECTION_URL")

	conn, err := amqp.Dial(connectionURL)
	LogError(err, "# [BlockPublisher] Couldn't establish a connection with RabbitMQ server.")
	defer conn.Close()

	channel, err := conn.Channel()
	LogError(err, "# [BlockPublisher] Couldn't create a channel.")
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"blocks",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	LogError(err, "# [BlockPublisher] Failed to declare the blocks exchange.")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_bHash := dst.BlockUnHashed{
		Index:        _blockToPublish.Index,
		Transactions: _blockToPublish.Transactions,
		Validator:    _blockToPublish.Validator,
		PreviousHash: _blockToPublish.PreviousHash,
		Capacity:     _blockToPublish.Capacity,
	}

	_bHashBody, err := json.Marshal(_bHash)
	if err != nil {
		log.Fatal("# [NODE] Failed to initialize genesis block.")
	}

	blockHash := sha256.Sum256(_bHashBody)

	_blockToPublish.Hash = fmt.Sprintf("%x", blockHash)
	_blockToPublish.PreviousHash = _node.BlockChain[len(_node.BlockChain)-1].Hash

	body, err := json.Marshal(_blockToPublish)
	LogError(err, "# [BlockPublisher] Failed to create json message from block.")
	err = channel.PublishWithContext(ctx,
		"blocks",
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		})
	LogError(err, "# [BlockPublisher] Failed to publish block.")
	fmt.Println("# [BlockPublisher] Block Sent.")
}

func LogError(err error, msg string) {
	if err != nil {
		log.Panicf("%v: %v", msg, err)
	}
}
