package tmplfunc

import (
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strings"
)

func Join(sep string, elems any) (string, error) {
	switch elems := elems.(type) {
	case nil:
		return "", nil
	case []string:
		return strings.Join(elems, sep), nil
	case []any:
		ss := make([]string, 0, len(elems))
		for _, elem := range elems {
			ss = append(ss, fmt.Sprint(elem))
		}
		return strings.Join(ss, sep), nil
	}
	return "", fmt.Errorf("join: unsupported type %T, expected a slice", elems)
}

func Limit(n int, data any) (any, error) {
	if data == nil {
		return nil, nil
	}
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, errors.New("limit: expected a slice or array")
	}
	if n < 0 {
		n = 0
	}
	if n > v.Len() {
		n = v.Len()
	}
	return v.Slice(0, n).Interface(), nil
}

func Highlight(word string) func(sentence string) template.HTML {
	return func(sentence string) template.HTML {
		if sentence != "" {
			word = strings.TrimSpace(word)
			re, err := regexp.Compile(
				fmt.Sprintf(`(?i)\b(%s)\b`, regexp.QuoteMeta(word)),
			)
			if err == nil {
				sentence = re.ReplaceAllString(sentence, `<span class="highlight">$1</span>`)
			}
		}
		return template.HTML(sentence)
	}
}

func Builtins() template.FuncMap {
	return template.FuncMap{
		"join":  Join,
		"limit": Limit,
	}
}
