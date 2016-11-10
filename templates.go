package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
)

var (
	templateMap = template.FuncMap{
		// "Upper": func(s string) string {
		// 	return strings.ToUpper(s)
		// },
		"partial": partial,
	}
)

type SimpleTemplate struct {
	prefix      string
	partialsDir string
	t           *template.Template
}

var templates *SimpleTemplate

func init() {
	// Templates with functions available to them
	templates = &SimpleTemplate{
		"tmpl/",
		"tmpl/partials/",
		template.New("").Funcs(templateMap),
	}
	load()
}

func load() {
	name := templates.partialsDir + "header.html"
	fmt.Println(name)
	b, err := ioutil.ReadFile(name)
	_, err = templates.t.New("tmpl/partials/header").Parse(string(b))
	if err != nil {
		fmt.Println(err)
	}

	// , "tmpl/settings.html"
	// fmt.Printf("%#v\n", x)
	// _, err := templates.t.New(name).Parse(tpl)
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
	// _ = context
	// var buf []byte
	b := &bytes.Buffer{}
	executeTemplate(context, b, templates.partialsDir+name+".html")
	return template.HTML(b.String())
	// return template.HTML("yello")
}

func executeTemplate(context interface{}, w io.Writer, name string) {
	// fmt.Println("fix me recursive issue")
	err := templates.t.ExecuteTemplate(w, "tmpl/partials/header", context)
	if err != nil {
		fmt.Println(err)
	}
	// err := templ.Execute(w, context)
}

// init takes in a directory like "tmpl/" to search templates from.
// you can use "partial" like hugo
// Walk through all the files
//
