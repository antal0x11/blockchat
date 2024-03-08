package main

import (
	// "context"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

func logError(err error, msg string) {
	if err != nil {
		log.Panicf("%v: %v", msg, err)
	}
}

func commandReceiver(linker chan string) {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	logError(err, "# Can't connect to rabbit.")

	defer conn.Close()

	channel, err := conn.Channel()
	logError(err, "# Can't create channel.")

	defer channel.Close()

	_q, err := channel.QueueDeclare(
		"command_request_stream",
		true,
		false,
		false,
		false,
		nil,
	)
	logError(err, "# Failed to declare queue.")

	cmd, err := channel.Consume(
		_q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	logError(err, "# Failed to register consumer.")

	go func() {
		for cmd != nil {
			for _msg := range cmd {
				linker <- string(_msg.Body)
			}
		}
	}()
}

func main() {

	fmt.Println("# Service running")

	var _w sync.WaitGroup
	_m := make(chan string)
	go commandReceiver(_m)

	_w.Wait()
	fmt.Println("# Received a message: ", <-_m)

}
