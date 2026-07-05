package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"merge/internal/model"
	"merge/internal/provider"
)

type GitHub struct {
	Token string
}

func (g GitHub) setupRequest(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
}

func hasNextPage(resp *http.Response) bool {
	link := resp.Header.Get("Link")
	return strings.Contains(link, `rel="next"`)
}

func (g GitHub) getPullRequestsJson(params provider.Params, options provider.Options) ([]byte, provider.PaginationResult, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=all&per_page=%d&page=%d", params.Owner, params.Repo, options.Count, options.Page)

	req, err := http.NewRequest("GET", url, nil)
	res := provider.PaginationResult{}
	if err != nil {
		return nil, res, err
	}

	g.setupRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, res, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, res, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, res, err
	}
	res.HasNext = hasNextPage(resp)

	return body, res, nil
}

func (g GitHub) GetPullRequests(params provider.Params, options provider.Options) ([]model.StampedPullRequest, provider.PaginationResult, error) {
	bytes, hasNext, err := g.getPullRequestsJson(params, options)
	if err != nil {
		return nil, hasNext, err
	}

	var prs []model.PullRequest

	if err := json.Unmarshal(bytes, &prs); err != nil {
		return nil, hasNext, err
	}

	// TODO(hayden): Implement `Mappable` interface
	stamped := model.StampNow(prs) // map from GitHub PRs to our PRs
	if params.Scope == nil {
		return stamped, hasNext, nil
	}

	inScope := make([]model.StampedPullRequest, 0, len(stamped))
	for _, pr := range stamped {
		for _, scope := range pr.Scopes {
			if scope == *params.Scope {
				inScope = append(inScope, pr)
			}
		}
	}

	return inScope, hasNext, nil
}
