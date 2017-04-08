package gitnotify

import (
	"encoding/gob"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	"github.com/sairam/kinli"
)

const (
	userInfoProvider = "provider"
	userInfoName     = "name"
	userInfoEmail    = "email"
	userInfoUser     = "user"
	userInfoToken    = "token"

	loginFlash = "Login to customize settings"
)

// InitSession registers and everything related to the user's session
func InitSession() {
	statCount("users.init_session")
	gob.Register(&Authentication{})
	kinli.SessionName = "_git_notify1"
	kinli.HomePathNonAuthed = "/home"
	kinli.HomePathAuthed = "/"

	var store = sessions.NewFilesystemStore("./sessions", []byte(os.Getenv("SESSION_FS_STORE")))
	// TODO mark session as httpOnly, secure
	// http://www.gorillatoolkit.org/pkg/sessions#Options
	store.Options = &sessions.Options{
		Path: "/",
		// Domain:   config.serverHostWithoutPort(),  // take from config
		MaxAge:   86400 * 30,                      // 30 days
		HttpOnly: true,                            // to avoid cookie stealing and session is on server side
		Secure:   (config.ServerProto == "https"), // for https
	}

	kinli.SessionStore = store

	// init Gothic
	gothic.Store = store
	gothic.GetProviderName = getProviderName

}

// TODO use gob for encoding. See example here - http://www.gorillatoolkit.org/pkg/sessions

func loginTheUser(hc *kinli.HttpContext, userInfo *Authentication, provider string) {
	hc.SetSessionData("user", userInfo)
	hc.SetSessionData("provider", provider)
	hc.AddFlash("Logged in via " + provider)
}

func isAuthed(hc *kinli.HttpContext) bool {
	provider := hc.GetSessionData("provider")
	data, ok := provider.(string)
	if !ok {
		return false
	}
	if data != "" {
		return true
	}
	return false
}

func getUserInfo(hc *kinli.HttpContext) *Authentication {
	var userInfo *Authentication
	var isUserAuthed = isAuthed(hc)
	var ok = false
	if isUserAuthed {
		data := hc.GetSessionData("user")
		userInfo, ok = data.(*Authentication)
	}
	if !isUserAuthed || !ok {
		userInfo = &Authentication{}
	}

	return userInfo
}
