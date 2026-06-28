package main

import (
	"log/slog"
	"os"

	"github.com/gorilla/mux"

	"merge/internal/github"
	"merge/internal/handler"
)

func main() {
	s := handler.Server{
		BaseURL: "https://www.merge.zone",
		Port:    8080,
		Router:  mux.NewRouter(),
		Logger:  slog.New(slog.NewTextHandler(os.Stderr, nil)),
		GitHub: github.GitHub{
			Token: os.Getenv("PROVIDER_TOKEN_GITHUB"),
		},
	}

	if err := s.Start(); err != nil {
		s.Logger.Error(err.Error())
	}
}
