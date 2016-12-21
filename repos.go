package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	Provider        string
}
type reference string

// SettingsPage ..
type SettingsPage struct {
	CronRunning  bool
	EmailPresent bool
}

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

	// set provider at repo level
	if c.Auth.Provider != "" {
		for _, repo := range c.Repos {
			repo.Provider = c.Auth.Provider
		}
	}

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
	settingsHandler(w, r, formUpdateString)
}

// src=github&repo=harshjv/donut&tree=master
func parseAutoFillOptions(hc *httpContext, q url.Values) *Repo {
	if len(q["repo"]) == 0 {
		return &Repo{}
	}

	if len(q["src"]) > 0 && q["src"][0] != hc.userLoggedinInfo().Provider {
		// TODO: provide href link for q["src"][0] after validation of provider
		hc.addFlash("Please login through " + q["src"][0] + ". You are currently logged in via " + hc.userLoggedinInfo().Provider)
		return &Repo{}
	}

	// validate q["repo"][0]
	references := []reference{}
	if len(q["tree"]) > 0 && q["tree"][0] != "null" {
		references = append(references, reference(q["tree"][0]))
	}
	return &Repo{
		Repo:            "#" + q["repo"][0],
		NamedReferences: references,
	}

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

	if formAction == formUpdateString {
		r.ParseForm()
		if len(r.Form["_delete"]) > 0 && r.Form["_delete"][0] == "true" {
			formAction = "delete"
		}
	}

	var repo *Repo

	switch formAction {
	case "show":
	case formUpdateString:
		var references []reference
		var provider = userInfo.Provider

		// TODO based on the provider, we need to load the config file
		if len(r.Form["provider"]) > 0 {
			provider = r.Form["provider"][0]
		}

		for _, t := range r.Form["references"] {
			// TODO - add flash in case reference name is not a branch
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

		repoPresent := validateRemoteRepoName(provider, conf.Auth.Token, repoName)
		if !repoPresent {
			hc.addFlash("Could not find Repo on " + provider)
			break
		}

		repo = &Repo{
			repoName,
			references,
			contains(r.Form["branches"], "true"),
			contains(r.Form["tags"], "true"),
			provider,
		}

		// TODO move method under repo/settings struct
		info := upsertRepo(conf, repo)
		if info {
			formAction = "create"
		}

		if err := conf.save(configFile); err != nil {
			hc.addFlash("Error saving configuration " + err.Error() + " for " + repo.Repo)
		} else {
			if formAction == "create" {
				hc.addFlash("Started tracking '" + repo.Repo + "'")
			} else {
				hc.addFlash("Updated config for '" + repo.Repo + "'")
			}
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

	newRepo := parseAutoFillOptions(hc, r.URL.Query())

	conf.Repos = append([]*Repo{newRepo}, conf.Repos...)

	t := &SettingsPage{isCronPresentFor(configFile), false}
	if isValidEmail(conf.usersEmail()) {
		t.EmailPresent = true
	}

	page := newPage(hc, "Edit/Add Repos to Track", "Edit/Add Repos to Track", conf, t)
	displayPage(w, "repos", page)
}

func validateRepoName(repo string) string {
	data := repoValidator.FindAllString(repo, -1)
	if len(data) == 1 {
		return data[0]
	}
	return ""

}