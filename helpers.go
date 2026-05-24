package main

import "fmt"

func getStatusClass(pr StampedPullRequest) string {
	switch pr.State {
	case Merged:
		return "fa-code-merge special"
	}
	switch pr.ExpiryStatus {
	case Expired:
		return "fa-recycle error"
	case Stale:
		return "fa-apple-whole warn"
	case Fresh:
		return "fa-apple-whole ok"
	default:
		return "fa-apple-whole warn"
	}
}

func getAgeLabel(state State) string {
	if state == Merged {
		return "ago"
	}
	return "old"
}

func formatNumber(n int) string {
	return fmt.Sprintf("%d", n)
}