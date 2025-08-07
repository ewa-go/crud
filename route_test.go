package crud

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"testing"
	"time"

	"github.com/ewa-go/ewa"
	"github.com/ewa-go/ewa/security"
)

type Handlers struct{}

func (h *Handlers) Columns(r *CRUD, fields ...string) []string {
	fmt.Println("Columns")
	fmt.Printf("tableName: %s\n", r.ModelName)
	return []string{"id", "name"}
}

func (h *Handlers) SetRecord(c *ewa.Context, r *CRUD, data *Body, params *QueryParams) (int, any, error) {
	if data.IsArray {
		fmt.Println(data.Array)
		return 200, "result", nil
	}
	fmt.Println("SetRecord")
	fmt.Printf("tableName: %s\n", r.ModelName)
	fmt.Printf("data: %v\n", data)
	fmt.Printf("params: %v\n", params)
	if c.Identity != nil {
		fmt.Println("author", c.Identity.Username)
	}
	return 200, "result", nil
}

func (h *Handlers) GetRecord(c *ewa.Context, r *CRUD, params *QueryParams) (int, Map, error) {
	fmt.Println("GetRecord")
	if params.ID != nil {
		fmt.Println("ID", params.ID.Value)
	}
	data := Map{
		"id":   1,
		"name": "Name",
	}
	return 200, data, nil
}

func (h *Handlers) GetRecords(c *ewa.Context, r *CRUD, params *QueryParams) (int, Maps, int64, error) {
	fmt.Println("GetRecords")
	query, values := r.Query(params, h.Columns(r))
	fmt.Println("query:", query)
	fmt.Printf("values: %v\n", values)
	data := Maps{}
	data = append(data, Map{"id": 1, "name": "Name"})
	data = append(data, Map{"id": 2, "name": "Name2"})
	return 200, data, 2, nil
}

func (h *Handlers) UpdateRecord(c *ewa.Context, r *CRUD, data *Body, params *QueryParams) (int, any, error) {
	fmt.Println("UpdateRecord")
	fmt.Printf("tableName: %s\n", r.ModelName)
	fmt.Printf("data: %v\n", data)
	fmt.Printf("params: %v\n", params)
	if c.Identity != nil {
		fmt.Println("author", c.Identity.Username)
	}
	return 200, "result", nil
}

func (h *Handlers) DeleteRecord(c *ewa.Context, r *CRUD, params *QueryParams) (int, any, error) {
	fmt.Println("DeleteRecord")
	fmt.Printf("tableName: %s\n", r.ModelName)
	fmt.Printf("params: %v\n", params)
	return 200, "result", nil
}

func (h *Handlers) Audit(action string, c *ewa.Context, r *CRUD) {
	fmt.Println(action)
}

func (h *Handlers) Unmarshal(body *Body, contentType string, data []byte) (err error) {
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

var (
	h = new(Handlers)
)

func TestSetModelName(t *testing.T) {
	crud := New(h).SetModelName("table")
	fmt.Println(crud.ModelName)
}

func TestFunctions_Unmarshal(t *testing.T) {
	crud := New(h).SetModelName("table")
	body := NewBody("id").SetIsArray(false)
	err := crud.Unmarshal(body, "application/json", []byte(`{"name": "Name"}`))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(body.Data)
	body.SetIsArray(true)
	err = crud.Unmarshal(body, "application/json", []byte(`[{"name": "Name1"},{"name": "Name2"}]`))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(body.Array)
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
					status, data, err := r.GetRecord(c, r, queryParams)
					if err != nil {
						return c.SendString(status, err.Error())
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
