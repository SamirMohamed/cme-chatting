package chatting

import (
	"context"
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
	cache *cache.Redis
}

type message struct {
	Id        string `json:"id,omitempty"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
}

func NewChattingHandler(db *datastore.Cassandra, cache *cache.Redis) *Handler {
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
	err = h.cache.Client.Del(context.Background(), cacheKey).Err()
	if err != nil {
		log.Printf("Error deleting cached messages: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	sender := r.URL.Query().Get("sender")
	recipient := r.URL.Query().Get("recipient")
	lastMessageId := r.URL.Query().Get("prev_page")

	cacheKey := fmt.Sprintf("%s:%s:%s", sender, recipient, lastMessageId)
	cachedMessages, err := h.cache.Client.LRange(context.Background(), cacheKey, 0, -1).Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string][]string{
			"messages": cachedMessages,
		})
		if err != nil {
			log.Printf("Error encoding messages: %v\n", err)
			return
		}
		return
	} else {
		log.Printf("Error retrieving messages from cache: %v\n", err)
	}

	var messages []message
	var cacheMessages []interface{}

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
		cacheMessages = append(cacheMessages, msg)
	}
	if err := iter.Close(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = h.cache.Client.RPush(context.Background(), cacheKey, cacheMessages...).Err()
	if err != nil {
		log.Printf("Error caching messages: %v\n", err)
	}

	err = h.cache.Client.ExpireNX(context.Background(), cacheKey, time.Hour*1).Err()
	if err != nil {
		log.Printf("Error setting caching expire: %v\n", err)
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
