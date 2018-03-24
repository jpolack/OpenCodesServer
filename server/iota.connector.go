package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	gmam "github.com/habpygo/mam.client.go"
	"github.com/habpygo/mam.client.go/mamutils"
	"github.com/iotaledger/giota"
)

const SEED = "ENBRKGYXMKWNSNECKCBYJWODEEBHPCCYMRURXOVOMZVB99HNNCYMUXRDEDKHCPZRFBNYPQIDAPCEAHOQW"

type iotaConnector interface {
	SendToApi([]giota.Transfer) (giota.Bundle, error)
	FindTransactions(req giota.FindTransactionsRequest) ([]giota.Transaction, error)
}

func initConnector() connector {
	c, err := gmam.NewFakeConnection("http://node.lukaseder.de:14265", SEED)
	// c, err := gmam.NewConnection("http://node.lukaseder.de:14265", SEED)
	if err != nil {
		log.Fatalln("Error building connection", err)
		return connector{}
	}
	return connector{c}
}

type saveCapsule struct {
	Data        string    `json:"data"`
	OpeningDate time.Time `json:"openingDate"`
}

type connector struct {
	conn iotaConnector
}

func (c *connector) newCapsule(message, address string, openingDate time.Time) {
	if c.conn == nil {
		log.Fatalln("Connection is required")
		return
	}

	bytes, err := json.Marshal(saveCapsule{message, openingDate})
	if err != nil {
		log.Fatalln("Could marshal capsule", message, openingDate, err)
		return
	}

	_, err = gmam.Send(address, 0, string(bytes), c.conn)
	if err != nil {
		log.Fatalln("Could not send capsule", message, err)
		return
	}

	fmt.Printf("Sent Transaction to: %v\n", address)
}

type readCapsule struct {
	meta saveCapsule
}

func (c *connector) readCapsule(address string) (readCapsule, error) {
	iotaAdress, err := giota.ToAddress(address)
	if err != nil {
		return readCapsule{}, errors.New("Invalid Address")
	}

	ts, err := c.conn.FindTransactions(giota.FindTransactionsRequest{
		Addresses: []giota.Address{iotaAdress},
	})

	if err != nil {
		return readCapsule{}, errors.New("Address not found")
	}

	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Timestamp.UnixNano() < ts[j].Timestamp.UnixNano()
	})

	metaEncoded := ts[0]
	// rest := ts[1:]

	metaString, err := mamutils.FromMAMTrytes(metaEncoded.SignatureMessageFragment)
	if err != nil {
		log.Fatalln("Could parse trytes", err)
		return readCapsule{}, errors.New("Internal Error")
	}

	meta := saveCapsule{}
	err = json.Unmarshal([]byte(metaString), &meta)

	if err != nil {
		log.Fatalln("Could not unmarshal", err)
		return readCapsule{}, errors.New("Internal Error")
	}

	return readCapsule{meta}, nil
}
