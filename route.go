package main

// This file is used for testing

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	githubApp "github.com/google/go-github/github"
)

func getData() {
	var conf *Setting
	var accessTokenByUser = os.Getenv("GITHUB_USER_TOKEN") // this is temporary for validating responses
	var configFile = "data/github/sairam/settings.yml"

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessTokenByUser},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := githubApp.NewClient(tc)

	// list all repositories for the authenticated user
	// repos, _, _ := client.Repositories.List("", nil)
	//
	// for _, repo := range repos {
	// 	log.Println(*repo.Name, *repo.DefaultBranch, *repo.BranchesURL)
	// }

	conf = new(Setting)
	err := conf.load(configFile)
	// out, err := yaml.Marshal(conf)
	// log.Println("output is ")
	// log.Printf("%s\n", out)
	log.Println(err)
	branchesURL := "https://api.github.com/repos/sairam/daata-portal/branches"
	log.Println(conf)

	// client = githubApp.NewClient(tc)
	v := new([]*BranchInfo)
	req, _ := http.NewRequest("GET", branchesURL, nil)
	client.Do(req, v)
	log.Println(*v)
	log.Println("Done")

	// check data difference with previously saved one
	// TODO
	// diff := diffData(v)
	// sendEmail(diff)
	// persistChanges(v)
}
