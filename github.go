package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GitHub struct {
	Token string
}

func (g *GitHub) SetupRequest(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
}

type User struct {
	Name string `json:"login"`
	URL string `json:"html_url"`
}

type PullRequest struct {
	Number int `json:"number"`
	Title string `json:"title"`
	Body string `json:"body"`
	URL string `json:"html_url"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	ClosedAt *time.Time `json:"closed_at"`
	MergedAt *time.Time `json:"merged_at"`
	IsDraft bool `json:"draft"`
	Author User `json:"user"`
	Assignees []User `json:"assignees"`
	Reviewers []User `json:"requested_reviewers"`
}

func (pr *PullRequest) TimeOpen(time time.Time) time.Duration {
	created := pr.CreatedAt != nil
	if !created {
		// How?
		return 0
	}

	return time.Sub(*pr.CreatedAt)
}

func (pr *PullRequest) DaysOpen(time time.Time) int {
	hrs := pr.TimeOpen(time).Hours()
	return int(hrs / 24)
}

// TODO Add ExpiryStatus enum to handle Ok/Stale/Expired

func (pr *PullRequest) IsStale(time time.Time) bool {
	return pr.DaysOpen(time) >= 14
}

func (pr *PullRequest) IsExpired(time time.Time) bool {
	return pr.DaysOpen(time) >= 30
}

type StampedPullRequest struct {
	PullRequest
	Time time.Time
	TimeOpen time.Duration
	DaysOpen int
	IsStale bool
	IsExpired bool
}

func (pr *PullRequest) Stamp(time time.Time) StampedPullRequest {
	return StampedPullRequest{
		PullRequest: *pr,
		Time: time,
		TimeOpen: pr.TimeOpen(time),
		DaysOpen: pr.DaysOpen(time),
		IsStale: pr.IsStale(time),
		IsExpired: pr.IsExpired(time),
	}
}

func (g *GitHub) GetPullRequestsJson(owner, repo string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	g.SetupRequest(req)

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

func (g *GitHub) GetPullRequests(owner, repo string) ([]PullRequest, error) {
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

func (g *GitHub) GetStampedPullRequests(owner, repo string) ([]StampedPullRequest, error) {
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
