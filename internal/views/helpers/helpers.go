package helpers

import (
	"merge/internal/model"
	"strings"
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

func StripQueryParam(path, key string) string {
	prefix := key + "="
	if idx := strings.Index(path, "?"+prefix); idx != -1 {
		rest := path[idx+1:]
		end := strings.Index(rest, "&")
		if end == -1 {
			return path[:idx]
		}
		return path[:idx] + rest[end:]
	}
	if idx := strings.Index(path, "&"+prefix); idx != -1 {
		rest := path[idx+1:]
		end := strings.Index(rest, "&")
		if end == -1 {
			return path[:idx]
		}
		return path[:idx] + rest[end:]
	}
	return path
}
