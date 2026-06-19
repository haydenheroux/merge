package model

import (
	"regexp"
	"strings"
	"time"
)

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

type State string

const (
	Draft  State = "Draft"
	Open         = "Open"
	Merged       = "Merged"
)

func (pr *PullRequest) State() State {
	if pr.IsDraft {
		return Draft
	} else if pr.MergedAt != nil {
		return Merged
	}

	return Open
}

type ExpiryStatus string

const (
	Fresh   ExpiryStatus = "Fresh"
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

func split(scopes string) []string {
	result := strings.Split(scopes, ",")
	for i := range result {
		result[i] = strings.TrimSpace(result[i])
	}
	return result
}

func extractConventionalCommitScopes(title string) []string {
	re := regexp.MustCompile(`^\w+\(([^)]+)\)(?:!)?:`)
	matches := re.FindStringSubmatch(title)
	if len(matches) == 0 {
		return []string{}
	}

	return split(matches[1])
}

func extractSquareBracketScopes(title string) []string {
	re := regexp.MustCompile(`^\[([^\]]+)\]`)
	matches := re.FindStringSubmatch(title)
	if len(matches) == 0 {
		return []string{}
	}

	return split(matches[1])
}

func (pr *PullRequest) Scopes() []string {
	result := make([]string, 0)
	for _, s := range extractConventionalCommitScopes(pr.Title) {
		result = append(result, s)
	}
	for _, s := range extractSquareBracketScopes(pr.Title) {
		result = append(result, s)
	}
	return result
}

type StampedPullRequest struct {
	PullRequest
	Time         time.Time
	State        State
	TimeOpen     time.Duration
	DaysOpen     int
	ExpiryStatus ExpiryStatus
	Type         string
	Scopes       []string
}

func (pr *PullRequest) Stamp(time time.Time) StampedPullRequest {
	spr := StampedPullRequest{
		PullRequest:  *pr,
		Time:         time,
		State:        pr.State(),
		TimeOpen:     pr.TimeOpen(time),
		DaysOpen:     pr.DaysOpen(time),
		ExpiryStatus: pr.ExpiryStatus(time),
		Scopes:       pr.Scopes(),
	}

	return spr
}

func StampNow(prs []PullRequest) []StampedPullRequest {
	stamped := make([]StampedPullRequest, len(prs))
	now := time.Now()
	for i, pr := range prs {
		stamped[i] = pr.Stamp(now)
	}
	return stamped
}
