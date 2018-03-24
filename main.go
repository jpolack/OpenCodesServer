package main

import (
	"fmt"
	"time"

	gmam "github.com/habpygo/mam.client.go"
)

const SEED = "ENBRKGYXMKWNSNECKCBYJWODEEBHPCCYMRURXOVOMZVB99HNNCYMUXRDEDKHCPZRFBNYPQIDAPCEAHOQW"
const address = "KDFOXSUPVNEDGHTCLFJTOJIZFPNZHTHXUGCEGSUENLFKTFGRGNEE9UNFFUKMMMSHYJYONJMOWUP9RNVRBWJHFPWFSZ"

func main() {

	c, err := gmam.NewConnection("http://node.lukaseder.de:14265", SEED)
	if c != nil && err == nil {
		fmt.Println("Connection is valid")
	}

	msgTime := time.Now().UTC().String()
	message := "It's the most wonderful message of the year ;-) on: " + msgTime

	id, err := gmam.Send(address, 0, message, c)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Sent Transaction: %v\n", id)
}

// func main() {
// 	r := chi.NewRouter()
// 	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("welcome"))
// 	})
// 	http.ListenAndServe(":8000", r)
// }
