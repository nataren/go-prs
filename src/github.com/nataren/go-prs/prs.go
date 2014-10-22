package main

import (
	"log"
	"net/http"
	"time"
)

type PullRequestHandler struct{}

func (h *PullRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("go-prs alive ..."))
}

func main() {
	s := &http.Server{
		Addr:           ":8080",
		Handler:        new(PullRequestHandler),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
