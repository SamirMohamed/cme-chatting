package authentication

import (
	"encoding/json"
	"github.com/SamirMohamed/cme-chatting/pkg/authentication"
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
	db            *datastore.Cassandra
	authenticator *authentication.Jwt
}

func NewAuthenticationHandler(db *datastore.Cassandra) *Handler {
	return &Handler{
		db:            db,
		authenticator: authentication.NewJwtAuthenticator(),
	}
}

func (auth *Handler) Register(w http.ResponseWriter, r *http.Request) {
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

	err = auth.db.Session.Query(`INSERT INTO users (username, password) VALUES (?, ?)`, u.Username, password).Exec()
	if err != nil {
		log.Printf("Error while insert registration data: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (auth *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var u user
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var password string
	err = auth.db.Session.Query(`SELECT password FROM users WHERE username = ? LIMIT 1`, u.Username).Scan(&password)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(u.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.authenticator.GenerateJWT(u.Username)
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
