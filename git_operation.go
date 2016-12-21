package main

import (
	"errors"

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
		if branch.option == "branches" {
			return githubBranches(branch.client.(*githubApp.Client), branch.repo.Repo)
		} else if branch.option == "tags" {
			return githubTags(branch.client.(*githubApp.Client), branch.repo.Repo)
		}
		return nil, errors.New("Operation " + branch.option + " not supported")
	} else if provider == GitlabProvider {
		if branch.option == "branches" {
			return gitlabBranches(branch.client.(*gitlabApp.Client), branch.repo.Repo)
		} else if branch.option == "tags" {
			return gitlabTags(branch.client.(*gitlabApp.Client), branch.repo.Repo)
		}
		return nil, errors.New("Operation " + branch.option + " not supported")
	}
	return nil, errors.New("Provider " + provider + " not supported")

}
