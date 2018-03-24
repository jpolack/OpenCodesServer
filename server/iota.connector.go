package server

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	gmam "github.com/habpygo/mam.client.go"
)

const SEED = "ENBRKGYXMKWNSNECKCBYJWODEEBHPCCYMRURXOVOMZVB99HNNCYMUXRDEDKHCPZRFBNYPQIDAPCEAHOQW"

func initConnector() connector {
	c, err := gmam.NewConnection("http://node.lukaseder.de:14265", SEED)
	if err != nil {
		log.Fatalln("Error building connection", err)
		return connector{}
	}
	return connector{c}
}

type safeCapsule struct {
	Data        string    `json:"data"`
	OpeningDate time.Time `json:"openingDate"`
}

type connector struct {
	conn *gmam.Connection
}

func (c *connector) newCapsule(message, address string, openingDate time.Time) {
	if c.conn == nil {
		log.Fatalln("Connection is required")
		return
	}

	bytes, err := json.Marshal(safeCapsule{message, openingDate})
	if err != nil {
		log.Fatalln("Could marshal capsule", message, openingDate, err)
	}

	_, err = gmam.Send(address, 0, string(bytes), c.conn)
	if err != nil {
		log.Fatalln("Could not send capsule", message, err)
	}

	fmt.Printf("Sent Transaction to: %v\n", address)
}
