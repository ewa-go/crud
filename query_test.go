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
	assertEq(t, query, `"id" = '3' and "name" in('1','2','4') `)

	q = QueryParams{}
	q.ID = QueryFormat("id", "4")
	q.Set("name", QueryFormat("name[:]", "[01-08-2024:31-08-2024]"))
	query = q.GetQuery(h.Columns("table"))
	assertEq(t, query, `"id" = '4' and "name" between '01-08-2024' and '31-08-2024'`)
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
