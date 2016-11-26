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
		return nil
	}

	err = yaml.Unmarshal(data, c)

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
func settingsShowHandler(w http.ResponseWriter, r *http.Request) {

	conf := new(Setting)
	conf.load(configFile)
	out, err := yaml.Marshal(conf)
	fmt.Printf("%s", out)
	fmt.Printf("%s", err)
	fmt.Println(conf)

	displayPage(w, "settings", conf)
}

// SettingsSaveHandler is responsible for persisting the information into the file
// and displays any errors in case of failure. If success redirects to /
func settingsSaveHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	fmt.Println("username:", r.Form["username"])
	fmt.Println("password:", r.Form["password"])

	displayPage(w, "settings", struct{}{})
}
