package main

import (
	"fmt"

	gitlabApp "github.com/xanzy/go-gitlab"
)

// https://docs.gitlab.com/ce/api/branches.html
// https://gitlab.com/api/v3/projects/2020683/repository/branches

// https://docs.gitlab.com/ce/api/tags.html
// https://gitlab.com/api/v3/projects/2020683/repository/tags
// :proto://:hostname/api/v3/projects/:id/repository/tags
/*
Example:
  [{
    "name":"master",
    "commit":{
      "id":"c36c69c0613a359a41fe5da8e70047bffe7f97c2"
    }
  }]
*/
//

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
	git.SetBaseURL("https://gitlab.com/api/v3")
	return git
}

// repoID can be integer or user/repo format
func gitlabDefaultBranch(client *gitlabApp.Client, repoID string) (string, error) {
	p, _, err := client.Projects.GetProject(repoID)
	return p.DefaultBranch, err
}

func gitlabTagsWithCommits(client *gitlabApp.Client, repoID string) []string {
	listBranches, _, _ := client.Tags.ListTags(repoID)
	branches := make([]string, 0, len(listBranches))
	for _, b := range listBranches {
		branches = append(branches, b.Name)
	}
	return branches
}

func gitlabBranchesWithCommits(client *gitlabApp.Client, repoID string) []*TagInfo {
	listBranches, _, _ := client.Branches.ListBranches(repoID)
	branches := make([]*TagInfo, 0, len(listBranches))
	for _, b := range listBranches {
		t := &TagInfo{
			Name:   b.Name,
			Commit: &CommitRef{Sha: b.Commit.ID},
		}
		branches = append(branches, t)
	}
	return branches
}

func gitlabBranches(client *gitlabApp.Client, repoID string) []string {
	listBranches, _, _ := client.Branches.ListBranches(repoID)
	branches := make([]string, 0, len(listBranches))
	for _, b := range listBranches {
		branches = append(branches, b.Name)
	}
	return branches
}

// Project.Description contains links as well
// TODO - modify Visibility to an option linked with oauth scope
func gitlabSearchRepos(client *gitlabApp.Client, search string) ([]*searchRepoItem, error) {
	opt := &gitlabApp.ListProjectsOptions{Search: gitlabApp.String(search), Visibility: gitlabApp.String("public")}
	projects, _, _ := client.Projects.ListProjects(opt)

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
