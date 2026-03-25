package main

import (
	"log/slog"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	s := Server{
		Port: 3000,
		Router: mux.NewRouter(),
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
		GitHub: GitHub{
			Token: os.Getenv("GH_TOKEN"),
		},
	}

	if err := s.Start(); err != nil {
		s.Logger.Error(err.Error())
	}
}
