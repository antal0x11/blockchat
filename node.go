package main

import (
	"fmt"
	"log"

	"github.com/antal0x11/blockchat/dst"
	"github.com/antal0x11/blockchat/lib"
	"github.com/joho/godotenv"
)

func main() {

	fmt.Println("# Service running")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("# [Node] Failed to load configuration.[Node]\n", err)
	}

	wallet := lib.GenerateWallet()

	fmt.Println("# [Node] Public/Private key generated.")

	node := dst.Node{
		Id:        1,
		BootStrap: true,
		Nonce:     0,
		Stake:     20,
		PublicKey: wallet.PublicKeyToString(),
		Balance:   100,
		Validator: "",
	}

	if node.BootStrap {
		lib.BootStrapBlockInitialize(&node, wallet)
	}

	go lib.TransactionCosumer(&node, wallet)
	go lib.BlockConsumer(&node)

	select {}

}
