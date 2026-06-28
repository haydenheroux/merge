package helpers

import (
	"merge/internal/model"
)

func GetStatusClass(pr model.StampedPullRequest) string {
	switch pr.State {
	case model.Merged:
		return "fa-code-merge special"
	}
	switch pr.ExpiryStatus {
	case model.Expired:
		return "fa-skull error"
	case model.Stale:
		return "fa-leaf warn"
	case model.Fresh:
		return "fa-seedling ok"
	default:
		return "fa-eye-slash"
	}
}
