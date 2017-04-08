package gitnotify

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/sairam/kinli"
)

// Repository is of the name ^ab-c/d_ef$
var repoValidator = regexp.MustCompile("^[\\p{L}\\d_-]+/[\\.\\p{L}\\d_-]+$")
var orgValidator = regexp.MustCompile("^[\\p{L}\\d_-]+$")

// move to helper file
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func deleteOrg(conf *Setting, delOrg *Organisation) (bool, *Organisation) {
	var orgs []*Organisation
	isProcessed := false
	for _, org := range conf.Orgs {
		if org.Name == delOrg.Name {
			delOrg = org
			isProcessed = true
		} else {
			orgs = append(orgs, org)
		}
	}
	conf.Orgs = orgs
	return isProcessed, delOrg
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

func upsertOrg(conf *Setting, newOrg *Organisation) bool {
	var orgs []*Organisation
	isProcessed := false
	for _, org := range conf.Orgs {
		if org.Name == newOrg.Name {
			orgs = append(orgs, newOrg)
			isProcessed = true
		} else {
			orgs = append(orgs, org)
		}
	}
	if isProcessed == false {
		orgs = append(orgs, newOrg)
	}
	conf.Orgs = orgs
	return !isProcessed
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
func parseAutoFillOptions(hc *kinli.HttpContext, provider string, q url.Values) *Repo {
	if len(q["repo"]) == 0 {
		return &Repo{}
	}

	if len(q["src"]) > 0 && q["src"][0] != provider {
		// TODO: provide href link for q["src"][0] after validation of provider
		hc.AddFlash("Please login through " + q["src"][0] + ". You are currently logged in via " + provider)
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
	hc := &kinli.HttpContext{W: w, R: r}
	if hc.RedirectUnlessAuthed("Kindly Login to customize settings") {
		return
	}
	userInfo := getUserInfo(hc)
	configFile := userInfo.getConfigFile()

	conf := new(Setting)
	conf.load(configFile)

	if formAction == formUpdateString {
		r.ParseForm()
		if len(r.Form["_delete"]) > 0 && r.Form["_delete"][0] == "true" {
			formAction = "delete"
		}
	}

	if getFirstValue(r.Form, "repo") != "" {
		actOnRepos(hc, formAction, r, conf)
	} else if getFirstValue(r.Form, "org") != "" {
		actOnOrgs(hc, formAction, r, conf)
	}

	newRepo := parseAutoFillOptions(hc, userInfo.Provider, r.URL.Query())

	conf.Repos = append([]*Repo{newRepo}, conf.Repos...)

	t := make(map[string]string)
	t["IsCronRunning"] = fmt.Sprintf("%v", isCronPresentFor(configFile))

	page := kinli.NewPage(hc, "Edit/Add Repos to Track", userInfo, conf, t)
	kinli.DisplayPage(w, "repos", page)
}

func actOnOrgs(hc *kinli.HttpContext, formAction string, r *http.Request, conf *Setting) {
	statCount("settings.org." + formAction)
	configFile := conf.Auth.getConfigFile()

	switch formAction {
	case "show":
	case formUpdateString:
		var provider = conf.Auth.Provider

		orgName := validateOrgName(getFirstValue(r.Form, "org"))
		if orgName == "" {
			hc.AddFlash("Invalid Org Name Provided")
			break
		}

		orgType, present := getRemoteOrgType(provider, conf.Auth.Token, orgName)
		if present == false {
			hc.AddFlash(fmt.Sprintf("Org/User Name Not Found with %s", provider))
			return
		}

		org := &Organisation{
			orgName,
			orgType,
			provider,
		}

		// TODO move method under repo/settings struct
		info := upsertOrg(conf, org)
		if info {
			formAction = "create"
		}

		if err := conf.save(configFile); err != nil {
			hc.AddFlash("Error saving configuration " + err.Error() + " for " + org.Name)
		} else {
			if formAction == "create" {
				hc.AddFlash("Started tracking 'user/org:" + org.Name + "'")
			} else {
				hc.AddFlash("Updated config for 'user/org:" + org.Name + "'")
			}
		}
	case "delete":
		orgName := validateOrgName(getFirstValue(r.Form, "org"))
		if orgName == "" {
			hc.AddFlash("Invalid Org Name Provided")
			break
		}

		org := &Organisation{
			Name: orgName,
		}

		// TODO move method under repo/settings struct
		var success bool
		success, org = deleteOrg(conf, org)
		if success == false {
			hc.AddFlash("Error deleting org/user" + org.Name)
		} else {
			if err := conf.save(configFile); err != nil {
				hc.AddFlash("Error saving configuration " + err.Error() + " for " + org.Name)
			} else {
				hc.AddFlash("Delete config for " + org.Name)
			}
		}

	}
}

func actOnRepos(hc *kinli.HttpContext, formAction string, r *http.Request, conf *Setting) {
	statCount("settings.repo." + formAction)
	configFile := conf.Auth.getConfigFile()

	switch formAction {
	case "show":
	case formUpdateString:
		var references []reference
		var provider = conf.Auth.Provider

		repoName := validateRepoName(getFirstValue(r.Form, "repo"))
		if repoName == "" {
			hc.AddFlash("Invalid Repo Name Provided")
			break
		}

		repoPresent := validateRemoteRepoName(provider, conf.Auth.Token, repoName)
		if !repoPresent {
			hc.AddFlash("Could not find Repo on " + provider)
			break
		}

		for _, t := range r.Form["references"] {
			str := strings.TrimSpace(t)
			if str == "" {
				continue
			}
			references = append(references, reference(str))
		}

		repo := &Repo{
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
			hc.AddFlash("Error saving configuration " + err.Error() + " for " + repo.Repo)
		} else {
			if formAction == "create" {
				hc.AddFlash("Started tracking '" + repo.Repo + "'")
			} else {
				hc.AddFlash("Updated config for '" + repo.Repo + "'")
			}
		}
	case "delete":
		repoName := validateRepoName(getFirstValue(r.Form, "repo"))
		if repoName == "" {
			hc.AddFlash("Invalid Repo Name Provided")
			break
		}

		repo := &Repo{
			Repo: repoName,
		}

		// TODO move method under repo/settings struct
		var success bool
		success, repo = deleteRepo(conf, repo)
		if success == false {
			hc.AddFlash("Error deleting Repository " + repo.Repo)
		} else {
			if err := conf.save(configFile); err != nil {
				hc.AddFlash("Error saving configuration " + err.Error() + " for " + repo.Repo)
			} else {
				hc.AddFlash("Delete config for " + repo.Repo)
			}
		}
		// TODO - not going to send an email on Create/Delete
		// we may want to get the latest branch or remove from the main tracked list
		// sendEmail(formAction, repo)

	}

}

func validateRepoName(repo string) string {
	if repo == "" {
		return ""
	}
	data := repoValidator.FindAllString(repo, -1)
	if len(data) == 1 {
		return data[0]
	}
	return ""
}

func validateOrgName(org string) string {
	if org == "" {
		return ""
	}
	data := orgValidator.FindAllString(org, -1)
	if len(data) == 1 {
		return data[0]
	}
	return ""
}
