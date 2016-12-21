package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	githubApp "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

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
func githubBranches(client *githubApp.Client, repoName string) ([]*GitRefWithCommit, error) {
	return githubBranchTagInfo(client, repoName, "branches")
}

// caches branch response
func githubTags(client *githubApp.Client, repoName string) ([]*GitRefWithCommit, error) {
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

	if gr.StatusCode >= 400 {
		// TODO: 401 - Re-auth the user
		return "", errors.New("repo not found")
	}
	return v.DefaultBranch, nil
}

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
type tagInfo struct {
	Name   string     `json:"name"`
	Commit *commitRef `json:"commit"`
}

func (e *tagInfo) String() string {
	return Stringify(e)
}

type commitRef struct {
	Sha string `json:"sha"`
	URL string `json:"url"`
}

func (e *commitRef) String() string {
	return Stringify(e)
}

func githubBranchTagInfo(client *githubApp.Client, repoName, option string) ([]*GitRefWithCommit, error) {
	v := new([]*tagInfo)
	branchesURL := fmt.Sprintf("%srepos/%s/%s", config.GithubAPIEndPoint, repoName, option)
	req, _ := http.NewRequest("GET", branchesURL, nil)
	client.Do(req, v)
	refs := make([]*GitRefWithCommit, 0, len(*v))

	for _, r := range *v {
		ref := &GitRefWithCommit{
			Name:   r.Name,
			Commit: r.Commit.Sha,
		}
		refs = append(refs, ref)
	}

	return refs, nil
}

// TODO searchRepo is github specific struct
func githubSearchRepos(client *githubApp.Client, search string) ([]*searchRepoItem, error) {
	searchRepositoryURL := fmt.Sprintf("%ssearch/repositories?page=%d&q=%s", config.GithubAPIEndPoint, 1, search)
	req, _ := http.NewRequest("GET", searchRepositoryURL, nil)
	v := new(searchRepo)
	gr, _ := client.Do(req, v)
	if gr.StatusCode >= 400 {
		return nil, errors.New("issue")
	}
	return v.Items, nil
}

func githubCleanRepoName(search string) string {
	search = strings.Replace(search, " ", "+", -1)
	// Add support for regular searches
	if strings.Contains(search, "/") {
		var modifiedRepoValidator = regexp.MustCompile("[\\p{L}\\d_-]+/[\\.\\p{L}\\d_-]*")
		data := modifiedRepoValidator.FindAllString(search, -1)
		d := strings.Split(data[0], "/")
		rep := fmt.Sprintf("%s+user:%s", d[1], d[0])
		search = strings.Replace(search, data[0], rep, 1)
	}
	return search
}

// TODO run synchronously
func githubBranchInfo(client *githubApp.Client, repoName string) (*typeAheadBranchList, error) {
	defaultBranch, err := githubDefaultBranch(client, repoName)
	if err != nil {
		return nil, err
	}

	result, err := githubBranches(client, repoName)
	if err != nil {
		return nil, err
	}

	tab := &typeAheadBranchList{}
	tab.DefaultBranch = defaultBranch
	tab.AllBranches = make([]string, 0, len(result))
	for _, r := range result {
		tab.AllBranches = append(tab.AllBranches, r.Name)
	}
	return tab, nil
}
