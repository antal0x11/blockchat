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

func BoostrapInformationConsumer(neighboors *dst.Neighboors, loop chan *dst.Neighboors) {

	fmt.Println("# [BootstapInformationConsumer] Waiting messages from neighboors.")

	conn, err := amqp.Dial(os.Getenv("CONNECTION_URL"))
	if err != nil {
		log.Fatal("# [BootstapInformationConsumer] Connection failure.\n")
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("# [BootstapInformationConsumer] Can't create channel.\n")
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"information",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("# [BootstapInformationConsumer] Failed to declare Exchange.\n")
	}

	queue, err := channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("# [BootstapInformationConsumer] Failed to declare queue.\n")
	}

	err = channel.QueueBind(
		queue.Name,
		"bootstrap",
		"information",
		false,
		nil,
	)
	if err != nil {
		log.Fatal("# [BootstapInformationConsumer] Failed to bind to a queue.\n")
	}

	message, err := channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("# [BootstapInformationConsumer] Failed to consume messages from neighboors.\n")
	}

	neighboorsReached, err := strconv.ParseInt(os.Getenv("NEIGHBOORS"), 10, 64)
	if err != nil {
		log.Fatal("# [BootstapInformationConsumer] Failed to load env configuration.\n")
	}

	wait := make(chan bool)
	go func() {
		for _message := range message {

			var _node *dst.NeighboorNode
			err = json.Unmarshal(_message.Body, &_node)
			if err != nil {
				log.Fatal("# [BootstapInformationConsumer] Failed to unmarshall node information.\n")
			}
			neighboors.Mu.Lock()

			_node.Id = uint32(len(neighboors.DSNodes))

			neighboors.DSNodes = append(neighboors.DSNodes, *_node)

			neighboors.Mu.Unlock()

			if len(neighboors.DSNodes) == int(neighboorsReached)+1 {
				loop <- neighboors
				wait <- true
			}
		}
	}()
	<-wait
	fmt.Println("# [BootstapInformationConsumer] Closing BootstrapInformationConsumer, all neighboors have introduced.")
}

func NodeInformationConsumer(neighboors *dst.Neighboors, node *dst.Node) {

	conn, err := amqp.Dial(os.Getenv("CONNECTION_URL"))
	if err != nil {
		log.Fatal("# [NodeInformationConsumer] Connection failure.\n")
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("# [NodeInformationConsumer] Can't create channel.\n")
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"information",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("# [NodeInformationConsumer] Failed to declare Exchange.\n")
	}

	queue, err := channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("# [NodeInformationConsumer] Failed to declare queue.\n")
	}

	err = channel.QueueBind(
		queue.Name,
		node.PublicKey,
		"information",
		false,
		nil,
	)
	if err != nil {
		log.Fatal("# [NodeInformationConsumer] Failed to bind to a queue.\n")
	}

	message, err := channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("# [NodeInformationConsumer] Failed to consume messages from neighboors.\n")
	}

	wait := make(chan bool)
	go func() {
		for _message := range message {

			fmt.Println("# [NodeInformationConsumer] Received feedback.")

			var _infoReceived *dst.NeighboorInformationMessage
			err = json.Unmarshal(_message.Body, &_infoReceived)
			if err != nil {
				log.Fatal("# [NodeInformationConsumer] Failed to unmarshall node information.\n")
			}

			node.Mu.Lock()

			node.Id = _infoReceived.Info.Id
			node.Balance = _infoReceived.Info.Balance
			node.BlockChain = _infoReceived.Blockchain

			node.Mu.Unlock()

			neighboors.Mu.Lock()

			neighboors.DSNodes = _infoReceived.Peers

			neighboors.Mu.Unlock()

			wait <- true
		}
	}()
	<-wait
	fmt.Println("# [NodeInformationConsumer] Closing NodeInformationConsumer, information received.")
}

func NodeInformationPublisher(node *dst.Node) {

	fmt.Println("# [NodeInformationPublisher] Preparing to send information.")

	conn, err := amqp.Dial(os.Getenv("CONNECTION_URL"))
	if err != nil {
		log.Fatal("# [NodeInformationPublisher] Failed to create connection.\n")
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("# [NodeInformationPublisher] Failed to create channel.\n")
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"information",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("# [NodeInformationPublisher] Failed to declare exchange.\n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_info := dst.NeighboorNode{
		BootStrap: node.BootStrap,
		PublicKey: node.PublicKey,
		Balance:   node.Balance,
	}

	body, err := json.Marshal(_info)
	if err != nil {
		log.Fatal("# [NodeInformationPublisher] Failed to marshall neighboor before sending.\n")
	}

	err = channel.PublishWithContext(
		ctx,
		"information",
		"bootstrap",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Fatal("# [NodeInformationPublisher] Failed to publish message to bootstrap.\n")
	}
	fmt.Println("# [NodeInformationPublisher] Sent Information to Bootstrap.")

}

func BootstrapInitTransactionAndBlockChain(loop chan *dst.Neighboors, node *dst.Node, wallet Wallet) {
	neighboors, ok := <-loop

	if ok {

		conn, err := amqp.Dial(os.Getenv("CONNECTION_URL"))
		if err != nil {
			log.Fatal("# [BootstrapInitTransactionAndBlockChain] Failed to create connection.\n")
		}
		defer conn.Close()

		channel, err := conn.Channel()
		if err != nil {
			log.Fatal("# [BootstrapInitTransactionAndBlockChain] Failed to create channel.\n")
		}
		defer channel.Close()

		err = channel.ExchangeDeclare(
			"information",
			"direct",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Fatal("# [BootstrapInitTransactionAndBlockChain] Failed to declare exchange.\n")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, neighboor := range neighboors.DSNodes {

			_info := dst.NeighboorInformationMessage{
				Info:       neighboor,
				Peers:      neighboors.DSNodes,
				Blockchain: node.BlockChain,
			}
			body, err := json.Marshal(_info)
			if err != nil {
				log.Fatal("# [BootstrapInitTransactionAndBlockChain] Failed to marshall neighboor before sending.\n")
			}

			err = channel.PublishWithContext(
				ctx,
				"information",
				neighboor.PublicKey,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				},
			)
			if err != nil {
				log.Fatalf("# [BootstrapInitTransactionAndBlockChain] Failed to publish message to neighboor with id: %d\n", neighboor.Id)
			}
		}

		for _, neighboor := range neighboors.DSNodes {

			node.Mu.Lock()

			transaction := dst.TransactionJSON{
				SenderAddress:    node.PublicKey,
				RecipientAddress: neighboor.PublicKey,
				Amount:           1000, // adjust this value
				Nonce:            node.Nonce,
			}

			fixTransactionFields := UnsignedTransaction{
				SenderAddress:     transaction.SenderAddress,
				RecipientAddress:  transaction.RecipientAddress,
				TypeOfTransaction: transaction.TypeOfTransaction,
				Amount:            transaction.Amount,
				Message:           transaction.Message,
				Nonce:             transaction.Nonce,
			}

			transactionID, signature := wallet.SignTransaction(fixTransactionFields)
			transaction.TransactionId = transactionID
			transaction.Signature = signature

			node.Nonce++

			node.Mu.Unlock()

			go func(_t dst.TransactionJSON) {
				conn, err := amqp.Dial(os.Getenv("CONNECTION_URL"))
				LogError(err, "# [BootstrapInitTransactionAndBlockChain] Couldn't establish a connection with RabbitMQ server.")
				defer conn.Close()

				channel, err := conn.Channel()
				LogError(err, "# [BootstrapInitTransactionAndBlockChain] Couldn't create a channel.")
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
				LogError(err, "# [BootstrapInitTransactionAndBlockChain] Failed to declare the transactions exchange.")

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				body, err := json.Marshal(transaction)
				LogError(err, "# [BootstrapInitTransactionAndBlockChain] Failed to create json message from transaction.")
				err = channel.PublishWithContext(ctx,
					"transactions",
					"",
					false,
					false, amqp.Publishing{
						ContentType: "application/json",
						Body:        []byte(body),
					})
				LogError(err, "# [BootstrapInitTransactionAndBlockChain] Failed to publish transaction.")
				fmt.Println("# [BootstrapInitTransactionAndBlockChain] Transaction Sent.")
			}(transaction)
		}

	} else {
		log.Fatal("# [BootstrapInitTransactionAndBlockChain] Failed to send genesis block and initial Transactions.")
	}
	fmt.Println("# [BootstrapInitTransactionAndBlockChain] Send Genesis Block And Initial Transactions")
}
