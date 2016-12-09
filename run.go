package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/aryann/difflib"

	githubApp "github.com/google/go-github/github"
)

const noneString = "<none>"

type branches struct {
	repo    *Repo
	auth    *Authentication
	client  *githubApp.Client
	option  string
	oldList []string
	newList []string
}

func (e *branches) String() string {
	return Stringify(e)
}

type BranchCommit struct {
	OldCommit string
	NewCommit string
}

func (e *BranchCommit) String() string {
	return Stringify(e)
}

type LocalDiffs struct {
	RepoName   string
	Provider   string
	References map[string]*BranchCommit
	Others     []*LocalRef
}

func (e *LocalDiffs) String() string {
	return Stringify(e)
}

var mailContent = &MailContent{}

type MailContent struct {
	WebsiteURL string
	User       string // provider/username
	Name       string
	Data       []*LocalDiffs
}

// LocalRef is used tracking Repo and Branch from the email
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

	// hidden feature: useful for testing. pass ?save=false in the POST url
	isSaveFalse := isSaveSetToFalse(r.URL.Query())

	cronJob{configFile, !isSaveFalse}.Run()

	hc.addFlash("Check email to see latest information")

	http.Redirect(w, r, homePageForLoggedIn, 302)
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
	files := fetchFiles("github")
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
func process(conf *Setting) ([]*LocalDiffs, error) {
	if conf.usersEmail() == "" {
		log.Printf("No email address for %s\n", conf.Auth.UserName)
		return nil, &userNotFound{}
	}
	client := newGithubClient(conf.Auth.Token)
	branch := &branches{
		client: client,
		auth:   conf.Auth,
	}

	var allLocalDiffs = make([]*LocalDiffs, 0, len(conf.Repos))

	// loop through repos and their branches
	for _, repo := range conf.Repos {
		var localDiffs = &LocalDiffs{
			RepoName: repo.Repo,
			Provider: conf.Auth.Provider,
		}
		allLocalDiffs = append(allLocalDiffs, localDiffs)

		// branch is reused here without creating new ones
		branch.repo = repo

		if repo.Branches || len(repo.NamedReferences) > 0 {
			newBranches := getNewInfo(branch, "branches")
			if len(repo.NamedReferences) > 0 {

				data := make(map[string]*BranchCommit)
				b := conf.Info[repo.Repo]
				// TODO set newInformation as part of the config loader
				// TODO bug - if we are no longer tracking a branch, we need to remove it from the new list
				if b == nil {
					conf.Info[branch.repo.Repo] = newInformation()
					b = conf.Info[branch.repo.Repo]
				}
				for branch, commitID := range b.Commits {
					data[branch] = &BranchCommit{
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
			newTags := getNewInfo(branch, "tags")
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
func processForMail(conf *Setting) error {

	diff, _ := process(conf)

	mailContent = &MailContent{
		WebsiteURL: config.ServerProto + config.ServerHost,
		User:       fmt.Sprintf("%s/%s", conf.Auth.Provider, conf.Auth.UserName),
		Name:       conf.usersName(),
	}
	mailContent.Data = diff

	htmlBuffer := &bytes.Buffer{}

	displayPage(htmlBuffer, "changes_mail", mailContent)
	html, _ := ioutil.ReadAll(htmlBuffer)

	loc, _ := time.LoadLocation(conf.User.TimeZoneName)
	t := time.Now().In(loc)

	subject := "[GitNotify] New Updates from your Repositories - " + t.Format("02 Jan 2006 | 15 Hrs")

	to := &recepient{
		Name:     conf.usersName(),
		Address:  conf.usersEmail(),
		UserName: conf.Auth.UserName,
		Provider: conf.Auth.Provider,
	}

	ctx := &emailCtx{
		Subject:  subject,
		TextBody: "diff", // text
		HTMLBody: string(html),
	}

	sendEmail(to, ctx)
	return nil
}

func getNewInfo(branch *branches, option string) []*TagInfo {
	branch.option = option
	branchesURL := fmt.Sprintf("%srepos/%s/%s", config.GithubAPIEndPoint, branch.repo.Repo, branch.option)
	// fmt.Println(branchesURL)
	v := new([]*TagInfo)
	req, _ := http.NewRequest("GET", branchesURL, nil)
	branch.client.Do(req, v)
	return *v
}

func diffWithOldBranches(v []*TagInfo, branch *branches, option string, info map[string]*Information) []string {
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
	if option == "tags" {
		t.Tags = branch.newList
	} else if option == "branches" {
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
func diffWithOldCommits(v []*TagInfo, branch *branches, data map[string]*BranchCommit) {
	for _, a := range branch.repo.NamedReferences {
		s := string(a)
		c := data[s]
		if c == nil {
			c = &BranchCommit{}
			data[s] = c
		}
		c.NewCommit = findBranchCommit(v, s)
	}
}

func findBranchCommit(v []*TagInfo, branch string) string {
	for _, a := range v {
		if a.Name == branch {
			return a.Commit.Sha
		}
	}
	return noneString
}
