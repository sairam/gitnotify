package gitnotify

import (
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	"github.com/sairam/kinli"
	"github.com/sairam/timezone"
)

// UserPage ..
type UserPage struct {
	NextRunTimes []string
}

func userSettingsShowHandler(w http.ResponseWriter, r *http.Request) {

	hc := &kinli.HttpContext{W: w, R: r}
	// Redirect user if not logged in
	if hc.RedirectUnlessAuthed(loginFlash) {
		return
	}
	userInfo := getUserInfo(hc)
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

	// Display next cron entries if cron is valid
	pagedata := struct {
		NextRunTimes []string
		Conf         *Setting
	}{
		NextRunTimes: checkCronEntries(conf.Auth.getConfigFile()),
		Conf:         conf,
	}
	page := kinli.NewPage(hc, "User Settings", userInfo, pagedata, nil)

	kinli.DisplayPage(w, "user", page)
}

func userSettingsSaveHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect user if not logged in
	hc := &kinli.HttpContext{W: w, R: r}
	// Redirect user if not logged in
	if hc.RedirectUnlessAuthed(loginFlash) {
		return
	}

	userInfo := getUserInfo(hc)
	configFile := userInfo.getConfigFile()

	conf := new(Setting)
	conf.load(configFile)

	// isReset := len(r.URL.Query()["reset"]) > 0
	isReset := false

	formAction := "update"
	if isReset &&
		r.URL.Query()["reset"][0] == "default" &&
		conf.User.Disabled == true {
	} else if formAction == "update" {

		r.ParseForm()

		// validate
		if len(r.Form["email"]) > 0 {
			e, err := mail.ParseAddress(r.Form["email"][0])
			if err == nil {
				conf.User.Email = e.Address
			} else {
				hc.AddFlash("email address provided is invalid format")
			}
		}

		// validate
		if len(r.Form["name"]) > 0 {
			conf.User.Name = r.Form["name"][0]
			if len(conf.User.Name) > 100 {
				conf.User.Name = conf.User.Name[0:100]
			}
		}

		if len(r.Form["disabled"]) > 0 {
			if r.Form["disabled"][0] == "tRu3" {
				conf.User.Disabled = true
			}
			if r.Form["disabled"][0] == "enable" {
				conf.User.Disabled = false
			}
		}

		conf.User.TimeZone = cleanTz(r.Form["tz"][0])

		if len(r.Form["tzName"]) > 0 {
			err := cleanTzName(r.Form["tzName"][0])
			if err == nil {
				conf.User.TimeZoneName = r.Form["tzName"][0]
			} else {
				hc.AddFlash(err.Error())
			}
		} else {
			offset := convertTzOffsetToInt(conf.User.TimeZone)
			timeZoneNames := tzByOffset[offset]
			if len(timeZoneNames) > 0 {
				conf.User.TimeZoneName = timeZoneNames[0].Location
			} else {
				log.Println("Could not find TimeZone Name for ", conf.User.TimeZone)
				conf.User.TimeZoneName = "UTC"
			}
		}

		conf.User.Hour = cleanHour(r.Form["hour"])

		// validate weekday
		conf.User.WeekDay = cleanWeekday(r.Form["weekday"])

		if len(r.Form["webhookType"]) > 0 {
			conf.User.WebhookType = r.Form["webhookType"][0]
		}

		if len(r.Form["webhookURL"]) > 0 {
			if _, err := url.ParseRequestURI(r.Form["webhookURL"][0]); err == nil {
				conf.User.WebhookURL = r.Form["webhookURL"][0]
			}
		}

		conf.save(configFile)
		upsertCronEntry(conf)

	}
	http.Redirect(w, r, "/user", 302)

}

func convertTzOffsetToInt(offset string) int {
	intOffset := 1
	if offset[0:1] == "-" {
		intOffset = -1
	}
	hour, _ := strconv.Atoi(offset[1:3])
	minute, _ := strconv.Atoi(offset[3:5])
	return (intOffset*hour*60 + minute) * 60
}

type invalidTimezone struct{}

func (invalidTimezone) Error() string {
	return fmt.Sprintf("Invalid TimeZone selected")
}

func cleanTzName(tzName string) error {

	if !timezone.ValidLocation(tzName) {
		return &invalidTimezone{}
	}
	return nil

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

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
