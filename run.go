package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aryann/difflib"
)

const noneString = "<none>"

// contains oldList, newList of the branches
// Used for tracking tags/branches/repoNames
type gitBranchList struct {
	repo    *Repo
	option  string
	oldList []string
	newList []string
}

func (e *gitBranchList) String() string {
	return Stringify(e)
}

// gitCommitDiff tracks old and new commits
type gitCommitDiff struct {
	OldCommit string
	NewCommit string
}

func (e *gitCommitDiff) String() string {
	return Stringify(e)
}

// gitRepoDiffs has the diff for a repoName that is being tracked
// this is used to send emails / hooks
type gitRepoDiffs struct {
	RepoName   string
	Provider   string
	References map[string]*gitCommitDiff
	Others     []*LocalRef
}

func (e *gitRepoDiffs) String() string {
	return Stringify(e)
}

// LocalRef is used tracking Repo and Branch inside the diff
type LocalRef struct {
	Title      string
	References []string
}

func (e *LocalRef) String() string {
	return Stringify(e)
}

func forceRunHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect user if not logged in
	hc := &httpContext{w, r}
	redirected := hc.redirectUnlessLoggedIn()
	if redirected {
		return
	}
	userInfo := hc.userLoggedinInfo()
	configFile := userInfo.getConfigFile()

	conf := new(Setting)
	conf.load(configFile)
	if !isValidEmail(conf.usersEmail()) {
		hc.addFlash("Email is not set. Go to <a href=\"/user\">/user</a> to set")
	} else {
		// hidden feature: useful for testing. pass ?save=false in the POST url
		isSaveFalse := isSaveSetToFalse(r.URL.Query())
		cronJob{configFile, !isSaveFalse}.Run()
		hc.addFlash("Check email to see latest changes")
	}

	http.Redirect(w, r, homePageForLoggedIn, 302)
}

func isValidEmail(email string) bool {
	if email == "" || strings.Contains(email, "@users.noreply.github.com") {
		return false
	}
	return true
}

func isSaveSetToFalse(q url.Values) bool {
	if len(q["save"]) == 0 {
		return false
	}
	if q["save"][0] == "false" {
		return true
	}
	return false
}

// Move to helper. statndard directory walk
func fetchFiles(provider string) []string {
	dir := fmt.Sprintf("%s/%s", config.DataDir, provider)
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	files := make([]string, len(fis))
	for i, fi := range fis {
		if fi.IsDir() {
			// TODO - merge by os.sep
			files[i] = dir + "/" + fi.Name() + "/" + config.SettingsFile
		}
	}
	return files
}

// load all files and adds the cron entries
func getData(provider string) {
	files := fetchFiles(provider)
	for i, filename := range files {
		if filename == "" {
			continue
		}
		conf := new(Setting)
		log.Printf("Processing file %d - %s\n", i, filename)
		conf.load(filename)
		upsertCronEntry(conf)
	}
}

type userNotFound struct{}

func (userNotFound) Error() string {
	return fmt.Sprintf("No email address found for account. Visit <a href=\"/user\">Settings Page</a>")
}

// process is called from the cron job
func process(conf *Setting) (allLocalDiffs []*gitRepoDiffs, err error) {
	if !isValidEmail(conf.usersEmail()) {
		log.Printf("No email address for %s\n", conf.Auth.UserName)
		return nil, &userNotFound{}
	}

	client := getGitClient(conf.Auth.Provider, conf.Auth.Token)
	branch := &gitBranchList{}

	allLocalDiffs = make([]*gitRepoDiffs, 0, len(conf.Repos))

	// loop through repos and their branches
	for _, repo := range conf.Repos {
		var localDiffs = &gitRepoDiffs{
			RepoName: repo.Repo,
			Provider: conf.Auth.Provider,
		}
		allLocalDiffs = append(allLocalDiffs, localDiffs)

		// branch is reused here without creating new ones
		branch.repo = repo

		if repo.Branches || len(repo.NamedReferences) > 0 {
			newBranches, _ := getNewInfo(client, branch, "branches")
			if len(repo.NamedReferences) > 0 {

				data := make(map[string]*gitCommitDiff)
				b := conf.Info[repo.Repo]
				// TODO set newInformation as part of the config loader
				// TODO bug - if we are no longer tracking a branch, we need to remove it from the new list
				if b == nil {
					conf.Info[branch.repo.Repo] = newInformation()
					b = conf.Info[branch.repo.Repo]
				}
				for branch, commitID := range b.Commits {
					data[branch] = &gitCommitDiff{
						OldCommit: commitID,
					}
				}

				// check if data still keeps the data
				diffWithOldCommits(newBranches, branch, data)

				for i, t := range data {
					// save new data from commitDiff.data
					if t.NewCommit != noneString {
						b.Commits[i] = t.NewCommit
					}
				}
				localDiffs.References = data
			}

			if repo.Branches {
				branchesDiff := diffWithOldBranches(newBranches, branch, "branches", conf.Info)
				l := &LocalRef{
					Title:      "Branches",
					References: branchesDiff,
				}
				localDiffs.Others = append(localDiffs.Others, l)
			}
		}

		if repo.Tags {
			newTags, _ := getNewInfo(client, branch, "tags")
			tagsDiff := diffWithOldBranches(newTags, branch, "tags", conf.Info)
			l := &LocalRef{
				Title:      "Tags",
				References: tagsDiff,
			}
			localDiffs.Others = append(localDiffs.Others, l)
		}
	}
	return allLocalDiffs, nil
}

func processDiffForUser(conf *Setting) {
	diff, err := process(conf)

	if err == nil {
		processForMail(diff, conf)
		processForWebhook(diff, conf)
	} else {
		log.Printf("Failure processing %s/%s, %s\n", conf.Auth.Provider, conf.Auth.UserName, err)
	}
}

// option can be tags or branches
func getNewInfo(client GitRemoteIface, branch *gitBranchList, option string) ([]*GitRefWithCommit, error) {
	branch.option = option
	return getBranchTagInfo(client, branch)
}

func diffWithOldBranches(v []*GitRefWithCommit, branch *gitBranchList, option string, info map[string]*Information) []string {
	newBranches := make([]string, len(v))
	for i, a := range v {
		newBranches[i] = a.Name
	}

	branch.newList = newBranches
	t := info[branch.repo.Repo]
	if option == "tags" && t != nil {
		branch.oldList = t.Tags
	} else if option == "branches" && t != nil {
		branch.oldList = t.Branches
	}

	diff := getNewStrings(branch.oldList, branch.newList)
	if t == nil {
		info[branch.repo.Repo] = newInformation()
		t = info[branch.repo.Repo]
	}

	if option == gitRefTag {
		t.Tags = branch.newList
	} else if option == gitRefBranch {
		t.Branches = branch.newList
	}

	return diff
}

func getNewStrings(old, new []string) []string {
	strs := make([]string, 0, 1)
	for _, s := range difflib.Diff(old, new) {
		if s.Delta == difflib.RightOnly {
			strs = append(strs, s.Payload)
		}
	}
	return strs
}

// in the branches we are tracking,
// newcommit is "" // means that we are no longer tracking the branch/ref
// newcommit is <none> if branch is not found/deleted in remote
func diffWithOldCommits(v []*GitRefWithCommit, branch *gitBranchList, data map[string]*gitCommitDiff) {
	for _, a := range branch.repo.NamedReferences {
		s := string(a)
		c := data[s]
		if c == nil {
			c = &gitCommitDiff{}
			data[s] = c
		}
		c.NewCommit = findBranchCommit(v, s)
	}
}

func findBranchCommit(v []*GitRefWithCommit, branch string) string {
	for _, a := range v {
		if a.Name == branch {
			return a.Commit
		}
	}
	return noneString
}
