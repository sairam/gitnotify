package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	githubApp "github.com/google/go-github/github"
)

var (
	conf    *Setting
	verbose bool
)

var accessTokenByUser = os.Getenv("GITHUB_USER_TOKEN") // this is temporary for validating responses

const configFile = "data/github/sairam/settings.yml"

func getData() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessTokenByUser},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := githubApp.NewClient(tc)

	// list all repositories for the authenticated user
	// repos, _, _ := client.Repositories.List("", nil)
	//
	// for _, repo := range repos {
	// 	fmt.Println(*repo.Name, *repo.DefaultBranch, *repo.BranchesURL)
	// }

	verbose = true
	conf = new(Setting)
	err := conf.load(configFile)
	// out, err := yaml.Marshal(conf)
	// fmt.Println("output is ")
	// fmt.Printf("%s\n", out)
	fmt.Println(err)
	branchesURL := "https://api.github.com/repos/sairam/daata-portal/branches"
	fmt.Println(conf)

	// client = githubApp.NewClient(tc)
	v := new([]*BranchInfo)
	req, _ := http.NewRequest("GET", branchesURL, nil)
	client.Do(req, v)
	fmt.Println(*v)
	fmt.Println("Done")

	// check data difference with previously saved one
	// TODO
	// diff := diffData(v)
	// sendEmail(diff)
	// persistChanges(v)
}
