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

func (a *Audit) Insert(tableName, action string, identity Identity, path string) {
	a.Table = tableName
	a.Action = action
	a.Author = identity.Author
	a.Path = path
}

type Func struct{}

func (c *Func) Columns(tableName string, fields ...string) []string {
	fmt.Println("Columns")
	fmt.Printf("tableName: %s\n", tableName)
	return []string{"id", "name"}
}

func (c *Func) SetRecord(tableName string, data Map, params *QueryParams) (uint, error) {
	fmt.Println("SetRecord")
	fmt.Printf("tableName: %s\n", tableName)
	fmt.Printf("data: %v\n", data)
	fmt.Printf("params: %v\n", params)
	return 0, nil
}

func (c *Func) GetRecord(tableName string, params *QueryParams) (Map, error) {
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

func (c *Func) GetRecords(tableName string, params *QueryParams) (Maps, int64, error) {
	fmt.Println("GetRecords")
	fmt.Println("query", params.GetQuery([]string{"name"}, c.Columns(tableName)))
	data := Maps{}
	data = append(data, Map{"id": 1, "name": "Name"})
	data = append(data, Map{"id": 2, "name": "Name2"})
	return data, 2, nil
}

func (c *Func) UpdateRecord(tableName string, data Map, params *QueryParams) error {
	fmt.Println("UpdateRecord")
	fmt.Printf("tableName: %s\n", tableName)
	fmt.Printf("data: %v\n", data)
	fmt.Printf("params: %v\n", params)
	return nil
}

func (c *Func) DeleteRecord(tableName string, params *QueryParams) error {
	fmt.Println("DeleteRecord")
	fmt.Printf("tableName: %s\n", tableName)
	fmt.Printf("params: %v\n", params)
	return nil
}

var (
	a = new(Audit)
	f = new(Func)
)

func TestGet(t *testing.T) {

	route := &ewa.Route{
		Handler: func(c *ewa.Context) error {
			return NewCRUD(f, a).
				SetModelName("table").
				SetFieldIdName("id").
				ReadHandler(c)
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
			return NewCRUD(f, a).
				SetModelName("table").
				SetFieldIdName("id").
				CustomHandler(c, func(c *ewa.Context, r CRUD) error {
					identity := NewIdentity(c.Identity)
					defer r.Audit.Insert(r.ModelName, Read, identity, c.Path())
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
