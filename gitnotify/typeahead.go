package gitnotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sairam/kinli"
)

// Used to load github data
type searchUserItem struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Type  string `json:"type"`
}

// Used to load github data
type searchRepoItem struct {
	ID          string `json:"name"`
	Name        string `json:"full_name"`
	Description string `json:"description"`
	HomePage    string `json:"homepage"`
}

// this file is responsible for handling 2 types of typeaheads
// 1. Repository name
// 2. Branch Name
func repoTypeAheadHandler(w http.ResponseWriter, r *http.Request) {

	cacher, setCache := w.(CacheWriterIface)
	var cacheResponse = config.CacheMode
	provider := getFirstValue(r.URL.Query(), "provider")
	repoName := getFirstValue(r.URL.Query(), "repo")
	if repoName == "" {
		http.NotFound(w, r)
		return
	}

	if provider == GithubProvider {
	} else if provider == GitlabProvider {
		cacheResponse = false
	} else {
		provider = GithubProvider
	}

	if cacheResponse && setCache {
		cacher.SetCachePath("repotypeahead/" + provider + "/" + repoName)
		if cacher.WriteFromCache() {
			return
		}
	}

	hc := &kinli.HttpContext{W: w, R: r}
	// Redirect user if not logged in
	if hc.RedirectUnlessAuthed(loginFlash) {
		return
	}

	userInfo := getUserInfo(hc)
	provider = userInfo.Provider
	if cacheResponse && setCache {
		cacher.SetCachePath("repotypeahead/" + provider + "/" + repoName)
	}

	result, err := getGitTypeAhead(provider, userInfo.Token, repoName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if cacheResponse {
		setCacheHeaders(w)
	}

	json.NewEncoder(w).Encode(result)
}

type typeAheadBranchList struct {
	DefaultBranch string   `json:"default_branch"`
	AllBranches   []string `json:"branches"`
}

func branchTypeAheadHandler(w http.ResponseWriter, r *http.Request) {
	cacher, setCache := w.(CacheWriterIface)
	var cacheResponse = config.CacheMode
	provider := getFirstValue(r.URL.Query(), "provider")
	repoName := getFirstValue(r.URL.Query(), "repo")
	if repoName == "" {
		http.NotFound(w, r)
		return
	}

	if provider == GithubProvider {
	} else if provider == GitlabProvider {
		cacheResponse = false
	} else {
		provider = GithubProvider
	}

	if cacheResponse && setCache {
		cacher.SetCachePath("branchtypeahead/" + provider + "/" + repoName)
		if cacher.WriteFromCache() {
			return
		}
	}

	hc := &kinli.HttpContext{W: w, R: r}
	// Redirect user if not logged in
	if hc.RedirectUnlessAuthed(loginFlash) {
		return
	}

	userInfo := getUserInfo(hc)
	provider = userInfo.Provider
	// we are setting again in case provider details in url are different from what was requested
	// we are okay serving from cache in case they are available with the probably incorrect provider from the request
	if cacheResponse && setCache {
		cacher.SetCachePath("branchtypeahead/" + provider + "/" + repoName)
	}

	tab, err := getGitBranchInfoForRepo(provider, userInfo.Token, repoName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if cacheResponse {
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
