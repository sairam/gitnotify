package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Repository is of the name ^ab-c/d_ef$
var repoValidator = regexp.MustCompile("^[\\p{L}\\d_-]+/[\\.\\p{L}\\d_-]+$")

// Setting is the data structure that has all the details
//  data/$provider/$username/settings.yml
type Setting struct {
	Version `yaml:"version"`
	Repos   []*Repo                 `yaml:"repos"`
	Auth    *Authentication         `yaml:"auth"`
	User    *UserNotification       `yaml:"user_notification"`
	Info    map[string]*Information `yaml:"fetched_info"`
}

func (c *Setting) usersEmail() string {
	if c.User.Email == "" {
		return c.Auth.Email
	}
	return c.User.Email
}

func (c *Setting) usersName() string {
	if c.User.Name == "" {
		return c.Auth.Name
	}
	return c.User.Name
}

// UserNotification is the customization/scheduling is provided for user
// We are only going to send emails.
// TODO - UserNotification should be an array of type of notifications like name, Webhook with Disabled option etc.,
type UserNotification struct {
	Email     string `yaml:"email"`
	Name      string `yaml:"name"`
	Disabled  bool   `yaml:"disabled"`
	Frequency `yaml:",inline"`

	WebhookURL  string `yaml:"webhook_url"`
	WebhookType string `yaml:"webhook_type"`
}

// Frequency is the cron format along with a TimeZone to process
// Minute, Monthday and Month cannot be controlled. Consider them to be '*'
type Frequency struct {
	TimeZone     string `yaml:"tz"`
	TimeZoneName string `yaml:"tzname"`
	// Minute string // 0-59 allowed
	Hour string `yaml:"hour"` // Hour - 0-23 allowed
	// MonthDay string // 1-31 allowed. Ignore if you want to use weekday vs weekend
	// Month - cannot be set
	WeekDay string `yaml:"weekday"` // 0-6 to point SUN-SAT
}

// Information is all the information fetched from remote location, updated and saved
type Information struct {
	Tags     []string       `yaml:"tags"`
	Branches []string       `yaml:"branches"`
	Commits  LocalCommitRef `yaml:"commits"`
}

func newInformation() *Information {
	i := &Information{}
	i.Commits = make(LocalCommitRef)
	return i
}

// type RepoName string

// LocalCommitRef is of the form map[BranchName] = "1234567890abcdef"
type LocalCommitRef map[string]string

// Version of the structure
type Version string

// Repo is a repository that is being tracked
type Repo struct {
	Repo            string      `yaml:"repo"`
	NamedReferences []reference `yaml:"commits"`
	Branches        bool        `yaml:"new_branches"`
	Tags            bool        `yaml:"new_tags"`
}
type reference string

// var t = afero.Fs

func (c *Setting) String() string {
	arr := make([]string, len(c.Repos))
	for i, repo := range c.Repos {
		arr[i] = fmt.Sprint(repo)
	}
	return strings.Join(arr, "\n")
}

func (r *Repo) String() string {
	return fmt.Sprintf("repo: %s, references: %v, branches: %t, tags: %t", r.Repo, r.NamedReferences, r.Branches, r.Tags)
}

func (x reference) String() string { return fmt.Sprintf("%s", string(x)) }

// read setting from file into memory
func (c *Setting) load(settingFile string) error {

	if _, err := os.Stat(settingFile); os.IsNotExist(err) {
		return nil
	}

	data, err := ioutil.ReadFile(settingFile)
	if os.IsNotExist(err) {
		return err
	}

	err = yaml.Unmarshal(data, c)

	if err != nil {
		return err
	}
	if c.Info == nil {
		c.Info = make(map[string]*Information)
	}

	if c.User == nil {
		c.User = new(UserNotification)
	}

	return nil
}

// persists setting into file
func (c *Setting) save(settingFile string) error {
	out, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(settingFile, out, 0600)
}

// TODO: move to helper file
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func deleteRepo(conf *Setting, delRepo *Repo) (bool, *Repo) {
	var repos []*Repo
	isProcessed := false
	for _, repo := range conf.Repos {
		if repo.Repo == delRepo.Repo {
			delRepo = repo
			isProcessed = true
		} else {
			repos = append(repos, repo)
		}
	}
	conf.Repos = repos
	return isProcessed, delRepo
}

// returns true if newly added row
// returns false if update
func upsertRepo(conf *Setting, newRepo *Repo) bool {
	var repos []*Repo
	isProcessed := false
	for _, repo := range conf.Repos {
		if repo.Repo == newRepo.Repo {
			repos = append(repos, newRepo)
			isProcessed = true
		} else {
			repos = append(repos, repo)
		}
	}
	if isProcessed == false {
		repos = append(repos, newRepo)
	}
	conf.Repos = repos
	return !isProcessed
}

// SettingsShowHandler is responsible for displaying the form
func settingsShowHandler(w http.ResponseWriter, r *http.Request) {
	settingsHandler(w, r, "show")
}

// SettingsSaveHandler is responsible for persisting the information into the file
// and displays any errors in case of failure. If success redirects to /
func settingsSaveHandler(w http.ResponseWriter, r *http.Request) {
	settingsHandler(w, r, "update")
}

func settingsHandler(w http.ResponseWriter, r *http.Request, formAction string) {
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

	if formAction == "update" {
		r.ParseForm()
		if len(r.Form["_delete"]) > 0 && r.Form["_delete"][0] == "true" {
			formAction = "delete"
		}
	}

	var repo *Repo

	switch formAction {
	case "show":
	case "update":
		var references []reference

		for _, t := range r.Form["references"] {
			// TODO - add validation on name of references
			str := strings.TrimSpace(t)
			if str == "" {
				continue
			}
			references = append(references, reference(str))
		}

		repoName := validateRepoName(r.Form["repo"][0])
		if repoName == "" {
			hc.addFlash("Invalid Repo Name Provided")
			break
		}
		repoPresent := validateRemoteRepoName(conf, repoName, "github")
		if !repoPresent {
			hc.addFlash("Could not find Repo on Github")
			break
		}

		repo = &Repo{
			repoName,
			references,
			contains(r.Form["branches"], "true"),
			contains(r.Form["tags"], "true"),
		}

		// TODO move method under repo/settings struct
		info := upsertRepo(conf, repo)
		if info {
			formAction = "create"
		}

		// userInfo := hc.userLoggedinInfo()
		// configFile := userInfo.getConfigFile()
		if err := conf.save(configFile); err != nil {
			hc.addFlash("Error saving configuration " + err.Error() + " for " + repo.Repo)
		} else {
			hc.addFlash("Updated config for " + repo.Repo)
		}
	case "delete":
		repoName := validateRepoName(r.Form["repo"][0])
		if repoName == "" {
			hc.addFlash("Invalid Repo Name Provided")
			break
		}

		repo = &Repo{
			Repo: repoName,
		}
		// TODO move method under repo/settings struct
		var success bool
		success, repo = deleteRepo(conf, repo)
		if success == false {
			hc.addFlash("Error deleting Repository " + repo.Repo)
		} else {
			// userInfo := hc.userLoggedinInfo()
			// configFile := userInfo.getConfigFile()
			if err := conf.save(configFile); err != nil {
				hc.addFlash("Error saving configuration " + err.Error() + " for " + repo.Repo)
			} else {
				hc.addFlash("Delete config for " + repo.Repo)
			}
		}
		// TODO - not going to send an email on Create/Delete
		// we may want to get the latest branch or remove from the main tracked list
		// sendEmail(formAction, repo)

	}

	conf.Repos = append([]*Repo{&Repo{}}, conf.Repos...)

	t := &SettingsPage{isCronPresentFor(configFile), false}
	if conf.usersEmail() != "" {
		t.EmailPresent = true
	}

	page := newPage(hc, "Edit/Add Repos to Track", "Edit/Add Repos to Track", conf, t)
	displayPage(w, "repos", page)
}

type SettingsPage struct {
	CronRunning  bool
	EmailPresent bool
}

func validateRemoteRepoName(conf *Setting, repo string, provider string) bool {
	if provider == "github" {
		client := newGithubClient(conf.Auth.Token)
		branch, err := githubDefaultBranch(client, repo)
		if err != nil || branch == "" {
			return false
		}
		return true
	}
	return false
}

func validateRepoName(repo string) string {

	data := repoValidator.FindAllString(repo, -1)
	if len(data) == 1 {
		return data[0]
	}
	return ""

}
