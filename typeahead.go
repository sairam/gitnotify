package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// TypeAhead is responsible for handling 2 types of typeaheads
// 1. Repository name
// 2. Branch Name

func repoTypeAheadHandler(w http.ResponseWriter, r *http.Request) {

	// Redirect user if not logged in
	hc := &httpContext{w, r}
	redirected := hc.redirectUnlessLoggedIn()
	if redirected {
		return
	}
	userInfo := hc.userLoggedinInfo()

	// https://api.github.com/search/repositories?q=tetris+language:assembly&sort=stars&order=desc
	client := newGithubClient(userInfo.Token)
	search := getRepoName(r.URL.Query())

	searchRepositoryURL := fmt.Sprintf("%ssearch/repositories?page=%d&q=%s", githubAPIEndPoint, 1, search)
	fmt.Println(searchRepositoryURL)
	req, _ := http.NewRequest("GET", searchRepositoryURL, nil)
	// take items >full_name
	v := new(searchRepo)
	client.Do(req, v)
	fmt.Println(*v)

	suggestions := make([]string, 0, len(v.Items))
	for _, i := range v.Items {
		suggestions = append(suggestions, i.Name)
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(suggestions)

	cacheSince := time.Now().Format(http.TimeFormat)
	// cache for 1 day
	cacheUntil := time.Now().AddDate(0, 0, 1).Format(http.TimeFormat)
	maxAge := time.Now().AddDate(0, 0, 1).Unix()
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age:%d, public", maxAge))
	w.Header().Set("Last-Modified", cacheSince)
	w.Header().Set("Expires", cacheUntil)

	io.Copy(w, b)
	// write to w
}

type searchRepo struct {
	Items []searchRepoName `json:"items"`
}

type searchRepoName struct {
	Name string `json:"full_name"`
}

func branchTypeAheadHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect user if not logged in
	hc := &httpContext{w, r}
	redirected := hc.redirectUnlessLoggedIn()
	if redirected {
		return
	}
	// userInfo := hc.userLoggedinInfo()
	// repo := getRepoName(r.URL.Query())
	// if repo == "" {
	http.NotFound(w, r)
	return
	// }
	// client := newGithubClient(userInfo.Token)

	// write to w

}
func getRepoName(q url.Values) string {
	if len(q["repo"]) == 0 {
		return ""
	}
	return q["repo"][0]
}
