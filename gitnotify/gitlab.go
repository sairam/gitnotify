package gitnotify

import (
	"fmt"
	"log"
	"strings"

	gitlabApp "github.com/xanzy/go-gitlab"
)

/*
Example:
  [{
    "name":"master",
    "commit":{
      "id":"c36c69c0613a359a41fe5da8e70047bffe7f97c2"
    }
  }]
*/

type localGitlab struct {
	client GitClient
}

// Helpers

func (*localGitlab) WebsiteLink() string {
	return config.GitlabURLEndPoint
}

func (*localGitlab) RepoLink(repo string) string {
	return fmt.Sprintf(gitlabRepoEndPoint, repo)
}

func (*localGitlab) TreeLink(repo, ref string) string {
	return fmt.Sprintf(gitlabTreeURLEndPoint, repo, ref)
}

func (*localGitlab) CommitLink(repo, ref string) string {
	return fmt.Sprintf(gitlabCommitURLEndPoint, repo, ref)
}

func (*localGitlab) CompareLink(repo, oldCommit, newCommit string) string {
	return fmt.Sprintf(gitlabCompareURLEndPoint, repo, oldCommit, newCommit)
}

func (g *localGitlab) Client() *gitlabApp.Client {
	return g.client.(*gitlabApp.Client)
}

func newGitlabClient(token string) *localGitlab {
	if token == "" {
		return &localGitlab{}
	}
	git := gitlabApp.NewOAuthClient(nil, token)
	git.SetBaseURL(strings.TrimRight(config.GitlabAPIEndPoint, "/"))
	return &localGitlab{git}
}

// repoID can be integer or user/repo format
func (g *localGitlab) DefaultBranch(repoID string) (string, error) {
	p, _, err := g.Client().Projects.GetProject(repoID)
	return p.DefaultBranch, err
}

func (g *localGitlab) Tags(repoID string) ([]*GitRefWithCommit, error) {
	listBranches, _, err := g.Client().Tags.ListTags(repoID)
	if err != nil {
		return nil, err
	}
	branches := make([]*GitRefWithCommit, 0, len(listBranches))
	for _, b := range listBranches {
		t := &GitRefWithCommit{
			Name:   b.Name,
			Commit: b.Commit.ID,
		}
		branches = append(branches, t)
	}
	return branches, nil
}

func (g *localGitlab) Branches(repoID string) ([]*GitRefWithCommit, error) {
	listBranches, _, err := g.Client().Branches.ListBranches(repoID)
	if err != nil {
		return nil, err
	}
	branches := make([]*GitRefWithCommit, 0, len(listBranches))
	for _, b := range listBranches {
		t := &GitRefWithCommit{
			Name:   b.Name,
			Commit: b.Commit.ID,
		}
		branches = append(branches, t)
	}
	return branches, nil
}

func (g *localGitlab) BranchesWithoutRefs(repoID string) ([]string, error) {
	listBranches, _, err := g.Client().Branches.ListBranches(repoID)
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0, len(listBranches))
	for _, b := range listBranches {
		branches = append(branches, b.Name)
	}
	return branches, nil
}

// Project.Description contains links as well
func (g *localGitlab) SearchRepos(search string) ([]*searchRepoItem, error) {
	opt := &gitlabApp.ListProjectsOptions{Search: gitlabApp.String(search)}
	projects, _, err := g.Client().Projects.ListProjects(opt)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	t := make([]*searchRepoItem, 0, len(projects))
	for _, p := range projects {
		item := &searchRepoItem{}
		t = append(t, item)
		item.ID = string(p.ID)
		item.Name = p.PathWithNamespace
		item.Description = p.Description
	}
	return t, nil
}

func (g *localGitlab) SearchUsers(_ string) ([]*searchUserItem, error) {
	return []*searchUserItem{}, &providerNotPresent{GitlabProvider}
}

// provided not supported
func (g *localGitlab) RemoteOrgType(_ string) (string, error) {
	return "", &providerNotPresent{GitlabProvider}
}

func (g *localGitlab) ReposForUser(_ string) ([]*searchRepoItem, error) {
	return []*searchRepoItem{}, &providerNotPresent{GitlabProvider}
}
