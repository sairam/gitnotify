package main

// Helper Functions used in views like email,

// GithubProvider ..
const GithubProvider = "github"

func WebsiteLink(provider string) string {
	if provider == GithubProvider {
		return githubWebsiteLink()
	}
	return ""
}

func RepoLink(provider string, repo string) string {
	if provider == GithubProvider {
		return githubRepoLink(repo)
	}
	return ""
}

func TreeLink(provider string, repo, ref string) string {
	if provider == GithubProvider {
		return githubTreeLink(repo, ref)
	}
	return ""
}
func CommitLink(provider string, repo, ref string) string {
	if provider == GithubProvider {
		return githubCommitLink(repo, ref)
	}
	return ""
}

func CompareLink(provider string, repo, oldCommit, newCommit string) string {
	if provider == GithubProvider {
		return githubCompareLink(repo, oldCommit, newCommit)
	}
	return ""
}
