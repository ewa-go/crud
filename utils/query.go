package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type QueryParam struct {
	Key, Znak, Value, Type string
	IsQuotes               bool
}

func QueryFormat(key, value string) QueryParam {
	var (
		t string
	)
	znak := "="
	key = strings.Trim(key, " ")
	r := regexp.MustCompile(`\[(>|<|>-|<-|!|<>|~|!~|~\*|!~\*|\+|!\+|%|:|[aA-zZ]+)]$`)
	if r.MatchString(key) {
		matches := r.FindStringSubmatch(key)
		if len(matches) == 2 {
			znak = matches[1]
			key = r.ReplaceAllString(key, "")
			switch znak {
			case "!":
				znak = "!="
			case ">-":
				znak = ">="
			case "<-":
				znak = "<="
			case "%":
				t += "::text"
				znak = "like"
			case "~", "!~", "~*", "!~*":
				t += "::text"
			case "+":
				t += "::text"
				znak = "similar to"
			case "!+":
				t += "::text"
				znak = "not similar to"
			case ":":
				znak = "between"
				r = regexp.MustCompile(`^\[('.+'):('.+')]$`)
				if r.MatchString(value) {
					matches = r.FindStringSubmatch(value)
					if len(matches) == 3 {
						value = fmt.Sprintf("'%s' and '%s'", matches[1], matches[2])
					}
				}
			}
		}
	}
	if strings.ToLower(value) == "null" {
		switch znak {
		case "=":
			znak = "is"
		case "!=":
			znak = "is not"
		}
	}
	if znak == "=" {
		r = regexp.MustCompile(`^\[(.+)]$`)
		if r.MatchString(value) {
			matches := r.FindStringSubmatch(value)
			if len(matches) == 2 {
				znak = fmt.Sprintf("in(%s)", matches[1])
				value = ""
			}
		}
	}
	if len(value) > 0 && znak != "between" {
		switch strings.ToLower(value) {
		case "null", "true", "false":
		default:
			value = "'" + value + "'"
		}
	}
	return QueryParam{
		Key:   key,
		Znak:  znak,
		Value: value,
		Type:  t,
	}
}

func (q QueryParam) String() string {
	return fmt.Sprintf("\"%s\"%s %s %s", q.Key, q.Type, q.Znak, q.Value)
}
