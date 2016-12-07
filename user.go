package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
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

	// Only if not set
	if conf.User.Email == "" {
		conf.User.Email = conf.usersEmail()
	}
	if conf.User.Name == "" {
		conf.User.Name = conf.usersName()
	}

	page := newPage(hc, "User Settings", "User Settings", conf)
	displayPage(w, "user", page)
}

// map[email:[sairam.kunala@gmail.com] name:[Sairam] hour:[08 16] weekday:[* 1 2 3 4] tz:[5.5]]
func userSettingsSaveHandler(w http.ResponseWriter, r *http.Request) {
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

	formAction := "update"
	if formAction == "update" {
		r.ParseForm()

		// validate
		e, err := mail.ParseAddress(r.Form["email"][0])
		if err == nil {
			conf.User.Email = e.Address
		} else {
			hc.addFlash("email address provided is invalid format")
		}

		// TODO A bad name can spoil the email address "to" field when sending email
		// limit to 100 chars
		conf.User.Name = r.Form["name"][0]

		conf.User.TimeZone = cleanTz(r.Form["tz"][0])

		// validate hour
		conf.User.Hour = cleanHour(r.Form["hour"])

		// validate weekday
		conf.User.WeekDay = cleanWeekday(r.Form["weekday"])

		conf.save(configFile)

	}
	http.Redirect(w, r, "/user", 302)

}

func cleanTz(oldTZ string) string {
	dfault := "+0000"
	newTZ := ""
	var hour string
	var min string
	if len(oldTZ) < 4 || len(oldTZ) > 5 {
		return dfault
	}
	// 4 or 5 characters
	if len(oldTZ) == 4 {
		newTZ += "+"
		hour = oldTZ[0:2]
		min = oldTZ[2:4]
	} else {
		if oldTZ[0] != '+' && oldTZ[0] != '-' {
			return dfault
		}
		newTZ += oldTZ[0:1]
		hour = oldTZ[1:3]
		min = oldTZ[3:5]
	}
	// max allowed are +14 and -12
	// validate hour and min
	hourI, err := strconv.Atoi(hour)
	if err != nil || hourI > 14 {
		hourI = 0
	}
	minI, err := strconv.Atoi(min)
	if err != nil || minI >= 60 || minI%15 != 0 {
		minI = 0
	}
	return fmt.Sprintf("%c%02d%02d", newTZ[0], hourI, minI)

}

func cleanHour(options []string) string {
	if contains(options, "*") {
		return makeStrings(24, "%02d", ",")
	}
	cleanList := make([]string, 24)
	for _, t := range options {
		c, err := strconv.Atoi(t)
		if err != nil {
			continue
		}
		if c >= 0 && c < 24 {
			cleanList[c] = fmt.Sprintf("%02d", c)
		}
	}
	return strings.Join(deleteEmpty(cleanList), ",")

}

func cleanWeekday(options []string) string {
	if contains(options, "*") {
		return makeStrings(7, "%0d", ",")
	}
	cleanList := make([]string, 7)
	for _, t := range options {
		c, err := strconv.Atoi(t)
		if err != nil {
			continue
		}
		if c >= 0 && c < 7 {
			cleanList[c] = fmt.Sprintf("%d", c)
		}
	}
	return strings.Join(deleteEmpty(cleanList), ",")

}

// makes formatted strings by count and joins them with delim
func makeStrings(count int, format string, delim string) string {
	cleanList := make([]string, count)
	for i := 0; i < count; i++ {
		cleanList[i] = fmt.Sprintf(format, i)
	}
	return strings.Join(cleanList, delim)
}

// TODO: move to helper file
func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
