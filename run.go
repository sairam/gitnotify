package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
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

	conf := new(Setting)
	conf.load(configFile)
	if conf.usersEmail() == "" {
		hc.addFlash("Email is not set. Go to <a href=\"/user\">/user</a> to set")
	} else {
		// hidden feature: useful for testing. pass ?save=false in the POST url
		isSaveFalse := isSaveSetToFalse(r.URL.Query())
		cronJob{configFile, !isSaveFalse}.Run()
		hc.addFlash("Check email to see latest changes")
	}

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
			newBranches, _ := getNewInfo(branch, "branches")
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
			newTags, _ := getNewInfo(branch, "tags")
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

func processForWebhook(diff []*LocalDiffs, conf *Setting) error {
	if conf.User.WebhookType == "slack" && conf.User.WebhookURL != "" {
		log.Println("sending a slack message")
		return processForSlack(diff, conf.User.WebhookURL)
	}
	return nil
}

type SlackMessage struct {
	Username    string            `json:"username"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments"`
}

type SlackAttachment struct {
	Fallback       string                 `json:"fallback"`
	Title          string                 `json:"title"`
	Color          string                 `json:"color,omitempty"`
	PreText        string                 `json:"pretext"`
	AuthorName     string                 `json:"author_name"`
	AuthorLink     string                 `json:"author_link"`
	Fields         []SlackAttachmentField `json:"fields"`
	MarkdownFormat []string               `json:"mrkdwn_in"`
	Text           string                 `json:"text"`
	ThumbnailUrl   string                 `json:"thumb_url,omitempty"`
}

type SlackAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackTypeLink struct {
	Text string
	Href string
}

// <http://www.amazon.com|Amazon>
func (s *slackTypeLink) String() string {
	return fmt.Sprintf("<%s|%s>", s.Href, s.Text)
}

// TODO: we are repeating the logic too much. clean it up from text, html emails and this.
// Make a properly formatted struct / json
func processForSlack(diff []*LocalDiffs, slackURL string) error {
	// loop and construct the slack message and send it
	for _, repo := range diff {
		attachments := make([]SlackAttachment, 1)
		for branch, reference := range repo.References {
			var commitDiffLink string
			if reference.NewCommit == noneString {
				commitDiffLink = fmt.Sprintf("Branch is not present")
			} else if reference.OldCommit == "" {
				commitDiffLink = fmt.Sprintf("New branch being tracked. Current Commit is %s", &slackTypeLink{shortCommit(reference.NewCommit), TreeLink(repo.Provider, repo.RepoName, reference.NewCommit)})
			} else if reference.OldCommit == reference.NewCommit {
				commitDiffLink = fmt.Sprintf("No recent changes. Last Commit is %s", &slackTypeLink{shortCommit(reference.NewCommit), TreeLink(repo.Provider, repo.RepoName, reference.NewCommit)})
			} else {
				text := fmt.Sprintf("%s..%s", shortCommit(reference.OldCommit), shortCommit(reference.NewCommit))
				href := CompareLink(repo.Provider, repo.RepoName, reference.OldCommit, reference.NewCommit)
				commitDiffLink = fmt.Sprintf("Compare Diff: %s", (&slackTypeLink{text, href}).String())
			}
			branchTitle := fmt.Sprintf("Branch: %s", branch)

			attachment := SlackAttachment{
				Title:          (&slackTypeLink{branchTitle, TreeLink(repo.Provider, repo.RepoName, branch)}).String(),
				MarkdownFormat: []string{"text"},
				Text:           commitDiffLink,
			}
			attachments = append(attachments, attachment)
		}

		for _, r := range repo.Others {
			links := make([]string, 0, 1)
			for _, treeName := range r.References {
				href := TreeLink(repo.Provider, repo.RepoName, treeName)
				links = append(links, (&slackTypeLink{treeName, href}).String())
			}
			if len(r.References) == 0 {
				links = append(links, "None")
			}
			attachment := SlackAttachment{
				Title:          fmt.Sprintf("New %s:", r.Title),
				Text:           strings.Join(links, "\n"),
				MarkdownFormat: []string{"text"},
			}
			attachments = append(attachments, attachment)
		}

		message := &SlackMessage{
			Username:    "gitnotify",
			Text:        fmt.Sprintf("*Changes for %s*:", &slackTypeLink{repo.RepoName, RepoLink(repo.Provider, repo.RepoName)}),
			Attachments: attachments,
		}

		buf := &bytes.Buffer{}
		json.NewEncoder(buf).Encode(message)
		http.Post(slackURL, "application/json", buf)
	}
	return nil
}

func processForMail(diff []*LocalDiffs, conf *Setting) error {
	mailContent = &MailContent{
		WebsiteURL: config.ServerProto + config.ServerHost,
		User:       fmt.Sprintf("%s/%s", conf.Auth.Provider, conf.Auth.UserName),
		Name:       conf.usersName(),
	}
	mailContent.Data = diff

	htmlBuffer := &bytes.Buffer{}
	displayPage(htmlBuffer, "changes_mail", mailContent)
	html, _ := ioutil.ReadAll(htmlBuffer)

	textBuffer := &bytes.Buffer{}
	displayPage(textBuffer, "changes_mail_text", mailContent)
	text, _ := ioutil.ReadAll(textBuffer)
	textContent := strings.Replace(string(text), "\n\n", "\n", -1)
	textContent = strings.Replace(textContent, "\n\n", "\n", -1)

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
		TextBody: textContent,
		HTMLBody: string(html),
	}

	sendEmail(to, ctx)
	return nil
}

func getNewInfo(branch *branches, option string) ([]*TagInfo, error) {
	branch.option = option
	return githubBranchTagInfo(branch.client, branch.repo.Repo, option)
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
