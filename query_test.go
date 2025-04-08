package crud

import (
	"fmt"
	"testing"
	"time"
)

func assertEq(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("Values did not match, a: %v, b: %v\n", a, b)
	}
}

func assertArrayEq(t *testing.T, a []interface{}, b []interface{}) {
	if len(a) != len(b) {
		t.Fatalf("Values did not match, a: %v, b: %v\n", a, b)
	}
	for i, aa := range a {
		switch at := aa.(type) {
		case []string, []int, []time.Time:
			assertArrayStringEq(t, at, b[i])
		default:
			assertEq(t, aa, b[i])
		}
	}
}

func assertArrayStringEq(t *testing.T, a interface{}, b interface{}) {
	if fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b) {
		t.Fatalf("Values did not match, a: %v, b: %v\n", a, b)
	}
}

func getCRUD() *CRUD {
	return New(new(functions))
}

func TestName(t *testing.T) {
	var a any
	a = []string{"a", "b", "c"}
	fmt.Println(a)
}

func TestParams(t *testing.T) {
	r := getCRUD()
	q := &QueryParams{}
	q.ID = QueryFormat(r, "id", "1::int")
	q.Set("name", QueryFormat(r, "name", "Name"))
	query, values := r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "name" = ?`)
	assertArrayEq(t, []any{1, "Name"}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "2::int")
	q.Set("*", QueryFormat(r, "*", "Значение"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `("id" = ? or "name" = ?) and "id" = ?`)
	assertArrayEq(t, []any{"Значение", "Значение", 2}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "3::int")
	q.Set("name", QueryFormat(r, "name", "[1,2,3]"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "name" in(?)`)
	assertArrayEq(t, []any{3, []string{"1", "2", "3"}}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "4::int")
	q.Set("name", QueryFormat(r, "name[:]", "[2024-08-01 00:00:00|2024-08-31 23:59:59]::datetime"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "name" between ? and ?`)
	from, err := time.Parse(time.DateTime, "2024-08-01 00:00:00")
	if err != nil {
		t.Fatal(err)
	}
	to, err := time.Parse(time.DateTime, "2024-08-31 23:59:59")
	if err != nil {
		t.Fatal(err)
	}
	assertArrayEq(t, []any{4, []time.Time{from, to}}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "5::int")
	q.Set("name", QueryFormat(r, "name[array]", "[success,warning]"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "name" && ARRAY[?]`)
	assertArrayEq(t, []any{5, []string{"success", "warning"}}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "6::int")
	q.Set("name", QueryFormat(r, "name[&&]", "[success,warning]"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "name" && ARRAY[?]`)
	assertArrayEq(t, []any{6, []string{"success", "warning"}}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "7::int")
	q.Set("name", QueryFormat(r, "name[!array]", "[success,warning]"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and not "name" && ARRAY[?]`)
	assertArrayEq(t, []any{7, []string{"success", "warning"}}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "8")
	q.Set("name", QueryFormat(r, "name[!&&]", "[success,warning]"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and not "name" && ARRAY[?]`)
	assertArrayEq(t, []any{"8", []string{"success", "warning"}}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "9::int")
	q.Set("name", QueryFormat(r, "name[!]", "[1,2,4]::int"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "name" not in(?)`)
	assertArrayEq(t, []any{9, []int{1, 2, 4}}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "10::int")
	q.Set("name", QueryFormat(r, "name", "Name1"))
	q.Set("name", QueryFormat(r, "name[~]", "Name2"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "name" = ? and "name" ~ ?`)
	assertArrayEq(t, []any{10, "Name1", "Name2"}, values)
}

func TestOR(t *testing.T) {
	var err error
	r := getCRUD()
	q := &QueryParams{}
	q.ID, err = r.QueryFormat("id", "1::int")
	if err != nil {
		t.Fatal(err)
	}
	q.Set("name", QueryFormat(r, "name", "Name"))
	q.Set("index", QueryFormat(r, "[|]index", "2::int"))
	query, values := r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "name" = ? or "index" = ?`)
	assertArrayEq(t, []any{1, "Name", 2}, values)
}

func TestJSON(t *testing.T) {
	r := getCRUD()
	q := &QueryParams{}
	q.ID = QueryFormat(r, "id", "1::int")
	q.Set("result", QueryFormat(r, "result[->>]", "type=2"))
	query, values := r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "result" ->> ?`)
	assertArrayEq(t, []any{1, "'type' = '2'"}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "2::int")
	q.Set("result", QueryFormat(r, "result[->>]", "type[%]=2%"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "result" ->> ?`)
	assertArrayEq(t, []any{2, "'type' like '2%'"}, values)

	q = &QueryParams{}
	q.ID = QueryFormat(r, "id", "3::int")
	q.Set("result", QueryFormat(r, "result[->>]", "type[:]=[1|2]"))
	query, values = r.Query(q, r.Columns(r))
	assertEq(t, query, `"id" = ? and "result" ->> ?`)
	assertArrayEq(t, []any{3, "'type' between '1' and '2'"}, values)
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

func TestCast(t *testing.T) {
	r := getCRUD()
	q := QueryParam{}
	var err error
	err = r.Cast("text", &q)
	if err != nil {
		t.Fatal(err)
	}
	q = QueryParam{DataType: "int"}
	err = r.Cast("1", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, 1)
	q = QueryParam{DataType: "int64"}
	err = r.Cast("-123456789987654", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, int64(-123456789987654))
	q = QueryParam{DataType: "float"}
	err = r.Cast("1234.5678", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, 1234.5677490234375)
	q = QueryParam{DataType: "float64"}
	err = r.Cast("-1234.5678", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, -1234.5678)
	q = QueryParam{DataType: "uint"}
	err = r.Cast("123456789", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, uint64(123456789))
	q = QueryParam{DataType: "uint64"}
	err = r.Cast("123456789123321654", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, uint64(123456789123321654))
	q = QueryParam{DataType: "date"}
	err = r.Cast("2025-03-28", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value.(time.Time).Format(time.DateOnly), "2025-03-28")
	q = QueryParam{DataType: "time"}
	err = r.Cast("23:52:12", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value.(time.Time).Format(time.TimeOnly), "23:52:12")
	q = QueryParam{DataType: "datetime"}
	err = r.Cast("2025-03-28 23:52:12", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value.(time.Time).Format(time.DateTime), "2025-03-28 23:52:12")
	q = QueryParam{}
	err = r.Cast("true", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, true)
	q = QueryParam{DataType: "string"}
	err = r.Cast("false", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, "false")
	q = QueryParam{}
	err = r.Cast("null", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, nil)
	q = QueryParam{DataType: "string"}
	err = r.Cast("null", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertEq(t, q.Value, "null")
	q = QueryParam{}
	err = r.Cast("[t1,t2,t3]", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertArrayStringEq(t, q.Value, []string{"t1", "t2", "t3"})
	q = QueryParam{DataType: "string"}
	err = r.Cast("[t1,t2,t3]", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertArrayStringEq(t, q.Value, []string{"t1", "t2", "t3"})
	q = QueryParam{DataType: "int"}
	err = r.Cast("[1,2,3]", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertArrayStringEq(t, q.Value, []int{1, 2, 3})
	q = QueryParam{DataType: "int", Znak: ":"}
	err = r.Cast("[1|10]", &q)
	if err != nil {
		t.Fatal(err)
	}
	assertArrayStringEq(t, q.Value, []int{1, 10})
}
