package model

type RepoPageProps struct {
	BaseURL           string
	Owner             string
	Repo              string
	Scope             string
	Contributor       string
	PRs               []StampedPullRequest
	OverallCounts     ExpiryCounts
	ScopeCounts       []ScopeInfo
	ScopeSort         string
	ContributorCounts []ContributorInfo
	ContributorSort   string
	CurrentPage       int
	HasMore           bool
}
