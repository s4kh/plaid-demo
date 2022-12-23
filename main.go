package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/matryer/way"
	"github.com/plaid/plaid-go/v10/plaid"
)

const (
	PLAID_CLIENT_ID = "PLAID-CLIENT-ID"
	PLAID_SECRET    = "PLAID-SECRET"
	PORT            = ":8080"
)

type server struct {
	//db *someDatabase
	router *way.Router
	client *plaid.APIClient
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) handleLiabilities() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		userId := way.Param(r.Context(), "user")
		log.Printf("getting liabilities of user:%s", userId)
		accs, err := GetLiabilities(s.client, r.Context(), []string{})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(accs)
	}
}

func (s *server) handleSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := way.Param(r.Context(), "q")
		log.Printf("searching for :%s", query)

		inss, err := SearchInstitutions(s.client, r.Context(), query)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "internal server error")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(inss)
	}
}

func newServer() *server {
	s := &server{}
	s.router = way.NewRouter()
	s.routes()
	s.client = GetClient()

	return s
}

func run() error {
	err := godotenv.Load()
	server := newServer()

	if err != nil {
		return fmt.Errorf("error loading .env file")
	}

	log.Println("Listening on ", PORT)
	log.Fatal(http.ListenAndServe(PORT, server))

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}
