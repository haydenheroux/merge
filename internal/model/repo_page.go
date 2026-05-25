package model

type RepoPageProps struct {
	BaseURL      string
	Owner        string
	Repo         string
	PRs          []StampedPullRequest
	MergedCount  int
	FreshCount   int
	StaleCount   int
	ExpiredCount int
}
