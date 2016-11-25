package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Setting is the data structure that has all the details
//  data/$provider/$username/settings.yml
type Setting struct {
	Repos          []*Repo `yaml:"repos"`
	Authentication `yaml:"auth"`
}

// Repo is a repository that is being tracked
type Repo struct {
	Repo            string       `yaml:"repo"`
	NamedReferences []*reference `yaml:"commits"`
	Branches        bool         `yaml:"new_branches"`
	Tags            bool         `yaml:"new_tags"`
}

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

type reference string

func (x reference) String() string { return fmt.Sprintf("<%s>", string(x)) }

// read setting from file into memory
func (c *Setting) load(settingFile string) error {

	if _, err := os.Stat(settingFile); os.IsNotExist(err) {
		return nil
	}

	data, err := ioutil.ReadFile(settingFile)

	if os.IsNotExist(err) {
		return nil
	}

	conf := &Setting{}

	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return err
	}

	return nil
}

// persists setting into file
func (c *Setting) save(settingFile string) error {
	return nil
}

// SettingsShowHandler is responsible for displaying the form
func settingsShowHandler(res http.ResponseWriter, req *http.Request) {
	displayPage(res, "settings", struct{}{})
}

// SettingsSaveHandler is responsible for persisting the information into the file
// and displays any errors in case of failure. If success redirects to /
func settingsSaveHandler(res http.ResponseWriter, req *http.Request) {

	ReloadTemplates()

	tv := templates.t.Lookup(TemplatePath + "settings")
	tv.Execute(res, struct{}{})
}
