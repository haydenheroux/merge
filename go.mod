module merge

go 1.25.0

require (
	github.com/a-h/templ v0.3.1020
	github.com/gomarkdown/markdown v0.0.0-20260417124207-7d523f7318df
	github.com/gorilla/mux v1.8.1
	github.com/microcosm-cc/bluemonday v1.0.27
)

require (
	github.com/a-h/parse v0.0.0-20250122154542-74294addb73e // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cli/browser v1.3.0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/natefinch/atomic v1.0.1 // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
)

tool github.com/a-h/templ/cmd/templ

// +heroku install ./cmd/merge/
