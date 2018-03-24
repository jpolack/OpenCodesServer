package server

import (
	"fmt"
	"log"

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

type connector struct {
	conn *gmam.Connection
}

func (c *connector) newCapsule(message, address string) {
	if c.conn == nil {
		log.Fatalln("Connection is required")
		return
	}

	_, err := gmam.Send(address, 0, message, c.conn)
	if err != nil {
		log.Fatalln("Could not send capsule", message, err)
	}

	fmt.Printf("Sent Transaction to: %v\n", address)
}
