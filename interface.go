package crud

import "fmt"

type IAudit interface {
	String(statusCode int, data string) (int, string)
	JSON(statusCode int, v any) (int, any)
	Insert(action, tableName string, identity Identity, path string)
}

type IHandlers interface {
	Columns(tableName string, fields ...string) []string
	SetRecord(tableName string, data Map, params *QueryParams) (any, error)
	GetRecord(tableName string, params *QueryParams) (Map, error)
	GetRecords(tableName string, params *QueryParams) (Maps, int64, error)
	UpdateRecord(tableName string, data Map, params *QueryParams) error
	DeleteRecord(tableName string, params *QueryParams) error
}

type IResponse interface {
	Read(id any, data any, err error) any
	Created(id interface{}, err error) any
	Updated(id interface{}, err error) any
	Deleted(id interface{}, err error) any
}

type audit struct{}

func (a audit) String(statusCode int, data string) (int, string) {
	return statusCode, data
}

func (a audit) JSON(statusCode int, v any) (int, any) {
	return statusCode, v
}

func (a audit) Insert(tableName, action string, identity Identity, path string) {
	fmt.Printf("%s %s [%s] %s %s", identity.Datetime, tableName, action, identity.Author, path)
}

type functions struct{}

func (f functions) Columns(tableName string, fields ...string) []string {
	return nil
}

func (f functions) SetRecord(tableName string, data Map, params *QueryParams) (uint, error) {
	return 0, nil
}

func (f functions) GetRecord(tableName string, params *QueryParams) (Map, error) {
	return nil, nil
}

func (f functions) GetRecords(tableName string, params *QueryParams) (Maps, int64, error) {
	return nil, 0, nil
}

func (f functions) UpdateRecord(tableName string, data Map, params *QueryParams) error {
	return nil
}

func (f functions) DeleteRecord(tableName string, params *QueryParams) error {
	return nil
}
