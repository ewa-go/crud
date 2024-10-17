package crud

import (
	"fmt"
	"github.com/ewa-go/ewa"
	"github.com/ewa-go/ewa/security"
	"testing"
	"time"
)

type Audit struct {
	Action string
	Table  string
	Author string
	Path   string
}

func (a *Audit) String(statusCode int, data string) (int, string) {
	return statusCode, data
}

func (a *Audit) JSON(statusCode int, v any) (int, any) {
	return statusCode, v
}

func (a *Audit) Take(statusCode int, v any) {}

func (a *Audit) Exec(action string, c *ewa.Context, r *CRUD) {
	//a.Table = tableName
	a.Action = action
	a.Author = c.Identity.Username
	a.Path = c.Path()
}

type Handlers struct{}

func (c *Handlers) Columns(tableName string, fields ...string) []string {
	fmt.Println("Columns")
	fmt.Printf("tableName: %s\n", tableName)
	return []string{"id", "name"}
}

func (c *Handlers) SetRecord(tableName string, data *Body, params *QueryParams) (any, error) {
	if data.IsArray {
		fmt.Println(data.Array)
		return "result", nil
	}
	fmt.Println("SetRecord")
	fmt.Printf("tableName: %s\n", tableName)
	fmt.Printf("data: %v\n", data)
	fmt.Printf("params: %v\n", params)
	return "result", nil
}

func (c *Handlers) GetRecord(tableName string, params *QueryParams) (Map, error) {
	fmt.Println("GetRecord")
	if params.ID != nil {
		fmt.Println("ID", params.ID.String())
	}
	data := Map{
		"id":   1,
		"name": "Name",
	}
	return data, nil
}

func (c *Handlers) GetRecords(tableName string, params *QueryParams) (Maps, int64, error) {
	fmt.Println("GetRecords")
	fmt.Println("query", params.GetQuery(c.Columns(tableName)))
	data := Maps{}
	data = append(data, Map{"id": 1, "name": "Name"})
	data = append(data, Map{"id": 2, "name": "Name2"})
	return data, 2, nil
}

func (c *Handlers) UpdateRecord(tableName string, data *Body, params *QueryParams) (any, error) {
	fmt.Println("UpdateRecord")
	fmt.Printf("tableName: %s\n", tableName)
	fmt.Printf("data: %v\n", data)
	fmt.Printf("params: %v\n", params)
	return "result", nil
}

func (c *Handlers) DeleteRecord(tableName string, params *QueryParams) (any, error) {
	fmt.Println("DeleteRecord")
	fmt.Printf("tableName: %s\n", tableName)
	fmt.Printf("params: %v\n", params)
	return "result", nil
}

var (
	a = new(Audit)
	h = new(Handlers)
)

func TestSetModelName(t *testing.T) {
	crud := New(h).SetModelName("table")
	fmt.Println(crud.ModelName)
}

func TestGet(t *testing.T) {

	route := &ewa.Route{
		Handler: func(c *ewa.Context) error {
			return New(h).
				SetIAudit(a).
				SetModelName("table").
				SetFieldIdName("id").
				ReadHandler(c, nil, nil)
		},
	}

	ctx := &ewa.Context{
		Identity: &security.Identity{
			Username: "username",
			Datetime: time.Now(),
		},
	}

	if err := route.Handler(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestCustomHandler(t *testing.T) {

	route := &ewa.Route{
		Handler: func(c *ewa.Context) error {
			return New(h).
				SetModelName("table").
				SetFieldIdName("id").
				CustomHandler(c, func(c *ewa.Context, r *CRUD) error {
					defer r.Exec(Read, c, r)
					return c.SendString(200, "OK")
				})
		},
	}

	ctx := &ewa.Context{
		Identity: &security.Identity{
			Username: "username",
			Datetime: time.Now(),
		},
	}

	if err := route.Handler(ctx); err != nil {
		t.Fatal(err)
	}
}
