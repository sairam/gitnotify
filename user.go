package main

import (
	"fmt"
	"net/http"
)

func userSettingsShowHandler(w http.ResponseWriter, r *http.Request) {

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

	page := newPage(hc, "User Settings", "User Settings", conf)
	displayPage(w, "user", page)
}

// map[email:[sairam.kunala@gmail.com] name:[Sairam] hour:[08 16] weekday:[* 1 2 3 4] tz:[5.5]]
func userSettingsSaveHandler(w http.ResponseWriter, r *http.Request) {
	formAction := "update"
	if formAction == "update" {
		r.ParseForm()
		fmt.Println(r.Form)
	}

}
