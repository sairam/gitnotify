package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	githubApp "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

/*
Example:
  [{
    "name": "1-2-stable",
    "commit": {
      "sha": "5b3f7563ae1b4a7160fda7fe34240d40c5777dcd",
      "url": "https://api.github.com/repos/rails/rails/commits/5b3f7563ae1b4a7160fda7fe34240d40c5777dcd"
    }
  }]
*/

// TagInfo .
// In `Tag` ignore zipball_url, tarball_url
type TagInfo struct {
	Name   string     `json:"name" yaml:"name"`
	Commit *CommitRef `json:"commit" yaml:"commit"`
}

func (e *TagInfo) String() string {
	return Stringify(e)
}

// CommitRef is
type CommitRef struct {
	Sha string `json:"sha" yaml:"sha"`
	URL string `json:"url" yaml:"url"`
}

func (e *CommitRef) String() string {
	return Stringify(e)
}

// Helpers
func cleanRepoName(repo string) string {
	return strings.Replace(repo, "/", "__", 3)
}

func shortCommit(commit string) string {
	if len(commit) > 6 {
		return commit[0:6]
	}
	return commit
}

func githubWebsiteLink() string {
	return config.GithubURLEndPoint
}

func githubRepoLink(repo string) string {
	return fmt.Sprintf(githubRepoEndPoint, repo)
}

func githubTreeLink(repo, ref string) string {
	return fmt.Sprintf(githubTreeURLEndPoint, repo, ref)
}

func githubCommitLink(repo, ref string) string {
	return fmt.Sprintf(githubCommitURLEndPoint, repo, ref)
}

func githubCompareLink(repo, oldCommit, newCommit string) string {
	return fmt.Sprintf(githubCompareURLEndPoint, repo, oldCommit, newCommit)
}

// Helper method to create github client
func newGithubClient(token string) *githubApp.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return githubApp.NewClient(tc)
}

// caches branch response
func githubBranches(client *githubApp.Client, repoName string) []*TagInfo {
	return githubBranchTagInfo(client, repoName, "branches")
}

// caches branch response
func githubTags(client *githubApp.Client, repoName string) []*TagInfo {
	return githubBranchTagInfo(client, repoName, "tags")
}

type defaultBranch struct {
	DefaultBranch string `json:"default_branch"`
}

func githubDefaultBranch(client *githubApp.Client, repoName string) (string, error) {
	v := &defaultBranch{}
	repoURL := fmt.Sprintf("%srepos/%s", config.GithubAPIEndPoint, repoName)
	req, _ := http.NewRequest("GET", repoURL, nil)
	gr, _ := client.Do(req, v)
	if gr.StatusCode == 404 {
		return "", errors.New("repo not found")
	}
	fmt.Println(v.DefaultBranch, " is the default branch")
	return v.DefaultBranch, nil
}

func githubBranchTagInfo(client *githubApp.Client, repoName, option string) []*TagInfo {
	v := new([]*TagInfo)
	branchesURL := fmt.Sprintf("%srepos/%s/%s", config.GithubAPIEndPoint, repoName, option)
	req, _ := http.NewRequest("GET", branchesURL, nil)
	client.Do(req, v)
	return *v
}

func githubSearchRepos(client *githubApp.Client, search string) []*searchRepoItem {
	searchRepositoryURL := fmt.Sprintf("%ssearch/repositories?page=%d&q=%s", config.GithubAPIEndPoint, 1, search)
	fmt.Println("Search Request:", search)
	req, _ := http.NewRequest("GET", searchRepositoryURL, nil)
	v := new(searchRepo)
	client.Do(req, v)
	return v.Items
}
