package main

// This file is used for testing

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/aryann/difflib"
	"golang.org/x/oauth2"

	githubApp "github.com/google/go-github/github"
)

type branches struct {
	repo    *Repo
	auth    *Authentication
	client  *githubApp.Client
	option  string
	oldList []string
	newList []string
}

func getData() {
	// First get the authentication info
	auth := &Authentication{
		Provider: "github",
		UserName: "sairam",
		Token:    "",
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: auth.Token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := githubApp.NewClient(tc)

	// loop through repos and their branches
	repo := &Repo{
		Repo: "sairam/gitnotify",
	}

	branch := &branches{
		repo:   repo,
		client: client,
		auth:   auth,
	}

	branchesDiff := updateNewBranches(branch, "branches")
	tagsDiff := updateNewBranches(branch, "tags")

	to := &recepient{
		Name:    "Sairam",
		Address: "sairam.kunala@gmail.com",
	}
	ctx := &emailCtx{
		Subject: "[GitNotify] Diff for Repositories - 28th Nov 2016",
		Body:    strings.Join(branchesDiff, "\n") + "\n\n-------------------------\n" + strings.Join(tagsDiff, "\n"),
	}

	sendEmail(to, ctx)
}

func updateNewBranches(branch *branches, option string) []string {
	branch.option = option
	branchesURL := "https://api.github.com/repos/sairam/gitnotify/" + option
	v := new([]*BranchInfo)
	req, _ := http.NewRequest("GET", branchesURL, nil)
	branch.client.Do(req, v)
	newBranches := make([]string, len(*v))
	for i, a := range *v {
		newBranches[i] = a.Name
	}

	branch.newList = newBranches
	branch.load()
	diff := branch.diff()
	branch.save()
	return diff
}

// check data difference with previously saved one
func (b *branches) diff() []string {
	return getNewStrings(b.oldList, b.newList)
}

func (b *branches) fileName() string {
	repo := b.repo
	fileName := strings.Replace(repo.Repo, "/", "__", 1)
	dir := fmt.Sprintf("data/%s/%s/repo", b.auth.Provider, b.auth.UserName)
	if _, err := os.Stat(dir); err != nil {
		os.Mkdir(dir, 0700)
	}
	return fmt.Sprintf("%s/%s-%s.yml", dir, fileName, b.option)
}

// load copies data into oldList
func (b *branches) load() error {
	data, err := ioutil.ReadFile(b.fileName())
	if os.IsNotExist(err) {
		return err
	}

	err = yaml.Unmarshal(data, &b.oldList)
	if err != nil {
		return err
	}

	return nil
}

// save copies data from newList into file
func (b *branches) save() error {
	out, err := yaml.Marshal(b.newList)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(b.fileName(), out, 0600)
}

func getNewStrings(old, new []string) []string {
	var strs []string
	for _, s := range difflib.Diff(old, new) {
		if s.Delta == difflib.RightOnly {
			strs = append(strs, s.Payload)
		}
	}
	return strs
}

// func init() {
// 	getData()
// }
