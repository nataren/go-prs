package main

import (
	"encoding/json"
	"github.com/google/go-github/github"
	"log"
	"net/http"
	"time"
)

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

type PullRequestHandler struct{}

const (
	MT_PR_CLOSED                   = iota
	MT_PR_TARGETS_MASTER_BRANCH    = iota
	MT_PR_OPENED_NOT_IN_YOUTRACK   = iota
	MT_PR_REOPENED_NOT_IN_YOUTRACK = iota
	MT_PR_MERGED                   = iota
	MT_PR_UNKNOWN_MERGEABILITY     = iota
	MT_PR_AUTO_MERGEABLE           = iota
	MT_PR_UNCATEGORIZED            = iota
)

type MindTouchPullRequest struct {
	Type int
}

func GetMindTouchPullRequestTypeFromEvent(prEvent github.PullRequestEvent) MindTouchPullRequest {
	if *prEvent.Action == "closed" {
		return MindTouchPullRequest{MT_PR_CLOSED}
	} else {
		return MindTouchPullRequest{GetMindTouchPullRequestType(prEvent.PullRequest)}
	}
}

func GetMindTouchPullRequestType(pr *github.PullRequest) int {
	return MT_PR_UNCATEGORIZED
}

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
			if jsonDecodingErr != nil {
				w.Write([]byte("Could not decode PullRequestEvent"))
			} else {
				mtPr := GetMindTouchPullRequestTypeFromEvent(prEvent)
				marshalResp, _ := json.Marshal(mtPr.Type)
				w.Write(marshalResp)
			}
		} else if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}
}
