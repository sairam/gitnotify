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
	branchList []string
	repo       *Repo
}

func getData() {
	// var conf *Setting
	// var configFile = "data/github/sairam/settings.yml"
	var accessTokenByUser = os.Getenv("GITHUB_USER_TOKEN") // this is temporary for validating responses

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessTokenByUser},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := githubApp.NewClient(tc)
	branchesURL := "https://api.github.com/repos/sairam/gitnotify/branches"
	v := new([]*BranchInfo)
	req, _ := http.NewRequest("GET", branchesURL, nil)
	client.Do(req, v)
	newBranches := make([]string, len(*v))
	for i, a := range *v {
		newBranches[i] = a.Name
	}

	var repo *Repo
	repo = &Repo{
		Repo: "sairam/gitnotify",
	}
	new := &branches{
		branchList: newBranches,
		repo:       repo,
	}

	old := &branches{
		repo: repo,
	}
	old.load()

	// check data difference with previously saved one
	diff := getNewStrings(old.branchList, new.branchList)
	fmt.Println(diff)

	new.save()

	// TODO
	// diff := diffData(v)
	// sendEmail(diff)
	// persistChanges(v)
}

func (b *branches) fileName() string {
	repo := b.repo
	fileName := strings.Replace(repo.Repo, "/", "__", 1)
	dir := fmt.Sprintf("%s/repo", "data/github/sairam")
	if _, err := os.Stat(dir); err != nil {
		os.Mkdir(dir, 0700)
	}
	return fmt.Sprintf("%s/%s.yml", dir, fileName)
}

func (b *branches) load() error {
	data, err := ioutil.ReadFile(b.fileName())
	if os.IsNotExist(err) {
		return err
	}

	err = yaml.Unmarshal(data, &b.branchList)
	if err != nil {
		return err
	}

	return nil
}

func (b *branches) save() error {
	out, err := yaml.Marshal(b.branchList)
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
