package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"merge/internal/model"
)

type GitHubClient interface {
	GetPullRequestsJson(owner, repo string, page int) ([]byte, bool, error)
	GetPullRequests(owner, repo string, page int) ([]model.PullRequest, bool, error)
}

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

func (g GitHub) GetPullRequestsJson(owner, repo string, page int) ([]byte, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=all&per_page=100&page=%d", owner, repo, page)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}

	g.setupRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}

	return body, hasNextPage(resp), nil
}

func (g GitHub) GetPullRequests(owner, repo string, page int) ([]model.PullRequest, bool, error) {
	bytes, hasNext, err := g.GetPullRequestsJson(owner, repo, page)
	if err != nil {
		return nil, false, err
	}

	var prs []model.PullRequest

	if err := json.Unmarshal(bytes, &prs); err != nil {
		return nil, false, err
	}

	return prs, hasNext, nil
}
