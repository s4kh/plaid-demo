package main

import (
	"fmt"
	"net/http"
)

func (s *server) routes() {
	s.router.HandleFunc("GET", "/api/liabilities/:user", s.handleLiabilities())
	s.router.HandleFunc("GET", "/api/institutions/:q", s.handleSearch())

	s.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "This is not the page you are looking for")
	})
}
