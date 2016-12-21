package main

import (
	"errors"
	"fmt"

	githubApp "github.com/google/go-github/github"
	gitlabApp "github.com/xanzy/go-gitlab"
)

// This file provides helper functions to have business logic in run.go
// TODO: move common error types into this file as well

// GitClient is used to abstract out github vs gitlab client
type GitClient interface{}

// GitRefWithCommit contains branch or tag name with Commit
type GitRefWithCommit struct {
	Name   string
	Commit string
}

func getGitClient(provider, token string) (client GitClient, err error) {
	if provider == GithubProvider {
		client = newGithubClient(token)
	} else if provider == GitlabProvider {
		client = newGitlabClient(token)
	} else {
		err = errors.New("Provider " + provider + " not supported")
	}
	return
}

func getBranchTagInfo(branch *branches) ([]*GitRefWithCommit, error) {
	provider := branch.auth.Provider
	if provider == GithubProvider {
		if branch.option == gitRefBranch {
			return githubBranches(branch.client.(*githubApp.Client), branch.repo.Repo)
		} else if branch.option == gitRefTag {
			return githubTags(branch.client.(*githubApp.Client), branch.repo.Repo)
		}
		return nil, errors.New("Operation " + branch.option + " not supported")
	} else if provider == GitlabProvider {
		if branch.option == gitRefBranch {
			return gitlabBranches(branch.client.(*gitlabApp.Client), branch.repo.Repo)
		} else if branch.option == gitRefTag {
			return gitlabTags(branch.client.(*gitlabApp.Client), branch.repo.Repo)
		}
		return nil, errors.New("Operation " + branch.option + " not supported")
	}
	return nil, errors.New("Provider " + provider + " not supported")
}

func getGitTypeAhead(provider, token, search string) ([]*searchRepoItem, error) {
	fmt.Println("Search Request:", search, " Provider: ", provider)
	client, _ := getGitClient(provider, token)
	if provider == GithubProvider {
		search = githubCleanRepoName(search)
		return githubSearchRepos(client.(*githubApp.Client), search)
	} else if provider == GitlabProvider {
		return gitlabSearchRepos(client.(*gitlabApp.Client), search)
	}
	return nil, errors.New("Provider " + provider + " not supported")
}

func getGitBranchInfoForRepo(provider, token, repoName string) (*typeAheadBranchList, error) {
	client, _ := getGitClient(provider, token)
	if provider == GithubProvider {
		return githubBranchInfo(client.(*githubApp.Client), repoName)
	} else if provider == GitlabProvider {
		return gitlabBranchInfo(client.(*gitlabApp.Client), repoName)
	}
	return nil, errors.New("Provider " + provider + " not supported")
}
