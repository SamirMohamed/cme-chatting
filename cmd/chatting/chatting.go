package chatting

import (
	"encoding/json"
	"fmt"
	"github.com/SamirMohamed/cme-chatting/pkg/cache"
	"github.com/SamirMohamed/cme-chatting/pkg/datastore"
	"github.com/gocql/gocql"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	db    *datastore.Cassandra
	cache cache.Cache
}

type message struct {
	Id        string    `json:"id,omitempty"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func NewChattingHandler(db *datastore.Cassandra, cache cache.Cache) *Handler {
	return &Handler{
		db:    db,
		cache: cache,
	}
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	var msg message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Printf("Error parsing send data: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.db.Session.Query(`INSERT INTO messages (id, sender, recipient, content, timestamp) VALUES (?, ?, ?, ?, ?)`,
		gocql.TimeUUID(), msg.Sender, msg.Recipient, msg.Content, time.Now()).Exec()
	if err != nil {
		log.Printf("Error while insert message data: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cacheKey := fmt.Sprintf("%s:%s:*", msg.Sender, msg.Recipient)
	err = h.cache.Del(cacheKey)
	if err != nil {
		log.Printf("Error deleting cached messages: %v\n", err)
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	sender := r.URL.Query().Get("sender")
	recipient := r.URL.Query().Get("recipient")
	lastMessageId := r.URL.Query().Get("prev_page")

	messages := []message{}

	cacheKey := fmt.Sprintf("%s:%s:%s", sender, recipient, lastMessageId)
	cachedMessages, err := h.cache.LRange(cacheKey)
	if err == nil && len(cachedMessages) > 0 {
		for _, cachedMsg := range cachedMessages {
			var msg message
			json.Unmarshal([]byte(cachedMsg), &msg)
			messages = append(messages, msg)
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string][]message{
			"messages": messages,
		})
		if err != nil {
			log.Printf("Error encoding messages: %v\n", err)
		} else {
			return
		}
	} else {
		log.Printf("Error retrieving messages from cache: %v\n", err)
	}

	var cacheMessages []interface{}

	query := h.db.Session.Query(`SELECT id, sender, recipient, content, timestamp FROM messages WHERE sender = ? AND recipient = ? ORDER BY id DESC LIMIT 50`, sender, recipient)
	if len(lastMessageId) > 0 {
		query = h.db.Session.Query(`SELECT id, sender, recipient, content, timestamp FROM messages WHERE sender = ? AND recipient = ? AND id > ? ORDER BY id DESC LIMIT 50`, sender, recipient, lastMessageId)
	}
	iter := query.Iter()
	for {
		var msg message
		if !iter.Scan(&msg.Id, &msg.Sender, &msg.Recipient, &msg.Content, &msg.Timestamp) {
			break
		}
		messages = append(messages, msg)
		b, _ := json.Marshal(msg)
		cacheMessages = append(cacheMessages, string(b))
	}
	if err := iter.Close(); err != nil {
		log.Printf("Error closing iterator: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(cacheMessages) > 0 {
		err = h.cache.RPush(cacheKey, cacheMessages...)
		if err != nil {
			log.Printf("Error caching messages: %v\n", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string][]message{
		"messages": messages,
	})
	if err != nil {
		log.Printf("Error encoding messages: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
