package crud

import (
	"fmt"
	"testing"
)

func TestParams(t *testing.T) {
	q := QueryParams{}
	q.ID = QueryFormat("id", "1")
	q.Set("name", QueryFormat("name", "Name"))
	query := q.GetQuery(h.Columns("table"))
	fmt.Println(query)

	q = QueryParams{}
	q.ID = QueryFormat("id", "2")
	q.Set("*", QueryFormat("*", "Значение"))
	query = q.GetQuery(h.Columns("table"))
	fmt.Println(query)

	q = QueryParams{}
	q.ID = QueryFormat("id", "2")
	q.Set("name", QueryFormat("name", "[1,2,4]"))
	query = q.GetQuery(h.Columns("table"))
	fmt.Println(query)
}
