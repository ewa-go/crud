package crud

import (
	"encoding/json"
	"fmt"
	"github.com/ewa-go/ewa"
	"time"
)

type IAudit interface {
	Take(statusCode int, v any)
	Exec(action string, c *ewa.Context, r *CRUD)
}

type IHandlers interface {
	Columns(tableName string, fields ...string) []string
	SetRecord(tableName string, data *Body, params *QueryParams) (any, error)
	GetRecord(tableName string, params *QueryParams) (Map, error)
	GetRecords(tableName string, params *QueryParams) (Maps, int64, error)
	UpdateRecord(tableName string, data *Body, params *QueryParams) (any, error)
	DeleteRecord(tableName string, params *QueryParams) (any, error)
}

type IResponse interface {
	Send(c *ewa.Context, state string, status int, data any) error
}

type audit struct {
	Author, TableName, Path, Body, Action string
	Datetime                              time.Time
	Status                                int
}

func (a *audit) String() string {
	return fmt.Sprintf("%s %s [%s] %s %d %s", a.Datetime, a.TableName, a.Action, a.Author, a.Status, a.Path)
}

func (a *audit) Take(statusCode int, v any) {
	body, _ := json.Marshal(v)
	a.Body = string(body)
	a.Status = statusCode
}

func (a *audit) Exec(action string, c *ewa.Context, r *CRUD) {
	a.TableName = r.ModelName
	a.Action = action
	if c.Identity != nil {
		a.Author = c.Identity.Username
	}
	a.Path = c.Path()
	a.InsertToDB()
}

func (a *audit) InsertToDB() {
	fmt.Println(a)
}

type functions struct{}

func (f functions) Columns(tableName string, fields ...string) []string {
	return nil
}

func (f functions) SetRecord(tableName string, data *Body, params *QueryParams) (any, error) {
	return 0, nil
}

func (f functions) GetRecord(tableName string, params *QueryParams) (Map, error) {
	return nil, nil
}

func (f functions) GetRecords(tableName string, params *QueryParams) (Maps, int64, error) {
	return nil, 0, nil
}

func (f functions) UpdateRecord(tableName string, data *Body, params *QueryParams) (any, error) {
	return nil, nil
}

func (f functions) DeleteRecord(tableName string, params *QueryParams) (any, error) {
	return nil, nil
}
