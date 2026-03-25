package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	Port   int
	Router *mux.Router
	Logger *slog.Logger
	GitHub GitHub
}

func (s *Server) HandleOwnerRepoJson(w http.ResponseWriter, r *http.Request) {
	s.Logger.Info(fmt.Sprintf("Handling request for %s", r.URL.Path))

	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]

	json, err := s.GitHub.GetPullRequestsJson(owner, repo)
	if err != nil {
		s.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func (s *Server) HandleOwnerRepo(w http.ResponseWriter, r *http.Request) {
	s.Logger.Info(fmt.Sprintf("Handling request for %s", r.URL.Path))

	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]

	data, err := s.GitHub.GetPullRequests(owner, repo)
	if err != nil {
		s.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(data)
	if err != nil {
		s.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func (s *Server) Start() error {
	s.Router.HandleFunc("/{owner}/{repo}", s.HandleOwnerRepo)
	s.Router.HandleFunc("/{owner}/{repo}/json", s.HandleOwnerRepoJson)
	s.Logger.Info(fmt.Sprintf("Server starting on port %d", s.Port))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.Router)
}
