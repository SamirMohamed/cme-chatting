package main

import (
	"fmt"
	"github.com/SamirMohamed/cme-chatting/cmd/authentication"
	"github.com/SamirMohamed/cme-chatting/cmd/chatting"
	authenticator "github.com/SamirMohamed/cme-chatting/pkg/authentication"
	"github.com/SamirMohamed/cme-chatting/pkg/cache"
	"github.com/SamirMohamed/cme-chatting/pkg/datastore"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Init Cassandra
	cAddresses := []string{os.Getenv("CASSANDRA_HOST")}
	cKeyspace := os.Getenv("CASSANDRA_KEYSPACE")
	cUsername := os.Getenv("CASSANDRA_USERNAME")
	cPassword := os.Getenv("CASSANDRA_PASSWORD")
	db, err := datastore.NewCassandra(cAddresses, cKeyspace, cUsername, cPassword)
	if err != nil {
		log.Fatalf("Error connecting to Cassandra: %v", err)
	}
	defer db.Close()

	// Init Redis
	rAddress := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	rDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Error casting Redis db to integer: %v", err)
	}
	c, err := cache.NewRedis(rAddress, rDB)
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	defer func(c *cache.Redis) {
		err := c.Close()
		if err != nil {
			log.Fatalf("Error closing Redis connectino: %v", err)
		}
	}(c)

	// Init authenticator
	auth := authenticator.NewJwtAuthenticator()

	// Handle Routes
	authHandler := authentication.NewAuthenticationHandler(db, auth)
	chattingHandler := chatting.NewChattingHandler(db, c)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", healthCheckHandler)
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/send", authMiddleware(chattingHandler.Send, auth))
	mux.HandleFunc("/messages", authMiddleware(chattingHandler.GetMessages, auth))

	// Init server
	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", recoverMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{\"Status\":\"ok\"}")
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recovered from panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.HandlerFunc, auth authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer")
		if len(splitToken) != 2 {
			log.Printf("Error verifying jwt token: %v", fmt.Errorf("missing jwt token"))
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		tokenString := strings.TrimSpace(splitToken[1])

		err := auth.Verify(tokenString)
		if err != nil {
			log.Printf("Error verifying jwt token: %v", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
