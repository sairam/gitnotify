package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

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

func (c *Setting) anyValidNotifications() bool {
	return config.isEmailSetup() || c.User.isValidWebhook()
}

// UserNotification is the customization/scheduling is provided for user
// NOTE: this can be an array of notifications like email, Webhook options.
// not gonna make the change to the array
type UserNotification struct {
	Email     string `yaml:"email"`
	Name      string `yaml:"name"`
	Disabled  bool   `yaml:"disabled"`
	Frequency `yaml:",inline"`

	WebhookURL  string `yaml:"webhook_url"`
	WebhookType string `yaml:"webhook_type"`
}

func (u *UserNotification) isValidWebhook() bool {
	return u.WebhookType == "slack" && u.WebhookURL != ""
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
