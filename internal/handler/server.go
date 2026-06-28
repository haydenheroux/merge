package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"

	"merge/internal/github"
	"merge/internal/model"
	"merge/internal/views/components"
	"merge/internal/views/pages"

	"github.com/gorilla/mux"
)

type Server struct {
	BaseURL string
	Port    int
	Router  *mux.Router
	Logger  *slog.Logger
	GitHub  github.GitHubClient
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := pages.IndexPage(s.BaseURL, "", "").Render(r.Context(), w)
	if err != nil {
		s.Logger.Error(err.Error())
	}
}

type GitHubRoute struct {
	*Server
	Owner string
	Repo  string
}

func (s *Server) HandleGitHubRoute(f func(*GitHubRoute, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Info(fmt.Sprintf("Handling request for %s", r.URL.Path))

		vars := mux.Vars(r)
		gh := &GitHubRoute{
			Server: s,
			Owner:  vars["owner"],
			Repo:   vars["repo"],
		}

		f(gh, w, r)
	}
}

func RawJson(gh *GitHubRoute, w http.ResponseWriter, r *http.Request) {
	json, _, err := gh.GitHub.GetPullRequestsJson(gh.Owner, gh.Repo, 1)
	if err != nil {
		gh.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func Json(gh *GitHubRoute, w http.ResponseWriter, r *http.Request) {
	prs, _, err := gh.GitHub.GetPullRequests(gh.Owner, gh.Repo, 1)
	if err != nil {
		gh.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stamped := model.StampNow(prs)

	json, err := json.Marshal(stamped)
	if err != nil {
		gh.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func fetchAllPages(gh *GitHubRoute, upToPage int) (all, currentPRs []model.PullRequest, hasNext bool, err error) {
	for p := 1; p <= upToPage; p++ {
		prs, next, e := gh.GitHub.GetPullRequests(gh.Owner, gh.Repo, p)
		if e != nil {
			return nil, nil, false, e
		}
		all = append(all, prs...)
		if p == upToPage {
			currentPRs = prs
			hasNext = next
		}
	}
	return all, currentPRs, hasNext, nil
}

func Page(gh *GitHubRoute, w http.ResponseWriter, r *http.Request) {
	currentPage := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			currentPage = parsed
		}
	}

	// Load-more: fetch all pages 1..current to compute combined stats
	if currentPage > 1 && r.Header.Get("HX-Request") != "" {
		allPRs, curPRs, hasNext, err := fetchAllPages(gh, currentPage)
		if err != nil {
			gh.Logger.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		allStamped := model.StampNow(allPRs)
		newStamped := model.StampNow(curPRs)

		nextPage := currentPage + 1

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = pages.LoadMorePRs(
			newStamped,
			gh.Owner, gh.Repo, nextPage, hasNext,
			model.GetCounts(allStamped),
			model.ScopeAges(allStamped),
			model.ContributorActivity(allStamped),
		).Render(r.Context(), w)
		if err != nil {
			gh.Logger.Error(err.Error())
		}
		return
	}

	prs, hasNext, err := gh.GitHub.GetPullRequests(gh.Owner, gh.Repo, 1)
	if err != nil {
		gh.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stamped := model.StampNow(prs)

	props := model.RepoPageProps{
		BaseURL:           gh.BaseURL,
		Owner:             gh.Owner,
		Repo:              gh.Repo,
		PRs:               stamped,
		OverallCounts:     model.GetCounts(stamped),
		ScopeCounts:       model.ScopeAges(stamped),
		ContributorCounts: model.ContributorActivity(stamped),
		CurrentPage:       1,
		HasMore:           hasNext,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.URL.Query().Get("part") == "contributors" {
		contributors := props.ContributorCounts
		if r.URL.Query().Get("sort") == "top" {
			sort.SliceStable(contributors, func(i, j int) bool {
				return contributors[i].Count() > contributors[j].Count()
			})
		}
		err = components.Contributors(contributors).Render(r.Context(), w)
	} else if r.URL.Query().Get("part") == "scopes" {
		scopes := props.ScopeCounts
		if r.URL.Query().Get("sort") == "top" {
			sort.SliceStable(scopes, func(i, j int) bool {
				return scopes[i].Count() > scopes[j].Count()
			})
		}
		err = components.Scopes(scopes).Render(r.Context(), w)
	} else if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Push", "/"+gh.Owner+"/"+gh.Repo)
		err = pages.RepoContent(props).Render(r.Context(), w)
	} else {
		err = pages.RepoPage(props).Render(r.Context(), w)
	}
	if err != nil {
		gh.Logger.Error(err.Error())
	}
}

func (s *Server) Start() error {
	s.Router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	s.Router.HandleFunc("/", s.HandleIndex)

	s.Router.HandleFunc("/{owner}/{repo}", s.HandleGitHubRoute(Page))
	s.Router.HandleFunc("/{owner}/{repo}/", s.HandleGitHubRoute(Page))
	s.Router.HandleFunc("/{owner}/{repo}/json", s.HandleGitHubRoute(Json))
	s.Router.HandleFunc("/{owner}/{repo}/raw", s.HandleGitHubRoute(RawJson))

	s.Logger.Info(fmt.Sprintf("Server starting on port %d", s.Port))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.Router)
}
