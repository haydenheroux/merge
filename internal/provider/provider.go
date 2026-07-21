package provider

import (
	"merge/internal/model"
	"net/url"
	"strconv"
	"time"
)

type Params struct {
	Owner       string
	Repo        string
	AsOf        time.Time
	Scope       *string
	Contributor *string
	Status      *string
}

type Options struct {
	Count int
	Page  int
}

func (o Options) WithPage(page int) Options {
	o.Page = page
	return o
}

func tryInt(key string, query url.Values, fallback int) int {
	if p := query.Get(key); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			return parsed
		}
	}

	return fallback
}

func ParseMux(vars map[string]string, query url.Values) (Params, Options) {
	ps := Params{
		Owner: vars["owner"],
		Repo:  vars["repo"],
		// TODO(hayden): Move this to options for historical slicing?
		AsOf: time.Now(),
	}
	if scope, ok := vars["scope"]; ok {
		ps.Scope = &scope
	}
	if contributor := query.Get("contributor"); contributor != "" {
		ps.Contributor = &contributor
	}
	if status := query.Get("status"); status != "" {
		ps.Status = &status
	}

	options := Options{
		Count: 20,
		Page:  tryInt("page", query, 1),
	}

	return ps, options
}

type PaginationResult struct {
	HasNext bool
}

type Provider interface {
	GetPullRequests(params Params, options Options) ([]model.StampedPullRequest, PaginationResult, error)
	GetPullRequestDiff(owner, repo string, number int) (string, error)
}
