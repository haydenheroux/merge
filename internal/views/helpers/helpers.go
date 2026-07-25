package helpers

import (
	"fmt"
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

func BuildLoadMoreURL(owner, repo, scope, contributor, status string, page int) string {
	base := "/" + owner + "/" + repo
	if scope != "" {
		base += "/" + scope
	}

	parts := []string{}
	if page > 1 {
		parts = append(parts, fmt.Sprintf("page=%d", page))
	}
	if contributor != "" {
		parts = append(parts, "contributor="+contributor)
	}
	if status != "" {
		parts = append(parts, "status="+status)
	}
	if len(parts) > 0 {
		base += "?" + strings.Join(parts, "&")
	}
	return base
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

func PrLabel(count int) string {
	if count == 1 {
		return "pr"
	}
	return "prs"
}
