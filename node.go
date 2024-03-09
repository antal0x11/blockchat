package main

import (
	"fmt"
	"log"

	"github.com/antal0x11/blockchat/lib"
	"github.com/joho/godotenv"
)

func main() {

	fmt.Println("# Service running")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("# Failed to load configuration.[Node]\n", err)
	}

	go lib.TransactionCosumer()
	go lib.BlockConsumer()

	select {}
}
