package crud

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type QueryParam struct {
	Key, Znak, Value, Type string
	IsQuotes               bool
}

type QueryParams struct {
	Filter *Filter
	ID     *QueryParam
	m      map[string]*QueryParam
	values []string
}

type Filter struct {
	Fields []string `json:"fields,omitempty"`
	Orders []string `json:"orders,omitempty"`
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
}

type Map map[string]interface{}

type Maps []map[string]interface{}

// NewFilter Инициализация фильтра для таблиц
func NewFilter(data []byte) (f Filter, err error) {
	f = Filter{
		Limit:  -1,
		Offset: -1,
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &f)
	}
	return f, err
}

func (m Map) Excludes(e ...string) {
	if len(e) > 0 {
		for _, exclude := range e {
			delete(m, exclude)
		}
	}
}

func (m Maps) Excludes(e ...string) {
	if len(e) > 0 {
		for _, record := range m {
			for _, exclude := range e {
				delete(record, exclude)
			}
		}
	}
}

// Values Строковый массив для sql запроса
func (q *QueryParams) Values() []string {
	return q.values
}

func (q *QueryParams) Set(key string, param *QueryParam) {
	if q.m == nil {
		q.m = make(map[string]*QueryParam)
	}
	q.m[key] = param
	q.values = append(q.values, param.String())
}

// Get Вернуть карту параметров
func (q *QueryParams) Get() map[string]*QueryParam {
	return q.m
}

// Len Длина карты
func (q *QueryParams) Len() int {
	return len(q.m)
}

func QueryFormat(key, value string) *QueryParam {
	var (
		t    string
		znak = "="
	)
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
			case "!%":
				t += "::text"
				znak = "not like"
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
				var (
					values string
					z      = ","
					array  = strings.Split(matches[1], ",")
				)
				for i, m := range array {
					if i == len(array)-1 {
						z = ""
					}
					values += "'" + m + "'" + z
				}
				znak = fmt.Sprintf("in(%s)", values)
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
	return &QueryParam{
		Key:   key,
		Znak:  znak,
		Value: value,
		Type:  t,
	}
}

func (q *QueryParam) String() string {
	return fmt.Sprintf("\"%s\"%s %s %s", q.Key, q.Type, q.Znak, q.Value)
}

// GetQuery Формирование запроса
func (q *QueryParams) GetQuery(columns []string) string {
	var (
		values      []string
		valueFields []string
		query       string
	)

	if q.Len() == 0 {
		return ""
	}

	if q.ID != nil {
		values = append(values, q.ID.String())
	}

	for key, value := range q.m {
		if key == "*" {
			continue
		}
		values = append(values, value.String())
	}

	if value, ok := q.m["*"]; ok {
		// Параметр адресной строки *=
		if q.Filter != nil && len(q.Filter.Fields) > 0 {
			for _, field := range q.Filter.Fields {
				if _, ok = q.m[field]; !ok {
					value.Key = field
					valueFields = append(valueFields, value.String())
				}
			}
		} else {
			for _, column := range columns {
				if _, ok = q.m[column]; !ok {
					value.Key = column
					valueFields = append(valueFields, value.String())
				}
			}
		}
	}

	if len(valueFields) > 0 {
		query = "(" + strings.Join(valueFields, " or ") + ")"
	}
	if len(values) > 0 {
		v := strings.Join(values, " and ")
		if len(query) > 0 {
			query += " and " + v
		} else {
			query = v
		}
	}
	return query
}
