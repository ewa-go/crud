package crud

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ewa-go/ewa"
)

type IHandlers interface {
	Columns(r *CRUD, fields ...string) []string
	SetRecord(c *ewa.Context, r *CRUD, data *Body, params *QueryParams) (status int, result any, err error)
	GetRecord(c *ewa.Context, r *CRUD, params *QueryParams) (status int, data Map, err error)
	GetRecords(c *ewa.Context, r *CRUD, params *QueryParams) (status int, data Maps, total int64, err error)
	UpdateRecord(c *ewa.Context, r *CRUD, data *Body, params *QueryParams) (status int, result any, err error)
	DeleteRecord(c *ewa.Context, r *CRUD, params *QueryParams) (status int, result any, err error)
	Audit(action string, c *ewa.Context, r *CRUD)
	Unmarshal(body *Body, contentType string, data []byte) (err error)
}

type IResponse interface {
	Send(c *ewa.Context, state string, status int, data any) error
}

type IQueryParam interface {
	Format(r *CRUD, q *QueryParam) (*QueryParam, error)
	Query(q *QueryParams, columns []string) (string, []any)
	Cast(value string, q *QueryParam) error
	Pattern() string
}

type functions struct{}

func (f functions) Columns(r *CRUD, fields ...string) []string {
	return []string{"id", "name"}
}

func (f functions) SetRecord(c *ewa.Context, r *CRUD, data *Body, params *QueryParams) (int, any, error) {
	return 200, 0, nil
}

func (f functions) GetRecord(c *ewa.Context, r *CRUD, params *QueryParams) (int, Map, error) {
	return 200, nil, nil
}

func (f functions) GetRecords(c *ewa.Context, r *CRUD, params *QueryParams) (int, Maps, int64, error) {
	return 200, nil, 0, nil
}

func (f functions) UpdateRecord(c *ewa.Context, r *CRUD, data *Body, params *QueryParams) (int, any, error) {
	return 200, nil, nil
}

func (f functions) DeleteRecord(c *ewa.Context, r *CRUD, params *QueryParams) (int, any, error) {
	return 200, nil, nil
}

func (f functions) Audit(action string, c *ewa.Context, r *CRUD) {
	fmt.Println(action)
	if c.Identity != nil {
		fmt.Println(c.Identity.Username)
	}
	fmt.Println(r.ModelName)
}

func (f functions) Unmarshal(body *Body, contentType string, data []byte) (err error) {
	switch contentType {
	case "application/json", "application/json;utf-8":
		if body.IsArray {
			return json.Unmarshal(data, &body.Array)
		}
		return json.Unmarshal(data, &body.Data)
	case "application/xml":
		if body.IsArray {
			return xml.Unmarshal(data, &body.Array)
		}
		return xml.Unmarshal(data, &body.Data)
	}
	return nil
}

type PostgresFormat struct{}

const (
	inArray = "&& ARRAY[?]"
)

func (p *PostgresFormat) Pattern() string {
	return `\[(->|->>|>|<|>-|<-|!|<>|array|&&|!array|!&&|~|!~|~\*|!~\*|\+|!\+|%|:|[aA-zZ]+)]$`
}

func (p *PostgresFormat) Format(r *CRUD, q *QueryParam) (*QueryParam, error) {

	switch q.Znak {
	case "!":
		q.Znak = "!="
	case ">-":
		q.Znak = ">="
	case "<-":
		q.Znak = "<="
	case "%":
		q.Znak = "like"
		q.Type = "::text"
	case "!%":
		q.Znak = "not like"
	case "~", "!~", "~*", "!~*":
	case "+":
		q.Znak = "similar to"
	case "!+":
		//q.Type += "::text"
		q.Znak = "not similar to"
	case "->", "->>":
		switch v := q.Value.(type) {
		case string:
			a := strings.Split(v, "=")
			if len(a) == 2 {
				qf := QueryFormat(r, a[0], a[1])
				value := fmt.Sprintf("'%s' %s", qf.Key, qf.Znak)
				switch t := qf.Value.(type) {
				case string:
					value = strings.ReplaceAll(value, "?", "'"+t+"'")
				case []string:
					for i, tt := range t {
						value = strings.Replace(value, "?", "'"+tt+"'", i+1)
					}
				}
				q.Value = value
			}
		}
	case "array", "&&":
		if q.IsArray() {
			q.Znak = inArray
			return p.setTypeArray(q), nil
		}
	case "!array", "!&&":
		if q.IsArray() {
			q.Znak = inArray
			q.Key = fmt.Sprintf(`not "%s"`, q.Key)
			q.IsQuotes = false
			return p.setTypeArray(q), nil
		}
	}

	if q.Value == nil {
		switch q.Znak {
		case "=":
			q.Znak = "is ?"
		case "!=":
			q.Znak = "is not ?"
		}
		return q, nil
	}
	if q.IsArray() {
		switch q.Znak {
		case "=":
			q.Znak = "in(?)"
		case "!=", "<>":
			q.Znak = "not in(?)"
		}
		return q, nil
	}
	if q.IsRange() {
		q.Znak = "between ? and ?"
		return q, nil
	}

	q.Znak += " ?"

	return q, nil
}

func (p *PostgresFormat) setTypeArray(q *QueryParam) *QueryParam {
	if q.Znak == inArray {
		switch q.DataType {
		case "string":
			q.Znak += "::text[]"
		case "int":
			q.Znak += "::int[]"
		case "int64":
			q.Znak += "::bigint[]"
		case "float":
			q.Znak += "::real[]"
		case "float64":
			q.Znak += "::double precision[]"
		case "uint":
			q.Znak += "::serial[]"
		case "uint64":
			q.Znak += "::bigserial[]"
		}
	}
	return q
}

func (*PostgresFormat) Query(q *QueryParams, columns []string) (query string, values []any) {
	var (
		params []*QueryParam
		fields []string
	)
	// Если нет параметров, то выходим
	if q.Len() == 0 && q.ID == nil {
		return "", nil
	}
	// Отдельно передаём поле ID
	if q.ID != nil {
		params = append(params, q.ID)
	}
	// Заполнение параметры адресной строки
	for key, value := range q.m {
		if key == AllFieldsParamName || key == ExtraParamName {
			continue
		}
		for _, v := range value {
			params = append(params, v)
		}
	}

	// Формирование полей для поиска везде OR
	if vals, ok := q.m[AllFieldsParamName]; ok && len(vals) > 0 {
		value := vals[0]
		// Параметр адресной строки *=
		if q.Filter != nil && len(q.Filter.Fields) > 0 {
			for _, field := range q.Filter.Fields {
				if _, ok = q.m[field]; !ok {
					value.Key = field
					if value.IsQuotes {
						value.Key = `"` + value.Key + `"::text`
					}
					fields = append(fields, strings.Trim(fmt.Sprintf("%s %s", value.Key, value.Znak), " "))
					values = append(values, value.Value)
				}
			}
		} else {
			for _, column := range columns {
				if _, ok = q.m[column]; !ok {
					value.Key = column
					if value.IsQuotes {
						value.Key = `"` + value.Key + `"::text`
					}
					fields = append(fields, strings.Trim(fmt.Sprintf("%s %s", value.Key, value.Znak), " "))
					values = append(values, value.Value)
				}
			}
		}
	}
	if len(fields) > 0 {
		for i, field := range fields {
			var spliter string
			if i < len(fields)-1 {
				spliter = " or "
			}
			query += field + spliter
		}
		query = "(" + query + ")"
	}
	// Заполняем строку запроса и значения для неё
	if len(params) > 0 {
		var v string
		for i, param := range params {
			values = append(values, param.Value)
			var spliter string
			if i > 0 {
				spliter = " and "
			}
			if param.IsOR {
				spliter = " or "
			}
			if param.IsQuotes {
				param.Key = `"` + param.Key + `"`
			}
			v += spliter + strings.Trim(fmt.Sprintf("%s %s", param.Key, param.Znak), " ")
		}
		if len(query) > 0 {
			query += " and " + v
		} else {
			query = v
		}
	}

	return
}

// Cast Приведение переменной к типу данных
func (p *PostgresFormat) Cast(value string, q *QueryParam) (err error) {

	if q.DataType == "" {
		switch strings.ToLower(value) {
		case "null":
			q.Value = nil
		case "true", "false":
			q.Value, err = strconv.ParseBool(value)
		default:
			if rng, ok := p.IsRange(q.Znak, value); ok {
				q.Value = rng
				q.Type = RangeType
				break
			}
			if array, ok := p.IsArray(value); ok {
				q.Value = array
				q.Type = ArrayType
				break
			}
			q.Value = value
		}
		return
	}
	var (
		rng, array []string
		ok         bool
	)
	if rng, ok = p.IsRange(q.Znak, value); ok {
		q.Type = RangeType
	}
	if array, ok = p.IsArray(value); ok && !q.IsRange() {
		q.Type = ArrayType
	}
	switch q.DataType {
	case "string":
		switch {
		case q.IsArray():
			q.Value = array
		case q.IsRange():
			q.Value = rng
		default:
			q.Value = value
		}
	case "int":
		switch {
		case q.IsArray():
			q.Value = p.SetInt32Array(array)
		case q.IsRange():
			q.Value = p.SetInt32Array(rng)
		default:
			q.Value, err = strconv.Atoi(value)
		}
	case "int64":
		switch {
		case q.IsArray():
			q.Value = p.SetInt64Array(array)
		case q.IsRange():
			q.Value = p.SetInt64Array(rng)
		default:
			q.Value, err = strconv.ParseInt(value, 10, 64)
		}
	case "float":
		switch {
		case q.IsArray():
			q.Value = p.SetFloat32Array(array)
		case q.IsRange():
			q.Value = p.SetFloat32Array(rng)
		default:
			q.Value, err = strconv.ParseFloat(value, 32)
		}
	case "float64":
		switch {
		case q.IsArray():
			q.Value = p.SetFloat64Array(array)
		case q.IsRange():
			q.Value = p.SetFloat64Array(rng)
		default:
			q.Value, err = strconv.ParseFloat(value, 64)
		}
	case "uint":
		switch {
		case q.IsArray():
			q.Value = p.SetUIntArray(array)
		case q.IsRange():
			q.Value = p.SetUIntArray(rng)
		default:
			q.Value, err = strconv.ParseUint(value, 10, 32)
		}
	case "uint64":
		switch {
		case q.IsArray():
			q.Value = p.SetUInt64Array(array)
		case q.IsRange():
			q.Value = p.SetUInt64Array(rng)
		default:
			q.Value, err = strconv.ParseUint(value, 10, 64)
		}
	case "date":
		if q.IsRange() {
			q.Value = p.SetTimeArray(rng, time.DateOnly)
			break
		}
		q.Value, err = time.Parse(time.DateOnly, value)
	case "time":
		if q.IsRange() {
			q.Value = p.SetTimeArray(rng, time.TimeOnly)
			break
		}
		q.Value, err = time.Parse(time.TimeOnly, value)
	case "datetime":
		if q.IsRange() {
			q.Value = p.SetTimeArray(rng, time.DateTime)
			break
		}
		q.Value, err = time.Parse(time.DateTime, value)
	default:

	}
	if err != nil {
		return fmt.Errorf("invalid datatype %s", q.DataType)
	}
	return nil
}

// IsArray Проверка на массив
func (*PostgresFormat) IsArray(value string) ([]string, bool) {
	rgx := regexp.MustCompile(`^\[(.+)]$`)
	if rgx.MatchString(value) {
		matches := rgx.FindStringSubmatch(value)
		if len(matches) == 2 {
			return strings.Split(matches[1], ","), true
		}
	}
	return nil, false
}

func (*PostgresFormat) IsRange(znak, value string) ([]string, bool) {
	if znak == ":" {
		rgx := regexp.MustCompile(`^\[(.+)\|(.+)]$`)
		if rgx.MatchString(value) {
			matches := rgx.FindStringSubmatch(value)
			if len(matches) == 3 {
				return []string{matches[1], matches[2]}, true
			}
		}
	}
	return nil, false
}

func (*PostgresFormat) SetInt32Array(array []string) (a []int) {
	for _, v := range array {
		if value, err := strconv.Atoi(v); err == nil {
			a = append(a, value)
		}
	}
	return a
}

func (*PostgresFormat) SetInt64Array(array []string) (a []int64) {
	for _, v := range array {
		if value, err := strconv.ParseInt(v, 10, 64); err == nil {
			a = append(a, value)
		}
	}
	return a
}

func (*PostgresFormat) SetUIntArray(array []string) (a []uint64) {
	for _, v := range array {
		if value, err := strconv.ParseUint(v, 10, 32); err == nil {
			a = append(a, value)
		}
	}
	return a
}

func (*PostgresFormat) SetUInt64Array(array []string) (a []uint64) {
	for _, v := range array {
		if value, err := strconv.ParseUint(v, 10, 64); err == nil {
			a = append(a, value)
		}
	}
	return a
}

func (*PostgresFormat) SetFloat32Array(array []string) (a []float32) {
	for _, v := range array {
		if value, err := strconv.ParseFloat(v, 32); err == nil {
			a = append(a, float32(value))
		}
	}
	return a
}

func (*PostgresFormat) SetFloat64Array(array []string) (a []float64) {
	for _, v := range array {
		if value, err := strconv.ParseFloat(v, 64); err == nil {
			a = append(a, value)
		}
	}
	return a
}

func (*PostgresFormat) SetTimeArray(array []string, layout string) (a []time.Time) {
	for _, v := range array {
		if value, err := time.Parse(layout, v); err == nil {
			a = append(a, value)
		}
	}
	return a
}
