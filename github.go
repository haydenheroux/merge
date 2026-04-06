package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GitHubClient interface {
	GetPullRequestsJson(owner, repo string) ([]byte, error)
	GetPullRequests(owner, repo string) ([]PullRequest, error)
	GetStampedPullRequests(owner, repo string) ([]StampedPullRequest, error)
}

type GitHub struct {
	Token string
}

func (g GitHub) setupRequest(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
}

func (g GitHub) GetPullRequestsJson(owner, repo string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	g.setupRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (g GitHub) GetPullRequests(owner, repo string) ([]PullRequest, error) {
	bytes, err := g.GetPullRequestsJson(owner, repo)
	if err != nil {
		return nil, err
	}

	var prs []PullRequest

	if err := json.Unmarshal(bytes, &prs); err != nil {
		return nil, err
	}

	return prs, nil
}

func (g GitHub) GetStampedPullRequests(owner, repo string) ([]StampedPullRequest, error) {
	prs, err := g.GetPullRequests(owner, repo)
	if err != nil {
		return nil, err
	}

	stamped := make([]StampedPullRequest, len(prs))
	now := time.Now()
	for i, pr := range prs {
		stamped[i] = pr.Stamp(now)
	}
	return stamped, nil
}
