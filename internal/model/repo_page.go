package model

import (
	"net/url"
	"strconv"
)

type FilterSet struct {
	Owner       string
	Repo        string
	Scope       string
	Contributor string
	Status      string
}

func (f FilterSet) URL() string {
	u := url.URL{Path: f.basePath()}
	if q := f.queryValues(); len(q) > 0 {
		u.RawQuery = q.Encode()
	}
	return u.String()
}

func (f FilterSet) PageURL(page int) string {
	if page <= 1 {
		return f.URL()
	}
	u := url.URL{Path: f.basePath()}
	q := f.queryValues()
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()
	return u.String()
}

func (f FilterSet) basePath() string {
	p := "/" + f.Owner + "/" + f.Repo
	if f.Scope != "" {
		p += "/" + f.Scope
	}
	return p
}

func (f FilterSet) queryValues() url.Values {
	q := url.Values{}
	if f.Contributor != "" {
		q.Set("contributor", f.Contributor)
	}
	if f.Status != "" {
		q.Set("status", f.Status)
	}
	return q
}

func (f FilterSet) WithScope(scope string) FilterSet {
	f.Scope = scope
	return f
}

func (f FilterSet) WithContributor(contributor string) FilterSet {
	f.Contributor = contributor
	return f
}

func (f FilterSet) WithStatus(status string) FilterSet {
	f.Status = status
	return f
}

type RepoPageProps struct {
	BaseURL           string
	Owner             string
	Repo              string
	Scope             string
	Contributor       string
	Status            string
	PRs               []StampedPullRequest
	OverallCounts     ExpiryCounts
	ScopeCounts       []ScopeInfo
	ScopeSort         string
	ContributorCounts []ContributorInfo
	ContributorSort   string
	CurrentPage       int
	HasMore           bool
}
