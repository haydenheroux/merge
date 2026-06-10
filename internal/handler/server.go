package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"merge/internal/github"
	"merge/internal/model"
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
	json, err := gh.GitHub.GetPullRequestsJson(gh.Owner, gh.Repo)
	if err != nil {
		gh.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func Json(gh *GitHubRoute, w http.ResponseWriter, r *http.Request) {
	prs, err := gh.GitHub.GetPullRequests(gh.Owner, gh.Repo)
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

func Page(gh *GitHubRoute, w http.ResponseWriter, r *http.Request) {
	prs, err := gh.GitHub.GetPullRequests(gh.Owner, gh.Repo)
	if err != nil {
		gh.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stamped := model.StampNow(prs)

	merged := 0

	fresh := 0
	stale := 0
	expired := 0

	for _, pr := range stamped {
		if pr.State == model.Merged {
			merged += 1
			continue
		}

		switch pr.ExpiryStatus {
		case model.Fresh:
			fresh += 1
		case model.Expired:
			expired += 1
		case model.Stale:
			stale += 1
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = pages.RepoPage(model.RepoPageProps{
		BaseURL:      gh.BaseURL,
		Owner:        gh.Owner,
		Repo:         gh.Repo,
		PRs:          stamped,
		MergedCount:  merged,
		FreshCount:   fresh,
		StaleCount:   stale,
		ExpiredCount: expired,
	}).Render(r.Context(), w)
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
