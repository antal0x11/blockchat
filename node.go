package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/antal0x11/blockchat/dst"
	"github.com/antal0x11/blockchat/lib"
)

func main() {

	fmt.Println("# Service running")

	//uncomment for dev mode
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("# [Node] Failed to load configuration.\n", err)
	// }

	wallet := lib.GenerateWallet()
	neighboors := dst.Neighboors{}

	fmt.Println("# [Node] Public/Private key generated.")

	bootstrapNode, err := strconv.ParseBool(os.Getenv("BOOTSTRAP"))
	if err != nil {
		log.Fatal("# [Node] Failed to load configuration.\n")
	}

	node := dst.Node{
		BootStrap: bootstrapNode,
		Nonce:     0,
		Stake:     20,
		PublicKey: wallet.PublicKeyToString(),
		Balance:   100,
		Validator: "",
	}

	if node.BootStrap {

		node.Id = 0

		_n, err := strconv.ParseFloat(os.Getenv("NEIGHBOORS"), 64)
		if err != nil {
			log.Fatal("# [Node] Failed to load configuration\n")
		}
		initNeighboor := dst.NeighboorNode{
			Id:        node.Id,
			BootStrap: true,
			PublicKey: node.PublicKey,
			Balance:   _n * 1000,
		}

		neighboors.DSNodes = append(neighboors.DSNodes, initNeighboor)
		init := make(chan *dst.Neighboors)

		lib.BootStrapBlockInitialize(&node, wallet)
		go lib.BoostrapInformationConsumer(&neighboors, init)
		go lib.BootstrapInitTransactionAndBlockChain(init, &node, wallet)
	} else {
		go lib.NodeInformationConsumer(&neighboors, &node)
		go lib.NodeInformationPublisher(&node)
	}

	go lib.TransactionConsumer(&node, &neighboors, wallet)
	go lib.BlockConsumer(&node)

	select {}

}
