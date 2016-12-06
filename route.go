package main

// This file is used for testing

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

type branchCommits struct {
	data map[string]*branchCommit
}

type branchCommit struct {
	// branch    string
	oldCommit string
	newCommit string
}

func fetchFiles(provider string) []string {

	dir := fmt.Sprintf("%s/%s", dataDir, provider)
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	files := make([]string, len(fis))
	for i, fi := range fis {
		if fi.IsDir() {
			files[i] = dir + "/" + fi.Name() + "/" + settingsFile
		}
	}
	return files
}

func getData(provider string) {
	files := fetchFiles("github")
	for i, filename := range files {
		if filename == "" {
			continue
		}
		conf := new(Setting)
		log.Printf("Processing file %d - %s\n", i, filename)
		conf.load(filename)
		process(conf)
	}
}

func process(conf *Setting) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Auth.Token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := githubApp.NewClient(tc)

	branch := &branches{
		client: client,
		auth:   conf.Auth,
	}

	var diff LocalDiff
	// loop through repos and their branches
	for _, repo := range conf.Repos {
		branch.repo = repo

		if repo.Branches || len(repo.NamedReferences) > 0 {
			v := getNewInfo(branch, "branches")
			// commitRefs - TODO load from file
			if len(repo.NamedReferences) > 0 {
				r := &repoBranchCommit{}
				// load file based on repo args into r.data
				data, _ := ioutil.ReadFile("sample.yml")
				yaml.Unmarshal(data, &r.data)
				commitDiff := r.branches(repo.Repo)
				diffWithOldCommits(v, branch, commitDiff)
				// save to file
				for i, t := range commitDiff.data {
					fmt.Printf("%s,%s,%s\n", i, t.oldCommit, t.newCommit)
				}

				// l := &LocalRef{
				// 	Title:      "Commit",
				// 	Repo:       repo.Repo,
				// 	References: commitDiff,
				// }
				// diff.add(l)
			}

			if repo.Branches {
				branchesDiff := diffWithOldBranches(v, branch)
				l := &LocalRef{
					Title:      "Branches",
					Repo:       repo.Repo,
					References: branchesDiff,
				}
				diff.add(l)
			}
		}

		if repo.Tags {
			v := getNewInfo(branch, "tags")
			tagsDiff := diffWithOldBranches(v, branch)
			l := &LocalRef{
				Title:      "Tags",
				Repo:       repo.Repo,
				References: tagsDiff,
			}
			diff.add(l)
		}
	}

	t := time.Now()
	subject := "[GitNotify] New Stuff from your Repositories - " + t.Format("02 Jan 2006")

	to := &recepient{
		Name:    conf.Auth.Name,
		Address: conf.Auth.Email,
	}

	ctx := &emailCtx{
		Subject:  subject,
		TextBody: diff.toText(),
		HTMLBody: diff.toHTML(),
	}

	sendEmail(to, ctx)
}

func getNewInfo(branch *branches, option string) []*TagInfo {
	branch.option = option
	branchesURL := fmt.Sprintf("%s%s/%s", githubAPIEndPoint, branch.repo.Repo, branch.option)
	fmt.Println(branchesURL)
	v := new([]*TagInfo)
	req, _ := http.NewRequest("GET", branchesURL, nil)
	branch.client.Do(req, v)
	return *v
}

func diffWithOldBranches(v []*TagInfo, branch *branches) []string {
	newBranches := make([]string, len(v))
	for i, a := range v {
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

type repoBranchCommit struct {
	data map[string]map[string]string
}

func (r *repoBranchCommit) repos() []string {
	keys := make([]string, 0, len(r.data))
	for k := range r.data {
		keys = append(keys, k)
	}
	return keys
}

func (r *repoBranchCommit) branches(key string) *branchCommits {
	abc := new(branchCommits)
	abc.data = make(map[string]*branchCommit)
	for a, b := range r.data[key] {
		abc.data[a] = &branchCommit{
			oldCommit: b,
		}
	}
	return abc
}

// in the branches we are tracking,
// newcommit is "" // means that we are no longer tracking the branch/ref
// newcommit is <none> if branch is not found/deleted in remote
func diffWithOldCommits(v []*TagInfo, branch *branches, commitRef *branchCommits) {
	for _, a := range branch.repo.NamedReferences {
		s := string(a)
		c := commitRef.data[s]
		if c == nil {
			c = &branchCommit{}
			commitRef.data[s] = c
		}
		c.newCommit = findBranchCommit(v, s)
	}
}

func findBranchCommit(v []*TagInfo, branch string) string {
	for _, a := range v {
		if a.Name == branch {
			return a.Commit.Sha
		}
	}
	return "<none>"
}

func init() {
	getData("github")
}

// 	r := &repoBranchCommit{}
// 	data, _ := ioutil.ReadFile("sample.yml")
// 	yaml.Unmarshal(data, &r.data)
// 	// r
// 	conf := new(Setting)
// 	conf.load("data/github/sairam/settings.yml")
//
// 	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Auth.Token})
// 	tc := oauth2.NewClient(oauth2.NoContext, ts)
// 	client := githubApp.NewClient(tc)
//
// 	branch := &branches{
// 		client: client,
// 		auth:   conf.Auth,
// 	}
// 	repo := conf.Repos[0]
// 	branch.repo = repo
//
// 	v := getNewInfo(branch, "branches")
// 	x := r.branches(repo.Repo)
// 	diffWithOldCommits(v, branch, x)
// 	for i, t := range x {
// 		fmt.Printf("%s,%s,%s\n", i, t.oldCommit, t.newCommit)
// 	}
//
// }
