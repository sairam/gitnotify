package gitnotify

//Page has all information about the page
type Page struct {
	Title        string
	PageTitle    string
	User         *Authentication
	Flashes      []string
	Context      interface{}
	Data         interface{}
	ClientConfig map[string]string
}

// newPage is used by all HTML contexts to display the template
// Emails do not use Page
func newPage(hc *httpContext, title string, pageTitle string, ctx interface{}, data interface{}) *Page {
	var userInfo *Authentication
	if hc.isUserLoggedIn() {
		userInfo = hc.userLoggedinInfo()
	} else {
		userInfo = &Authentication{}
	}

	page := &Page{
		Title:     title,
		PageTitle: pageTitle,
		User:      userInfo,
		Flashes:   hc.getFlashes(),
		Context:   ctx,
		Data:      data,
	}

	page.ClientConfig = make(map[string]string)
	page.ClientConfig["GoogleAnalytics"] = config.GoogleAnalytics

	return page
}
