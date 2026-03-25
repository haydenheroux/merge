package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

type Server struct {
	Port int
	Logger *slog.Logger
	GitHub GitHub
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Info(fmt.Sprintf("Handling request for %s", r.URL.Path))

		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 2 {
			return
		}

		owner := parts[0]
		repo := parts[1]

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
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	s.Logger.Info(fmt.Sprintf("Server starting on port %d", s.Port))
	return server.ListenAndServe()
}
