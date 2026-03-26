package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

type Server struct {
	Port   int
	Router *mux.Router
	Logger *slog.Logger
	GitHub *GitHub
}

type GitHubRoute struct {
	*Server
	Owner string
	Repo string
}

func (s *Server) HandleGitHubRoute(f func (*GitHubRoute, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Info(fmt.Sprintf("Handling request for %s", r.URL.Path))

		vars := mux.Vars(r)
		gh := &GitHubRoute{
			Server: s,
			Owner: vars["owner"],
			Repo: vars["repo"],
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
	prs, err := gh.GitHub.GetStampedPullRequests(gh.Owner, gh.Repo)
	if err != nil {
		gh.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(prs)
	if err != nil {
		gh.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

type pageData struct {
	Owner string
	Repo  string
	PRs   []StampedPullRequest
	OpenCount int
	StaleCount int
	ExpiredCount int
}

func Page(gh *GitHubRoute, w http.ResponseWriter, r *http.Request) {
	prs, err := gh.GitHub.GetStampedPullRequests(gh.Owner, gh.Repo)
	if err != nil {
		gh.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t, err := template.ParseFiles(filepath.Join("templates", "page.html"))
	if err != nil {
		gh.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stale := 0
	expired := 0
	for _, pr := range prs {
		if pr.IsExpired {
			expired += 1
		} else if pr.IsStale {
			stale += 1
		}
	}
	err = t.Execute(w, pageData{Owner: gh.Owner, Repo: gh.Repo, PRs: prs, OpenCount: len(prs), StaleCount: stale, ExpiredCount: expired})
	if err != nil {
		gh.Logger.Error(err.Error())
	}
}

func (s *Server) Start() error {
    s.Router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	s.Router.HandleFunc("/{owner}/{repo}", s.HandleGitHubRoute(Page))
	s.Router.HandleFunc("/{owner}/{repo}/", s.HandleGitHubRoute(Page))
	s.Router.HandleFunc("/{owner}/{repo}/json", s.HandleGitHubRoute(Json))
	s.Router.HandleFunc("/{owner}/{repo}/raw", s.HandleGitHubRoute(RawJson))
	s.Logger.Info(fmt.Sprintf("Server starting on port %d", s.Port))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.Router)
}
