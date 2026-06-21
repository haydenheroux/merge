package model

type RepoPageProps struct {
	BaseURL           string
	Owner             string
	Repo              string
	PRs               []StampedPullRequest
	OverallCounts     ExpiryCounts
	ScopeCounts       []ScopeInfo
	ContributorCounts []ContributorInfo
}
