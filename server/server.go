package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/iotaledger/giota"
)

func Start() {
	iotaConnector := initConnector()
	handler := httpHandler{iotaConnector}
	fmt.Println("Connected")

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	r := chi.NewRouter()
	r.Use(cors.Handler)
	r.Post("/new", handler.newCapsuleHandler)
	r.Post("/capsule/{id}", handler.readCapsuleHandler)
	r.Put("/capsule/{id}", handler.writeCapsuleHandler)
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
	log.Printf("Create new Capsule\n")

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

	h.iotaConnector.newCapsule(string(message), string(address), inputCapsule.OpeningDate)

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

type feed struct {
	Meta        capsule               `json:"capsule"`
	OpeningDate time.Time             `json:"openingDate"`
	Memories    []memoryWithTimestamp `json:"memories"`
}

type memoryWithTimestamp struct {
	memory
	CreationDate time.Time `json:"creationDate"`
}

func (h *httpHandler) readCapsuleHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	readCapsule, err := h.iotaConnector.readCapsule(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ungültige Zeitkapsel"))
		return
	}

	var request struct {
		Password string `json:"password"`
	}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatalln("Error parsing JSON", err)
		w.Write([]byte("Error"))
		return
	}

	response := feed{}

	err = json.Unmarshal([]byte(readCapsule.meta.Data), &response.Meta)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ungültiges Passwort"))
		return
	}

	response.Memories = readCapsule.memories
	response.OpeningDate = readCapsule.meta.OpeningDate

	bytes, err := json.Marshal(response)
	if err != nil {
		log.Fatalln("Error stringifing JSON", err)
		w.Write([]byte("Error"))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

type memory struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func (h *httpHandler) writeCapsuleHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	readCapsule, err := h.iotaConnector.readCapsule(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ungültige Zeitkapsel"))
		return
	}

	var request struct {
		Memory   memory `json:"memory"`
		Password string `json:"password"`
	}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatalln("Error parsing JSON", err)
		w.Write([]byte("Error"))
		return
	}

	response := feed{}

	err = json.Unmarshal([]byte(readCapsule.meta.Data), &response.Meta)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ungültiges Passwort"))
		return
	}

	h.iotaConnector.writeToCapsule(request.Memory, id)

	w.Write([]byte("OK"))
}
