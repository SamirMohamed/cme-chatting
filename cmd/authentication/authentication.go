package authentication

import (
	"encoding/json"
	authenticator "github.com/SamirMohamed/cme-chatting/pkg/authentication"
	"github.com/SamirMohamed/cme-chatting/pkg/datastore"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

type user struct {
	Username string
	Password string
}

type Handler struct {
	db   *datastore.Cassandra
	auth authenticator.Authenticator
}

func NewAuthenticationHandler(db *datastore.Cassandra, authenticator authenticator.Authenticator) *Handler {
	return &Handler{
		db:   db,
		auth: authenticator,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var u user
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating hashed password: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = h.db.Session.Query(`INSERT INTO users (username, password) VALUES (?, ?)`, u.Username, password).Exec()
	if err != nil {
		log.Printf("Error while insert registration data: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var u user
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var password string
	err = h.db.Session.Query(`SELECT password FROM users WHERE username = ? LIMIT 1`, u.Username).Scan(&password)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(u.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := h.auth.Generate(u.Username)
	if err != nil {
		log.Printf("Error generating JWT token: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
	if err != nil {
		log.Printf("Error encoding JWT token: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
