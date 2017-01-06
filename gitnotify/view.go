package gitnotify

// Helper Functions used in views

import (
	"html/template"
	"strings"

	"github.com/sairam/kinli"
)

// GithubProvider ..
const GithubProvider = "github"

// GitlabProvider ..
const GitlabProvider = "gitlab"

// InitView initialises the view
func InitView() {
	kinli.CacheMode = config.CacheMode
	kinli.ViewFuncs = template.FuncMap{
		"WebsiteLink":      WebsiteLink,
		"RepoLink":         RepoLink,
		"TreeLink":         TreeLink,
		"CommitLink":       CommitLink,
		"CompareLink":      CompareLink,
		"shortCommit":      shortCommit,
		"cleanRepoName":    cleanRepoName,
		"WebhooksList":     WebhooksList,
		"capitalizeOrNone": capitalizeOrNone,
	}
	kinli.InitTmpl()
}

// WebsiteLink ..
func WebsiteLink(provider string) string {
	return getGitConfig(provider).WebsiteLink()
}

// RepoLink ..
func RepoLink(provider, repo string) string {
	return getGitConfig(provider).RepoLink(repo)
}

// TreeLink ..
func TreeLink(provider, repo, ref string) string {
	return getGitConfig(provider).TreeLink(repo, ref)
}

// CommitLink ..
func CommitLink(provider, repo, ref string) string {
	return getGitConfig(provider).CommitLink(repo, ref)
}

// CompareLink ..
func CompareLink(provider, repo, oldCommit, newCommit string) string {
	return getGitConfig(provider).CompareLink(repo, oldCommit, newCommit)
}

// WebhooksList is used while displaying list of webhooks specific integrations available
func WebhooksList() []string {
	return append([]string{""}, config.WebhookIntegrations...)
}

func capitalizeOrNone(option string) string {
	if option == "" {
		return "None"
	}
	return strings.ToTitle(option)
}

func cleanRepoName(repo string) string {
	return strings.Replace(repo, "/", "__", 3)
}

func shortCommit(commit string) string {
	if len(commit) > 6 {
		return commit[0:6]
	}
	return commit
}
