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
	Id        string `json:"id,omitempty"`
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

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	sender := r.URL.Query().Get("sender")
	recipient := r.URL.Query().Get("recipient")
	lastMessageId := r.URL.Query().Get("prev_page")

	var messages []message
	query := h.db.Session.Query(`SELECT * FROM messages WHERE sender = ? AND recipient = ? ORDER BY id DESC LIMIT 50`, sender, recipient)
	if len(lastMessageId) > 0 {
		query = h.db.Session.Query(`SELECT * FROM messages WHERE sender = ? AND recipient = ? AND id > ? ORDER BY id DESC LIMIT 50`, sender, recipient, lastMessageId)
	}
	iter := query.Iter()
	for {
		var msg message
		if !iter.Scan(&msg.Id, &msg.Sender, &msg.Recipient, &msg.Content, &msg.Timestamp) {
			break
		}
		messages = append(messages, msg)
	}
	if err := iter.Close(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string][]message{
		"messages": messages,
	})
	if err != nil {
		log.Printf("Error encoding messages: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
