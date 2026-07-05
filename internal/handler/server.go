package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"merge/internal/model"
	"merge/internal/provider"
	"merge/internal/views/components"
	"merge/internal/views/pages"

	"github.com/gorilla/mux"
)

type Server struct {
	BaseURL  string
	Port     int
	Router   *mux.Router
	Logger   *slog.Logger
	Provider provider.Provider
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := pages.IndexPage(s.BaseURL, "", "").Render(r.Context(), w)
	if err != nil {
		s.Logger.Error(err.Error())
	}
}

type ProviderRoute struct {
	*Server
	Params  provider.Params
	Options provider.Options
}

func (s *Server) HandleGitHubRoute(f func(*ProviderRoute, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Info(fmt.Sprintf("Handling request for %s", r.URL.Path))

		params, options := provider.ParseMux(mux.Vars(r), r.URL.Query())
		rt := &ProviderRoute{
			Server:  s,
			Params:  params,
			Options: options,
		}

		f(rt, w, r)
	}
}

func Json(rt *ProviderRoute, w http.ResponseWriter, r *http.Request) {
	prs, _, err := rt.Provider.GetPullRequests(rt.Params, rt.Options)
	if err != nil {
		rt.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(prs)
	if err != nil {
		rt.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func fetchAllPages(rt *ProviderRoute, upToPage int) (all, currentPRs []model.StampedPullRequest, hasNext bool, err error) {
	for p := 1; p <= upToPage; p++ {
		// TODO(hayden): Respect provided options, e.g. min/max
		options := rt.Options.WithPage(p)
		prs, next, e := rt.Provider.GetPullRequests(rt.Params, options)
		if e != nil {
			return nil, nil, false, e
		}
		all = append(all, prs...)
		if p == upToPage {
			currentPRs = prs
			hasNext = next.HasNext
		}
	}
	return all, currentPRs, hasNext, nil
}

func Page(rt *ProviderRoute, w http.ResponseWriter, r *http.Request) {
	// Load-more: fetch all pages 1..current to compute combined stats
	if rt.Options.Page > 1 && r.Header.Get("HX-Request") != "" {
		allPRs, curPRs, hasNext, err := fetchAllPages(rt, rt.Options.Page)
		if err != nil {
			rt.Logger.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextPage := rt.Options.Page + 1

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = pages.LoadMorePRs(
			curPRs,
			rt.Params.Owner, rt.Params.Repo, nextPage, hasNext,
			model.GetCounts(allPRs),
			model.ScopeAges(allPRs),
			"recent",
			model.ContributorActivity(allPRs),
			"recent",
			r.URL.Path,
		).Render(r.Context(), w)
		if err != nil {
			rt.Logger.Error(err.Error())
		}
		return
	}

	prs, hasNext, err := rt.Provider.GetPullRequests(rt.Params, rt.Options)
	if err != nil {
		rt.Logger.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	scope := ""
	if rt.Params.Scope != nil {
		scope = *rt.Params.Scope
	}
	props := model.RepoPageProps{
		BaseURL:           rt.BaseURL,
		Owner:             rt.Params.Owner,
		Repo:              rt.Params.Repo,
		Scope:             scope,
		PRs:               prs,
		OverallCounts:     model.GetCounts(prs),
		ScopeCounts:       model.ScopeAges(prs),
		ScopeSort:         "recent",
		ContributorCounts: model.ContributorActivity(prs),
		ContributorSort:   "recent",
		CurrentPage:       1,
		HasMore:           hasNext.HasNext,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.URL.Query().Get("part") == "contributors" {
		method := r.URL.Query().Get("sort")
		contributors := props.ContributorCounts
		if method == "top" {
			sort.SliceStable(contributors, func(i, j int) bool {
				return contributors[i].Count() > contributors[j].Count()
			})
		}
		err = components.Contributors(contributors, method).Render(r.Context(), w)
	} else if r.URL.Query().Get("part") == "scopes" {
		method := r.URL.Query().Get("sort")
		scopes := props.ScopeCounts
		if method == "top" {
			sort.SliceStable(scopes, func(i, j int) bool {
				return scopes[i].Count() > scopes[j].Count()
			})
		}
		err = components.Scopes(scopes, method, r.URL.Path).Render(r.Context(), w)
	} else if r.Header.Get("HX-Request") != "" {
		pushURL := "/" + rt.Params.Owner + "/" + rt.Params.Repo
		if scope != "" {
			pushURL += "/" + scope
		}
		w.Header().Set("HX-Push", pushURL)
		err = pages.RepoContent(props, r.URL.Path).Render(r.Context(), w)
	} else {
		err = pages.RepoPage(props, r.URL.Path).Render(r.Context(), w)
	}
	if err != nil {
		rt.Logger.Error(err.Error())
	}
}

func (s *Server) Start() error {
	s.Router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	s.Router.HandleFunc("/", s.HandleIndex)

	s.Router.HandleFunc("/{owner}/{repo}", s.HandleGitHubRoute(Page))
	s.Router.HandleFunc("/{owner}/{repo}/", s.HandleGitHubRoute(Page))
	s.Router.HandleFunc("/{owner}/{repo}/{scope}", s.HandleGitHubRoute(Page))

	s.Logger.Info(fmt.Sprintf("Server starting on port %d", s.Port))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.Router)
}
