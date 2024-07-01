package main

import (
	"fmt"
	"github.com/SamirMohamed/cme-chatting/cmd/authentication"
	"github.com/SamirMohamed/cme-chatting/cmd/chatting"
	jwtAuth "github.com/SamirMohamed/cme-chatting/pkg/authentication"
	"github.com/SamirMohamed/cme-chatting/pkg/cache"
	"github.com/SamirMohamed/cme-chatting/pkg/datastore"
	"log"
	"net/http"
	"os"
	"strconv"
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
		log.Fatalf("Error casting Redis post to integer: %v", err)
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

	// Handle Routes
	authHandler := authentication.NewAuthenticationHandler(db)
	chattingHandler := chatting.NewChattingHandler(db)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", healthCheckHandler)
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authMiddleware(authHandler.Login))
	mux.HandleFunc("/send", authMiddleware(chattingHandler.Send))

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

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		_, err := jwtAuth.NewJwtAuthenticator().VerifyJWT(tokenString)
		if err != nil {
			log.Printf("Error verifying jwt token: %v", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
