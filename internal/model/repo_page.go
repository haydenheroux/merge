package model

type RepoPageProps struct {
	BaseURL      string
	Owner        string
	Repo         string
	PRs          []StampedPullRequest
	OverallCounts ExpiryCounts
	ScopeCounts map[string]ExpiryCounts
}
