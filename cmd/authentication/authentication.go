package authentication

import (
	"encoding/json"
	"github.com/SamirMohamed/cme-chatting/pkg/datastore"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type user struct {
	Username string
	Password string
}

type Handler struct {
	db *datastore.Cassandra
}

func NewAuthenticationHandler(db *datastore.Cassandra) Handler {
	return Handler{
		db: db,
	}
}

func (auth Handler) Register(w http.ResponseWriter, r *http.Request) {
	var u user
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = auth.db.Session.Query(`INSERT INTO users (username, password) VALUES (?, ?)`, u.Username, password).Exec()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
