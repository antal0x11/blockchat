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

	// Uncomment for dev mode
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("# [Node] Failed to load configuration.\n", err)
	// }

	wallet := lib.GenerateWallet()

	neighboors := dst.Neighboors{
		DSNodes: make(map[uint32]*dst.NeighboorNode),
	}

	// Maps Public Key with node id
	_mapNodeId := make(map[string]uint32)

	fmt.Println("# [Node] Public/Private key generated.")

	bootstrapNode, err := strconv.ParseBool(os.Getenv("BOOTSTRAP"))
	if err != nil {
		log.Fatal("# [Node] Failed to load configuration.\n")
	}

	_stake, err := strconv.ParseFloat(os.Getenv("STAKE"), 64)
	if err != nil {
		log.Fatal("# [Node] Failed to load configuration.\n")
	}

	node := dst.Node{
		BootStrap: bootstrapNode,
		Nonce:     0,
		Stake:     _stake,
		PublicKey: wallet.PublicKeyToString(),
		Balance:   1000,
		Validator: "",
	}

	if node.BootStrap {

		node.Id = 0

		initNeighboor := dst.NeighboorNode{
			Id:        node.Id,
			BootStrap: true,
			PublicKey: node.PublicKey,
			Balance:   1000,
			Stake:     20,
		}

		neighboors.DSNodes[0] = &initNeighboor
		_mapNodeId[node.PublicKey] = node.Id

		init := make(chan *dst.Neighboors)

		lib.BootStrapBlockInitialize(&node, wallet)
		go lib.BoostrapInformationConsumer(&neighboors, init, _mapNodeId)
		go lib.BootstrapInitTransactionAndBlockChain(init, &node, wallet)
	} else {
		go lib.NodeInformationConsumer(&neighboors, &node, _mapNodeId)
		go lib.NodeInformationPublisher(&node)
	}

	go lib.TransactionConsumer(&node, &neighboors, wallet, _mapNodeId)
	go lib.BlockConsumer(&node, &neighboors, _mapNodeId)
	go lib.NodeHttpService(&node, &neighboors, &wallet)

	select {}

}
