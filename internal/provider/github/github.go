package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

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

	fetchDiffStats(prs, params.Owner, params.Repo, g.Token)

	stamped := model.StampNow(prs)
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

func (g GitHub) GetPullRequestDiff(owner, repo string, number int) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3.diff")
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func fetchDiffStats(prs []model.PullRequest, owner, repo, token string) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	for i := range prs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prs[i].Number)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return
			}
			req.Header.Set("Accept", "application/vnd.github+json")
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return
			}
			var detail struct {
				Additions int `json:"additions"`
				Deletions int `json:"deletions"`
			}
			if err := json.Unmarshal(body, &detail); err != nil {
				return
			}
			prs[i].Additions = detail.Additions
			prs[i].Deletions = detail.Deletions
		}(i)
	}
	wg.Wait()
}
