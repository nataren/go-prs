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

type MindTouchPullRequestType int

const (
	MT_PR_CLOSED MindTouchPullRequestType = iota
	MT_PR_TARGETS_MASTER_BRANCH
	MT_PR_OPENED_NOT_IN_YOUTRACK
	MT_PR_REOPENED_NOT_IN_YOUTRACK
	MT_PR_MERGED
	MT_PR_UNKNOWN_MERGEABILITY
	MT_PR_AUTO_MERGEABLE
	MT_PR_UNCATEGORIZED
)

type MindTouchPullRequest struct {
	Type MindTouchPullRequestType
}

func GetMindTouchPullRequestTypeFromEvent(prEvent github.PullRequestEvent) MindTouchPullRequest {
	if *prEvent.Action == "closed" {
		return MindTouchPullRequest{MT_PR_CLOSED}
	} else {
		return MindTouchPullRequest{GetMindTouchPullRequestType(prEvent.PullRequest)}
	}
}

func IsPullRequestMergeable(pr *github.PullRequest) bool {
	return pr.Merged != nil && !*pr.Merged && pr.Mergeable != nil && *pr.Mergeable
	//	&& pr.MergeableState != nil
	//  && *pr.MergeableState == "clean"
}

// TODO: Return a tuple (Time, err) instead
func GetBranchDate(branch *string) time.Time {
	dateFormat := "yyyyMMdd"
	return time.Parse(dateFormat, (*branch)[len(*branch)-len(dateFormat)])
}

func PullRequestTargetsOpenBranch(pr *github.PullRequest) bool {
	return GetBranchDate(pr.Base.Ref)-time.Now().UTC() >= (time.Hour * 138)
}

func IsAutoMergeablePullRequest(pr *github.PullRequest) bool {
	return IsPullRequestMergeable(pr) && PullRequestTargetsOpenBranch(pr)
}

func GetMindTouchPullRequestType(pr *github.PullRequest) MindTouchPullRequestType {
	prState := *pr.State
	switch *pr.State {
	case "closed":
		return MT_PR_CLOSED
	case "open":
		if IsAutoMergeablePullRequest(pr) {
			return MT_PR_AUTO_MERGEABLE
		}
	}
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
