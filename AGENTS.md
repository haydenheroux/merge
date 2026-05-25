# Merge Zone

Go web app that visualizes GitHub PR review workflow (merge.zone).

## Commands

```sh
make install      # install templ code generator
make generate     # run `go tool templ generate` — required after editing *.templ
make build        # go build -o merge .
make run          # build + run on :8080
make tidy         # go mod tidy
```

Run without Make: `go tool templ generate && go build -o merge . && ./merge`

## Setup

- **Env**: `source .env` — needs `PROVIDER_TOKEN_GITHUB` (GitHub PAT)
- **No tests exist** — `go test -v ./...` returns nothing

## Architecture

Single `package main` (no monorepo). Key files:

| File | Role |
|---|---|
| `main.go` | Entrypoint — wires Server, reads env |
| `server.go` | HTTP routes, handlers (Page, Json, RawJson) |
| `github.go` | GitHub API client (interface + impl) |
| `pr.go` | Domain types: `PullRequest`, `StampedPullRequest`, expiry logic |
| `markdown.go` | Renders + sanitizes PR body markdown |
| `*.templ` | Server-side HTML templates (a-h/templ) |
| `public/` | Static assets (CSS, JS) |

Routes: `/`, `/{owner}/{repo}`, `/{owner}/{repo}/json`, `/{owner}/{repo}/raw`

## Templ (templates)

- Edit `*.templ` files, then run `make generate` to produce `*_templ.go`
- Generated `*_templ.go` files are **checked into git** — CI runs `go tool templ generate && git diff --exit-code` to verify they're up to date
- If you edit a `.templ` and forget to regenerate, CI will fail

## CI

GitHub Actions: `go build -v ./...` → `go test -v ./...` → `go tool templ generate` → `git diff --exit-code`
