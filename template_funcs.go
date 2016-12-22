package main

// Most of the code in this file is re-used from
// https://github.com/spf13/hugo/blob/master/tpl/template_funcs.go
// https://github.com/spf13/hugo/blob/master/tpl/template.go

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

var (
	templateMap = template.FuncMap{
		// "Upper": func(s string) string {
		// 	return strings.ToUpper(s)
		// },
		"partial":  partial,
		"eq":       eq,
		"ge":       ge,
		"gt":       gt,
		"le":       le,
		"lt":       lt,
		"in":       in,
		"str":      str,
		"slice":    slice,
		"split":    split,
		"safeHTML": safeHTML,
		"dict":     dictionary,
		"minus":    minus,
		"div":      div,
		"mod":      mod,

		"WebsiteLink":      WebsiteLink,
		"RepoLink":         RepoLink,
		"TreeLink":         TreeLink,
		"CommitLink":       CommitLink,
		"CompareLink":      CompareLink,
		"shortCommit":      shortCommit,
		"cleanRepoName":    cleanRepoName,
		"WebhooksList":     WebhooksList,
		"capitalizeOrNone": capitalizeOrNone,
	}
)

func minus(x int, y int) int {
	return x - y
}

func div(x int, y int) int {
	return x / y
}

func mod(x int, y int) int {
	if y <= 0 {
		return 0
	}
	return x % y
}

// from integer
func str(x interface{}) string {
	return fmt.Sprintf("%d", x)
}

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

// slice returns a slice of all passed arguments
func slice(args ...interface{}) []interface{} {
	return args
}

// safeHTML returns a given string as html/template HTML content.
func safeHTML(a interface{}) (template.HTML, error) {
	s, err := cast.ToStringE(a)
	return template.HTML(s), err
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

// indirect is taken from 'text/template/exec.go'
func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return v, true
		}
		if v.Kind() == reflect.Interface && v.NumMethod() > 0 {
			break
		}
	}
	return v, false
}

// in returns whether v is in the set l.  l may be an array or slice.
func in(l interface{}, v interface{}) bool {
	lv := reflect.ValueOf(l)
	vv := reflect.ValueOf(v)

	switch lv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < lv.Len(); i++ {
			lvv := lv.Index(i)
			lvv, isNil := indirect(lvv)
			if isNil {
				continue
			}
			switch lvv.Kind() {
			case reflect.String:
				if vv.Type() == lvv.Type() && vv.String() == lvv.String() {
					return true
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch vv.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if vv.Int() == lvv.Int() {
						return true
					}
				}
			case reflect.Float32, reflect.Float64:
				switch vv.Kind() {
				case reflect.Float32, reflect.Float64:
					if vv.Float() == lvv.Float() {
						return true
					}
				}
			}
		}
	case reflect.String:
		if vv.Type() == lv.Type() && strings.Contains(lv.String(), vv.String()) {
			return true
		}
	}
	return false
}

// split slices an input string into all substrings separated by delimiter.
func split(a interface{}, delimiter string) ([]string, error) {
	aStr, err := cast.ToStringE(a)
	if err != nil {
		return []string{}, err
	}
	return strings.Split(aStr, delimiter), nil
}

// dictionary creates a map[string]interface{} from the given parameters by
// walking the parameters and treating them as key-value pairs.  The number
// of parameters must be even.
func dictionary(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
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
		log.Print(err)
	}
}
