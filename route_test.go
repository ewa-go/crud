package crud

import (
	"fmt"
	"github.com/ewa-go/ewa"
	"github.com/ewa-go/ewa/security"
	"testing"
	"time"
)

type Handlers struct{}

func (c *Handlers) Columns(r *CRUD, fields ...string) []string {
	fmt.Println("Columns")
	fmt.Printf("tableName: %s\n", r.ModelName)
	return []string{"id", "name"}
}

func (c *Handlers) SetRecord(r *CRUD, data *Body, params *QueryParams) (any, error) {
	if data.IsArray {
		fmt.Println(data.Array)
		return "result", nil
	}
	fmt.Println("SetRecord")
	fmt.Printf("tableName: %s\n", r.ModelName)
	fmt.Printf("data: %v\n", data)
	fmt.Printf("params: %v\n", params)
	return "result", nil
}

func (c *Handlers) GetRecord(r *CRUD, params *QueryParams) (Map, error) {
	fmt.Println("GetRecord")
	if params.ID != nil {
		fmt.Println("ID", params.ID.Value)
	}
	data := Map{
		"id":   1,
		"name": "Name",
	}
	return data, nil
}

func (c *Handlers) GetRecords(r *CRUD, params *QueryParams) (Maps, int64, error) {
	fmt.Println("GetRecords")
	query, values := r.Query(params, c.Columns(r))
	fmt.Println("query:", query)
	fmt.Printf("values: %v\n", values)
	data := Maps{}
	data = append(data, Map{"id": 1, "name": "Name"})
	data = append(data, Map{"id": 2, "name": "Name2"})
	return data, 2, nil
}

func (c *Handlers) UpdateRecord(r *CRUD, data *Body, params *QueryParams) (any, error) {
	fmt.Println("UpdateRecord")
	fmt.Printf("tableName: %s\n", r.ModelName)
	fmt.Printf("data: %v\n", data)
	fmt.Printf("params: %v\n", params)
	return "result", nil
}

func (c *Handlers) DeleteRecord(r *CRUD, params *QueryParams) (any, error) {
	fmt.Println("DeleteRecord")
	fmt.Printf("tableName: %s\n", r.ModelName)
	fmt.Printf("params: %v\n", params)
	return "result", nil
}

func (c *Handlers) Audit(action string, ctx *ewa.Context, r *CRUD) {
	fmt.Println(action)
}

var (
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
					defer r.Audit(Read, c, r)
					queryParams, err := r.NewQueryParams(c, false)
					if err != nil {
						return c.SendString(400, err.Error())
					}
					data, err := r.GetRecord(r, queryParams)
					if err != nil {
						return c.SendString(400, err.Error())
					}
					return c.JSON(200, data)
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
