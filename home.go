package main

import "net/http"

// Responsible for displaying about page and link to login etc.,
func homeHandler(res http.ResponseWriter, req *http.Request) {

	hc := &httpContext{w: res, r: req}
	var userInfo *userInfoSession
	if hc.isUserLoggedIn() {
		userInfo = hc.userLoggedinInfo()
	} else {
		userInfo = &userInfoSession{}
	}

	page := &Page{
		Title:   "Home Page",
		User:    userInfo,
		Flashes: hc.getFlashes(),
		Context: nil,
	}

	displayPage(res, "home", page)
}
