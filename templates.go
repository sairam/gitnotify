package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"strings"
)

var (
	templateMap = template.FuncMap{
		// "Upper": func(s string) string {
		// 	return strings.ToUpper(s)
		// },
		"partial": partial,
	}
)

const PartialTemplatePath = "tmpl/partials/"
const TemplatePath = "tmpl/"

type SimpleTemplate struct {
	prefix      string
	partialsDir string
	t           *template.Template
}

var templates *SimpleTemplate

func init() {
	ReloadTemplates()
}

func ReloadTemplates() {
	// Templates with functions available to them
	templates = &SimpleTemplate{
		"tmpl/",
		"tmpl/partials/",
		template.New("").Funcs(templateMap),
	}
	load()
	loadPartials()
}

func displayText(res io.Writer, text string) {
	ReloadTemplates()

	tmpl := `
  {{ partial "header" . }}
  <h3>` + text + `</h3>
  <a href="/logout">Logout</a>
  {{ partial "footer" . }}`

	t := templates.t.New("foo")
	t.Parse(tmpl)
	t.Execute(res, struct{}{})

}

// page path relative to 'tmpl', example "settings"
func displayPage(res io.Writer, page string, data interface{}) {
	// reload only in dev environments
	ReloadTemplates()

	tv := templates.t.Lookup(TemplatePath + page)
	tv.Execute(res, data)
}

func load() {
	fis, err := ioutil.ReadDir(templates.prefix)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		name := templates.prefix + fi.Name()
		tmplName := strings.Replace(fi.Name(), ".html", "", 1)

		b, err := ioutil.ReadFile(name)
		_, err = templates.t.New(TemplatePath + tmplName).Parse(string(b))

		if err != nil {
			fmt.Println(err)
		}

	}
}

// Duplicated from above. DRY
func loadPartials() {
	fis, err := ioutil.ReadDir(templates.partialsDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		name := templates.partialsDir + fi.Name()
		tmplName := strings.Replace(fi.Name(), ".html", "", 1)

		b, err := ioutil.ReadFile(name)
		_, err = templates.t.New(PartialTemplatePath + tmplName).Parse(string(b))

		if err != nil {
			fmt.Println(err)
		}
	}
}

// https://github.com/spf13/hugo/blob/master/tpl/template_funcs.go
// https://github.com/spf13/hugo/blob/master/tpl/template.go
func partial(name string, contextList ...interface{}) template.HTML {
	var context interface{}

	if len(contextList) == 0 {
		context = nil
	} else {
		context = contextList[0]
	}
	b := &bytes.Buffer{}
	executeTemplate(context, b, PartialTemplatePath+name)
	return template.HTML(b.String())
}

func executeTemplate(context interface{}, w io.Writer, tmplName string) {
	err := templates.t.ExecuteTemplate(w, tmplName, context)
	if err != nil {
		fmt.Println(err)
	}
}
