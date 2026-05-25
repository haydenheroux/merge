package helpers

import (
	"fmt"

	"merge/internal/model"
)

func GetStatusClass(pr model.StampedPullRequest) string {
	switch pr.State {
	case model.Merged:
		return "fa-code-merge special"
	}
	switch pr.ExpiryStatus {
	case model.Expired:
		return "fa-recycle error"
	case model.Stale:
		return "fa-apple-whole warn"
	case model.Fresh:
		return "fa-apple-whole ok"
	default:
		return "fa-apple-whole warn"
	}
}

func GetAgeLabel(state model.State) string {
	if state == model.Merged {
		return "ago"
	}
	return "old"
}

func FormatNumber(n int) string {
	return fmt.Sprintf("%d", n)
}
