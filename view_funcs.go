package main

import "strings"

// Helper Functions used in views like email,

// GithubProvider ..
const GithubProvider = "github"

// GitlabProvider ..
const GitlabProvider = "gitlab"

// WebsiteLink ..
func WebsiteLink(provider string) string {
	if provider == GithubProvider {
		return githubWebsiteLink()
	} else if provider == GitlabProvider {
		return gitlabWebsiteLink()
	}
	return ""
}

// RepoLink ..
func RepoLink(provider string, repo string) string {
	if provider == GithubProvider {
		return githubRepoLink(repo)
	} else if provider == GitlabProvider {
		return gitlabRepoLink(repo)
	}
	return ""
}

// TreeLink ..
func TreeLink(provider string, repo, ref string) string {
	if provider == GithubProvider {
		return githubTreeLink(repo, ref)
	} else if provider == GitlabProvider {
		return gitlabTreeLink(repo, ref)
	}
	return ""
}

// CommitLink ..
func CommitLink(provider string, repo, ref string) string {
	if provider == GithubProvider {
		return githubCommitLink(repo, ref)
	} else if provider == GitlabProvider {
		return gitlabCommitLink(repo, ref)
	}
	return ""
}

// CompareLink ..
func CompareLink(provider string, repo, oldCommit, newCommit string) string {
	if provider == GithubProvider {
		return githubCompareLink(repo, oldCommit, newCommit)
	} else if provider == GitlabProvider {
		return gitlabCompareLink(repo, oldCommit, newCommit)
	}
	return ""
}

// WebhooksList ..
func WebhooksList() []string {
	return append([]string{""}, config.WebhookIntegrations...)
}

func capitalizeOrNone(option string) string {
	if option == "" {
		return "None"
	}
	return strings.ToTitle(option)
}
