package main

import "time"

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

