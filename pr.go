package main

import "time"

type User struct {
	Name string `json:"login"`
	URL  string `json:"html_url"`
}

type PullRequest struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	URL       string     `json:"html_url"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	MergedAt  *time.Time `json:"merged_at"`
	IsDraft   bool       `json:"draft"`
	Author    User       `json:"user"`
	Assignees []User     `json:"assignees"`
	Reviewers []User     `json:"requested_reviewers"`
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

type ExpiryStatus string

const (
	Fresh   ExpiryStatus = "New"
	Stale                = "Stale"
	Expired              = "Expired"
)

func (pr *PullRequest) ExpiryStatus(time time.Time) ExpiryStatus {
	days := pr.DaysOpen(time)
	if days >= 30 {
		return Expired
	}
	if days >= 14 {
		return Stale
	}
	return Fresh
}

type StampedPullRequest struct {
	PullRequest
	Time         time.Time
	TimeOpen     time.Duration
	DaysOpen     int
	ExpiryStatus ExpiryStatus
}

func (pr *PullRequest) Stamp(time time.Time) StampedPullRequest {
	return StampedPullRequest{
		PullRequest:  *pr,
		Time:         time,
		TimeOpen:     pr.TimeOpen(time),
		DaysOpen:     pr.DaysOpen(time),
		ExpiryStatus: pr.ExpiryStatus(time),
	}
}

func StampNow(prs []PullRequest) []StampedPullRequest {
	stamped := make([]StampedPullRequest, len(prs))
	now := time.Now()
	for i, pr := range prs {
		stamped[i] = pr.Stamp(now)
	}
	return stamped
}
