package main

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Repository is of the name ^ab-c/d_ef$
var repoValidator = regexp.MustCompile("^[\\p{L}\\d_-]+/[\\.\\p{L}\\d_-]+$")

// move to helper file
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
