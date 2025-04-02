package crud

import (
	"encoding/json"
	"regexp"
	"strings"
)

type QueryParam struct {
	Key      string
	Znak     string
	Value    any
	Type     Type
	DataType string
	IsQuotes bool
	IsOR     bool
}

type QueryParams struct {
	Filter *Filter
	ID     *QueryParam

	m      map[string][]*QueryParam
	values []*QueryParam
}

type Filter struct {
	Fields []string               `json:"fields,omitempty"`
	Orders []string               `json:"orders,omitempty"`
	Limit  int                    `json:"limit,omitempty"`
	Offset int                    `json:"offset,omitempty"`
	Vars   map[string]interface{} `json:"vars,omitempty"`
}

type Map map[string]interface{}

type Maps []map[string]interface{}

type Type string

const (
	ValueType Type = "value"
	ArrayType Type = "array"
	RangeType Type = "range"
)

type Range struct {
	From string
	To   string
}

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
func (q *QueryParams) Values() []*QueryParam {
	return q.values
}

func (q *QueryParams) Set(key string, param *QueryParam) {
	if q.m == nil {
		q.m = make(map[string][]*QueryParam)
	}
	if param != nil {
		q.m[key] = append(q.m[key], param)
		q.values = append(q.values, param)
	}
}

// Get Вернуть карту параметров
func (q *QueryParams) Get() map[string][]*QueryParam {
	return q.m
}

// GetParams Вернуть параметры
func (q *QueryParams) GetParams(key string) []*QueryParam {
	return q.m[key]
}

// Len Длина карты
func (q *QueryParams) Len() int {
	return len(q.m)
}

// QueryFormat Получение параметров из адресной строки
func QueryFormat(r *CRUD, key, value string) *QueryParam {
	q, _ := r.QueryFormat(key, value)
	return q
}

// QueryFormat Получение параметров из адресной строки
func (r *CRUD) QueryFormat(key, value string) (q *QueryParam, err error) {

	q = &QueryParam{
		Key:      strings.Trim(key, " "),
		Znak:     "=",
		IsQuotes: true,
		Type:     ValueType,
		DataType: "",
	}
	if len(q.Key) > 3 && q.Key[:3] == "[|]" {
		q.IsOR = true
		q.Key = q.Key[3:]
	}
	rgx := regexp.MustCompile(r.Pattern())
	if rgx.MatchString(q.Key) {
		matches := rgx.FindStringSubmatch(q.Key)
		if len(matches) == 2 {
			q.Znak = matches[1]
			q.Key = rgx.ReplaceAllString(q.Key, "")
		}
	}
	index := strings.Index(value, "::")
	if index > -1 {
		q.DataType = value[index+2:]
		value = value[:index]
	}
	if err = r.Cast(value, q); err != nil {
		return nil, err
	}
	return r.Format(r, q)
}

func (q *QueryParam) IsValue() bool {
	return q.Type == ValueType
}

func (q *QueryParam) IsArray() bool {
	return q.Type == ArrayType
}

func (q *QueryParam) IsRange() bool {
	return q.Type == RangeType
}

// GetVar Найти значение
func (f Filter) GetVar(key string) any {
	return f.Vars[key]
}
