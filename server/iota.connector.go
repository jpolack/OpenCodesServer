package server

import (
	"encoding/json"
	"errors"
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

	log.Printf("Created new capsule: %v\n", address)
}

type readCapsule struct {
	meta     saveCapsule
	memories []memory
}

func (c *connector) readCapsule(address string) (readCapsule, error) {
	if c.conn == nil {
		log.Fatalln("Connection is required")
		return readCapsule{}, nil
	}

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

	if len(ts) == 0 {
		return readCapsule{}, errors.New("Address not found")
	}

	metaEncoded := ts[0]
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

	restEncoded := ts[1:]
	memories := []memory{}

	for _, m := range restEncoded {
		memoryString, err := mamutils.FromMAMTrytes(m.SignatureMessageFragment)
		if err != nil {
			log.Fatalln("Could parse trytes", err)
			continue
		}

		memo := memory{}
		err = json.Unmarshal([]byte(memoryString), &memo)

		if err != nil {
			log.Fatalln("Could not unmarshal", err)
			continue
		}

		memories = append(memories, memo)
	}

	return readCapsule{meta, memories}, nil
}

func (c *connector) writeToCapsule(message memory, address string) {
	if c.conn == nil {
		log.Fatalln("Connection is required")
		return
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		log.Fatalln("Could marshal message", message, err)
		return
	}

	_, err = gmam.Send(address, 0, string(bytes), c.conn)
	if err != nil {
		log.Fatalln("Could not send message", message, err)
		return
	}

	log.Printf("Sent memory to: %v\n", address)
}
