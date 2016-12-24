package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Used to load github data
type searchRepoItem struct {
	ID          string `json:"name"`
	Name        string `json:"full_name"`
	Description string `json:"description"`
	HomePage    string `json:"homepage"`
	Type        string `json:"type"`
}

// this file is responsible for handling 2 types of typeaheads
// 1. Repository name
// 2. Branch Name
func repoTypeAheadHandler(w http.ResponseWriter, r *http.Request) {

	hc := &httpContext{w, r}
	// Redirect user if not logged in
	redirected := hc.redirectUnlessLoggedIn()
	if redirected {
		return
	}

	userInfo := hc.userLoggedinInfo()
	provider := userInfo.Provider

	search := getFirstValue(r.URL.Query(), "repo")
	result, err := getGitTypeAhead(provider, userInfo.Token, search)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if config.RunMode != runModeDev && provider != GitlabProvider {
		setCacheHeaders(w)
	}

	json.NewEncoder(w).Encode(result)
}

type typeAheadBranchList struct {
	DefaultBranch string   `json:"default_branch"`
	AllBranches   []string `json:"branches"`
}

func branchTypeAheadHandler(w http.ResponseWriter, r *http.Request) {

	hc := &httpContext{w, r}
	// Redirect user if not logged in
	redirected := hc.redirectUnlessLoggedIn()
	if redirected {
		return
	}

	userInfo := hc.userLoggedinInfo()
	provider := userInfo.Provider
	repoName := getFirstValue(r.URL.Query(), "repo")
	if repoName == "" {
		http.NotFound(w, r)
		return
	}

	tab, err := getGitBranchInfoForRepo(provider, userInfo.Token, repoName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if config.RunMode != runModeDev && provider != GitlabProvider {
		setCacheHeaders(w)
	}

	json.NewEncoder(w).Encode(tab)
}

// move to a common helper file
func getFirstValue(q url.Values, key string) string {
	if len(q[key]) == 0 {
		return ""
	}
	return q[key][0]
}

// TODO add option for time or end of date or month or year
func setCacheHeaders(w http.ResponseWriter) {
	// cache for 1 day
	cacheUntil := time.Now().AddDate(0, 0, 1).Format(http.TimeFormat)
	maxAge := time.Now().AddDate(0, 0, 1).Unix()
	cacheSince := time.Now().Format(http.TimeFormat)
	w.Header().Set("Expires", cacheUntil)
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age:%d, public", maxAge))
	w.Header().Set("Last-Modified", cacheSince)
}
