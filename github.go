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

type localGithub struct {
	client GitClient
}

func (*localGithub) WebsiteLink() string {
	return config.GithubURLEndPoint
}

func (*localGithub) RepoLink(repo string) string {
	return fmt.Sprintf(githubRepoEndPoint, repo)
}

func (*localGithub) TreeLink(repo, ref string) string {
	return fmt.Sprintf(githubTreeURLEndPoint, repo, ref)
}

func (*localGithub) CommitLink(repo, ref string) string {
	return fmt.Sprintf(githubCommitURLEndPoint, repo, ref)
}

func (*localGithub) CompareLink(repo, oldCommit, newCommit string) string {
	return fmt.Sprintf(githubCompareURLEndPoint, repo, oldCommit, newCommit)
}

func (g *localGithub) Client() *githubApp.Client {
	return g.client.(*githubApp.Client)
}

// Helper method to create github client
func newGithubClient(token string) *localGithub {
	if token == "" {
		return &localGithub{}
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return &localGithub{githubApp.NewClient(tc)}
}

func (g *localGithub) BranchesWithoutRefs(repoName string) ([]string, error) {
	listBranches, err := g.Branches(repoName)
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0, len(listBranches))
	for _, b := range listBranches {
		branches = append(branches, b.Name)
	}
	return branches, nil
}

// caches branch response
func (g *localGithub) Branches(repoName string) ([]*GitRefWithCommit, error) {
	return g.branchTagInfo(repoName, gitRefBranch)
}

// caches branch response
func (g *localGithub) Tags(repoName string) ([]*GitRefWithCommit, error) {
	return g.branchTagInfo(repoName, gitRefTag)
}

type ghDefaultBranch struct {
	DefaultBranch string `json:"default_branch"`
}

func (g *localGithub) DefaultBranch(repoName string) (string, error) {
	v := &ghDefaultBranch{}
	repoURL := fmt.Sprintf("%srepos/%s", config.GithubAPIEndPoint, repoName)
	req, _ := http.NewRequest("GET", repoURL, nil)
	gr, _ := g.Client().Do(req, v)

	if gr.StatusCode >= 400 {
		// gr will be in case of connection interruption
		// 401 statusCode means the token is no longer valid
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
type ghTagInfo struct {
	Name   string       `json:"name"`
	Commit *ghCommitRef `json:"commit"`
}

func (e *ghTagInfo) String() string {
	return Stringify(e)
}

type ghCommitRef struct {
	Sha string `json:"sha"`
	URL string `json:"url"`
}

func (e *ghCommitRef) String() string {
	return Stringify(e)
}

func (g *localGithub) branchTagInfo(repoName, option string) ([]*GitRefWithCommit, error) {
	v := new([]*ghTagInfo)
	branchesURL := fmt.Sprintf("%srepos/%s/%s", config.GithubAPIEndPoint, repoName, option)
	req, _ := http.NewRequest("GET", branchesURL, nil)
	g.Client().Do(req, v)
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

type ghSearchRepo struct {
	Items []*searchRepoItem `json:"items"`
}

// searchRepoItem is used by the interface
func (g *localGithub) SearchRepos(search string) ([]*searchRepoItem, error) {
	search = g.cleanRepoName(search)
	searchRepositoryURL := fmt.Sprintf("%ssearch/repositories?page=%d&q=%s", config.GithubAPIEndPoint, 1, search)
	req, _ := http.NewRequest("GET", searchRepositoryURL, nil)
	v := new(ghSearchRepo)
	gr, _ := g.Client().Do(req, v)
	if gr.StatusCode >= 400 {
		return nil, errors.New("issue")
	}
	return v.Items, nil
}

func (g *localGithub) cleanRepoName(search string) string {
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
