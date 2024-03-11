package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/antal0x11/blockchat/dst"
)

func NodeHttpService(node *dst.Node, neighboors *dst.Neighboors) {

	fmt.Println("# [NodeHttpService] HttpService is running.")
	go func(node *dst.Node, neighboors *dst.Neighboors) {
		http.HandleFunc("/", nodeInfo(node))
		http.HandleFunc("/api/neighboors", neighboorsHttpService(neighboors))

		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			log.Fatal("# [NodeHttpService] HttpService failed.")
		}
	}(node, neighboors)

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

		_res, err := json.Marshal(neighboors)
		if err != nil {
			log.Fatal("# [NeighboorsHttpService] Failed to serialize neighboors.")
		}

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, string(_res[:]))
	}
}
