package main

// Most of the code in this file is re-used from
// https://github.com/spf13/hugo/blob/master/tpl/template_funcs.go
// https://github.com/spf13/hugo/blob/master/tpl/template.go

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"reflect"
	"strconv"
)

var (
	templateMap = template.FuncMap{
		// "Upper": func(s string) string {
		// 	return strings.ToUpper(s)
		// },
		"partial": partial,
		"eq":      eq,
		"ge":      ge,
		"gt":      gt,
		"le":      le,
		"lt":      lt,
	}
)

// eq returns the boolean truth of arg1 == arg2.
func eq(x, y interface{}) bool {
	normalize := func(v interface{}) interface{} {
		vv := reflect.ValueOf(v)
		switch vv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return vv.Int()
		case reflect.Float32, reflect.Float64:
			return vv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return vv.Uint()
		default:
			return v
		}
	}
	x = normalize(x)
	y = normalize(y)
	return reflect.DeepEqual(x, y)
}

// ne returns the boolean truth of arg1 != arg2.
func ne(x, y interface{}) bool {
	return !eq(x, y)
}

// ge returns the boolean truth of arg1 >= arg2.
func ge(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left >= right
}

// gt returns the boolean truth of arg1 > arg2.
func gt(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left > right
}

// le returns the boolean truth of arg1 <= arg2.
func le(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left <= right
}

// lt returns the boolean truth of arg1 < arg2.
func lt(a, b interface{}) bool {
	left, right := compareGetFloat(a, b)
	return left < right
}

func compareGetFloat(a interface{}, b interface{}) (float64, float64) {
	var left, right float64
	var leftStr, rightStr *string
	av := reflect.ValueOf(a)

	switch av.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		left = float64(av.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		left = float64(av.Int())
	case reflect.Float32, reflect.Float64:
		left = av.Float()
	case reflect.String:
		var err error
		left, err = strconv.ParseFloat(av.String(), 64)
		if err != nil {
			str := av.String()
			leftStr = &str
		}
	}

	bv := reflect.ValueOf(b)

	switch bv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		right = float64(bv.Len())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		right = float64(bv.Int())
	case reflect.Float32, reflect.Float64:
		right = bv.Float()
	case reflect.String:
		var err error
		right, err = strconv.ParseFloat(bv.String(), 64)
		if err != nil {
			str := bv.String()
			rightStr = &str
		}
	}

	switch {
	case leftStr == nil || rightStr == nil:
	case *leftStr < *rightStr:
		return 0, 1
	case *leftStr > *rightStr:
		return 1, 0
	default:
		return 0, 0
	}

	return left, right
}

func partial(name string, contextList ...interface{}) template.HTML {
	var context interface{}

	if len(contextList) == 0 {
		context = nil
	} else {
		context = contextList[0]
	}
	b := &bytes.Buffer{}
	executeTemplate(context, b, pathPartialTemplate+name)
	return template.HTML(b.String())
}

func executeTemplate(context interface{}, w io.Writer, tmplName string) {
	err := templates.t.ExecuteTemplate(w, tmplName, context)
	if err != nil {
		log.Println(err)
	}
}
