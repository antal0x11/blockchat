package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/antal0x11/blockchat/dst"
)

func NodeHttpService(node *dst.Node) {

	fmt.Println("# [NodeHttpService] HttpService is running.")
	go func(node *dst.Node) {
		http.HandleFunc("/", nodeInfo(node))

		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			log.Fatal("# [NodeHttpService] HttpService failed.")
		}
	}(node)

}

func nodeInfo(node *dst.Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		node.Mu.Lock()

		_res, err := json.Marshal(node)
		if err != nil {
			log.Fatal("# [NodeHttpService] Failed to Serialize node")
		}

		node.Mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, string(_res[:]))
	}
}
