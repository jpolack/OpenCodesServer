package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/iotaledger/giota"
)

func Start() {
	iotaConnector := initConnector()
	handler := httpHandler{iotaConnector}
	fmt.Println("Connected")

	r := chi.NewRouter()
	r.Post("/new", handler.newCapsuleHandler)
	fmt.Println("Listening on Port :8000")
	http.ListenAndServe(":8000", r)
}

type httpHandler struct {
	iotaConnector connector
}

type capsule struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	From     string `json:"from"`
	To       string `json:"to"`
}

type createCapsule struct {
	Capsule     capsule   `json:"capsule"`
	OpeningDate time.Time `json:"openingDate"`
	Password    string    `json:"password"`
}

func (h *httpHandler) newCapsuleHandler(w http.ResponseWriter, r *http.Request) {
	inputCapsule := createCapsule{}
	err := json.NewDecoder(r.Body).Decode(&inputCapsule)
	if err != nil {
		log.Fatalln("Error parsing JSON", err)
		w.Write([]byte("Error"))
		return
	}

	seed := giota.NewSeed()
	address, err := giota.NewAddress(seed, 0, 3)
	if err != nil {
		log.Fatalln("Could not generate address", err)
		w.Write([]byte("Error"))
		return
	}

	message, err := json.Marshal(inputCapsule.Capsule)
	if err != nil {
		log.Fatalln("Could not marshal capsule", inputCapsule, err)
	}

	encryptedMessage, err := encrypt([]byte(inputCapsule.Password), string(message))
	if err != nil {
		log.Fatalln("Could not encrypt capsule", inputCapsule, err)
	}

	h.iotaConnector.newCapsule(encryptedMessage, string(address))

	var response struct {
		Link string `json:"link"`
	}

	response.Link = "http://localhost:3000/capsule/" + string(address)

	bytes, err := json.Marshal(response)
	if err != nil {
		log.Fatalln("Error stringifing JSON", err)
		w.Write([]byte("Error"))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}
