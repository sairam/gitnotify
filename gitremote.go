package main

import (
	"errors"
	"fmt"
)

// This file provides helper functions to have business and view logic in run.go

// GitClient is used to abstract out github vs gitlab client
type GitClient interface{}

// GitRemoteIface defines the list of methods
type GitRemoteIface interface {
	// Helper methods
	WebsiteLink() string
	RepoLink(string) string
	TreeLink(string, string) string
	CommitLink(string, string) string
	CompareLink(string, string, string) string

	// Methods containing logic
	Branches(string) ([]*GitRefWithCommit, error)
	Tags(string) ([]*GitRefWithCommit, error)

	SearchRepos(string) ([]*searchRepoItem, error)
	SearchUsers(string) ([]*searchUserItem, error)
	DefaultBranch(string) (string, error)
	BranchesWithoutRefs(string) ([]string, error)

	RemoteOrgType(string) (string, error)
	ReposForUser(string) ([]*searchRepoItem, error)
}

type providerNotPresent struct {
	name string
}

func (e providerNotPresent) Error() string {
	return fmt.Sprintf("Provider [%s] is not supported", e.name)
}

// GitRefWithCommit contains branch or tag name with Commit
type GitRefWithCommit struct {
	Name   string
	Commit string
}

func getGitConfig(provider string) GitRemoteIface {
	return getGitClient(provider, "")
}

func getGitClient(provider, token string) GitRemoteIface {
	if provider == GithubProvider {
		return newGithubClient(token)
	} else if provider == GitlabProvider {
		return newGitlabClient(token)
	}
	return &localGitnull{provider}
}

func getBranchTagInfo(client GitRemoteIface, branch *gitBranchList) ([]*GitRefWithCommit, error) {
	if branch.option == gitRefBranch {
		return client.Branches(branch.repo.Repo)
	} else if branch.option == gitRefTag {
		return client.Tags(branch.repo.Repo)
	}
	return nil, errors.New("Operation " + branch.option + " not supported")
}

func getGitTypeAhead(provider, token, search string) ([]*searchRepoItem, error) {
	fmt.Println("Search Request:", search, " Provider: ", provider)
	client := getGitClient(provider, token)
	return client.SearchRepos(search)
}

func getGitBranchInfoForRepo(provider, token, repoName string) (*typeAheadBranchList, error) {
	client := getGitClient(provider, token)

	branchCh := make(chan string)
	branchListCh := make(chan []string)

	go func() {
		defaultBranch, _ := client.DefaultBranch(repoName)
		branchCh <- defaultBranch
		close(branchCh)
	}()

	go func() {
		branchList, _ := client.BranchesWithoutRefs(repoName)
		branchListCh <- branchList
		close(branchListCh)
	}()

	defaultBranch := <-branchCh
	branchList := <-branchListCh

	var err error
	if len(branchList) == 0 || defaultBranch == "" {
		err = errors.New("error fetching branches")
	}

	return &typeAheadBranchList{
		DefaultBranch: defaultBranch,
		AllBranches:   branchList,
	}, err

}

func validateRemoteRepoName(provider, token, repoName string) bool {
	client := getGitClient(provider, token)
	branch, err := client.DefaultBranch(repoName)
	if err != nil || branch == "" {
		return false
	}
	return true
}

func getRemoteOrgType(provider, token, orgName string) (string, bool) {
	client := getGitClient(provider, token)
	orgType, err := client.RemoteOrgType(orgName)
	if err != nil || orgType == "" {
		return "", false
	}
	return orgType, true
}
