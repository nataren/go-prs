package main

import (
	// "bytes"
	// "encoding/binary"
	"log"
	"net/http"
	"time"
	"encoding/json"
	"github.com/google/go-github/github"
)

func main() {
	s := &http.Server {
		Addr:           ":8080",
		Handler:        new(PullRequestHandler),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}

type PullRequestHandler struct{}

func (h *PullRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check the VERB
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Check the payload
    if r.ContentLength >= 0 {
		resp := make([]byte, r.ContentLength)
		bytesRead, err := r.Body.Read(resp)
		if bytesRead > 0 {
			var prEvent github.PullRequestEvent
			jsonDecodingErr := json.Unmarshal(resp, &prEvent)
			if(jsonDecodingErr != nil) {
				w.Write([]byte("Could not decode PullRequestEvent"))
			} else {
				marshalResp, _ := json.Marshal(*prEvent.Number)
				w.Write(marshalResp)
			}
		} else if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}
}