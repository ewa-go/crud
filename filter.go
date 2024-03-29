package crud

import (
	"encoding/json"
)

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

func (r Route) GetRecord(f Filter, query string, args ...interface{}) (Map, error) {
	return r.CRUD.GetRecord(r.ModelName, f, query, args...)
	//data, err = db.DB(r.ModelName).Select(f.Fields...).OrderBy(f.Orders...).Limit(f.Limit).Offset(f.Offset).GetMapRecord(query, args...)
}

func (r Route) GetRecords(f Filter, query string, args ...interface{}) (Maps, int64, error) {
	return r.CRUD.GetRecords(r.ModelName, f, query, args...)
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
