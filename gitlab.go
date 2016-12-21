package main

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

// Helpers

func gitlabWebsiteLink() string {
	return config.GitlabURLEndPoint
}

func gitlabRepoLink(repo string) string {
	return fmt.Sprintf(gitlabRepoEndPoint, repo)
}

func gitlabTreeLink(repo, ref string) string {
	return fmt.Sprintf(gitlabTreeURLEndPoint, repo, ref)
}

func gitlabCommitLink(repo, ref string) string {
	return fmt.Sprintf(gitlabCommitURLEndPoint, repo, ref)
}

func gitlabCompareLink(repo, oldCommit, newCommit string) string {
	return fmt.Sprintf(gitlabCompareURLEndPoint, repo, oldCommit, newCommit)
}

func newGitlabClient(token string) *gitlabApp.Client {
	git := gitlabApp.NewOAuthClient(nil, token)
	git.SetBaseURL(strings.TrimRight(config.GitlabAPIEndPoint, "/"))
	return git
}

// repoID can be integer or user/repo format
func gitlabDefaultBranch(client *gitlabApp.Client, repoID string) (string, error) {
	p, _, err := client.Projects.GetProject(repoID)
	return p.DefaultBranch, err
}

func gitlabTags(client *gitlabApp.Client, repoID string) ([]*GitRefWithCommit, error) {
	listBranches, _, err := client.Tags.ListTags(repoID)
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

func gitlabBranches(client *gitlabApp.Client, repoID string) ([]*GitRefWithCommit, error) {
	listBranches, _, err := client.Branches.ListBranches(repoID)
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

func gitlabBranchesWithoutRefs(client *gitlabApp.Client, repoID string) ([]string, error) {
	listBranches, _, err := client.Branches.ListBranches(repoID)
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
func gitlabSearchRepos(client *gitlabApp.Client, search string) ([]*searchRepoItem, error) {
	opt := &gitlabApp.ListProjectsOptions{Search: gitlabApp.String(search)}
	projects, _, err := client.Projects.ListProjects(opt)
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

// TODO run synchronously
func gitlabBranchInfo(client *gitlabApp.Client, repoName string) (*typeAheadBranchList, error) {
	defaultBranch, err := gitlabDefaultBranch(client, repoName)
	if err != nil {
		return nil, err
	}

	branchList, err := gitlabBranchesWithoutRefs(client, repoName)
	if err != nil {
		return nil, err
	}

	return &typeAheadBranchList{
		DefaultBranch: defaultBranch,
		AllBranches:   branchList,
	}, nil
}
