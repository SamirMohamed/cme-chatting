package chatting

import (
	"encoding/json"
	"github.com/SamirMohamed/cme-chatting/pkg/datastore"
	"github.com/gocql/gocql"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	db *datastore.Cassandra
}

type message struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

func NewChattingHandler(db *datastore.Cassandra) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	var msg message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Printf("Error parsing send data: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.db.Session.Query(`INSERT INTO messages (id, sender, recipient, content, timestamp) VALUES (?, ?, ?, ?)`,
		gocql.TimeUUID(), msg.Sender, msg.Recipient, msg.Content, time.Now()).Exec()
	if err != nil {
		log.Printf("Error while insert message data: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
