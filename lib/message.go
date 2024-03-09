package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/antal0x11/blockchat/dst"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TransactionCosumer(node *dst.Node) {
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
	counter := 1
	limit, err := strconv.ParseInt(os.Getenv("BLOCK_CAPACITY"), 10, 8)
	if err != nil {
		log.Fatal("# [TransactionExchangeConsumer] Can't load capacity configuration.")
	}
	var block []dst.TransactionJSON
	go func() {
		for _t := range _transaction {

			fmt.Println("# [TransactionExchangeConsumer] Received a transaction.")
			var data *dst.TransactionJSON
			err := json.Unmarshal(_t.Body, &data)
			if err != nil {
				log.Fatal("# [TransactionExchangeConsumer] Failed to create Transaction Object.")
			}

			data.Nonce = node.Nonce

			if counter == int(limit)-1 {

				block = append(block, *data)

				fmt.Printf("# [TransactionExchangeConsumer] Reached %d Transactions.\n", len(block))
				fmt.Printf("# [TransactionExchangeConsumer] Sending block to block Publisher.\n")

				// For now every node sends the block
				// only the node that the PoS points
				// will be able to send it

				node.Mu.Lock()

				node.Validator = node.PublicKey

				node.Mu.Unlock()

				_b := dst.BlockJSON{
					Index:        1,
					Transactions: block,
					Validator:    node.Validator,
					Hash:         "hash of block",
					PreviousHash: "previous hash of block",
					Capacity:     uint32(len(block)),
				}

				BlockPublisher(_b, node)

				block = nil
				counter = 1
			} else {
				block = append(block, *data)
				counter = len(block)

				node.Mu.Lock()

				node.Nonce++

				node.Mu.Unlock()
			}
		}
	}()
	<-wait
}

func BlockConsumer(node *dst.Node) {

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
			var data *dst.BlockJSON
			err := json.Unmarshal(_b.Body, &data)
			if err != nil {
				log.Fatal("# [BlockExchangeConsumer] Failed to create Block Object.")
			}
			if node.Validator != data.Validator {
				fmt.Printf("# [BlockExchangeConsumer] Block received with index:%d is not valid.", data.Index)
				// TODO discard invalid block
			} else {
				fmt.Printf("# [BlockExchangeConsumer] Block received with index:%d is valid.", data.Index)

				// TODO add valid block to blockchain
				// TODO loop through transactions to update neighboor states
			}
			//fmt.Println(string(_b.Body[:]))

		}
	}()
	<-loop
}

func BlockPublisher(_blockToPublish dst.BlockJSON, _node *dst.Node) {

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
