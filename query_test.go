package crud

import (
	"testing"
)

func assertEq(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("Values did not match, a: %v, b: %v\n", a, b)
	}
}

func TestParams(t *testing.T) {
	q := QueryParams{}
	q.ID = QueryFormat("id", "1")
	q.Set("name", QueryFormat("name", "Name"))
	query := q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '1' and "name" = 'Name'`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "2")
	q.Set("*", QueryFormat("*", "Значение"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `("id" = 'Значение' or "name" = 'Значение') and "id" = '2'`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "3")
	q.Set("name", QueryFormat("name", "[1,2,4]"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '3' and "name" in('1','2','4')`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "4")
	q.Set("name", QueryFormat("name[:]", "[01-08-2024:31-08-2024]"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '4' and "name" between '01-08-2024' and '31-08-2024'`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "5")
	q.Set("name", QueryFormat("name[array]", "[success,warning]"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '5' and "name" && ARRAY['success','warning']`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "6")
	q.Set("name", QueryFormat("name[&&]", "[success,warning]"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '6' and "name" && ARRAY['success','warning']`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "7")
	q.Set("name", QueryFormat("name[!array]", "[success,warning]"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '7' and not "name" && ARRAY['success','warning']`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "8")
	q.Set("name", QueryFormat("name[!&&]", "[success,warning]"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '8' and not "name" && ARRAY['success','warning']`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "9")
	q.Set("name", QueryFormat("name[!]", "[1,2,4]"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '9' and "name" not in('1','2','4')`)
}

func TestJSON(t *testing.T) {
	q := QueryParams{}
	q.ID = QueryFormat("id", "5")
	q.Set("result", QueryFormat("result[->>]", "type=2"))
	query := q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '5' and "result" ->> 'type' = '2'`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "6")
	q.Set("result", QueryFormat("result[->>]", "type[%]=2%"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '6' and "result" ->> 'type' like '2%'`)

	q = QueryParams{}
	q.ID = QueryFormat("id", "6")
	q.Set("result", QueryFormat("result[->>]", "type[:]=[1:2]"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '6' and "result" ->> 'type' between '1' and '2'`)
}

func TestFilter(t *testing.T) {
	data := `{
		"fields": [],
		"orders": [],
		"limit": 10,
		"offset": 0,
		"vars": {
			"is_server_member_role": true,
			"is_server_group": false
		}
	}`
	f, err := NewFilter([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, f.GetVar("is_server_member_role"), true)
	assertEq(t, f.GetVar("is_server_group"), false)
	assertEq(t, f.GetVar("any"), nil)
}
