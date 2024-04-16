package crud

import (
	"fmt"
	"testing"
)

func TestParams(t *testing.T) {
	q := QueryParams{}
	q.ID = QueryFormat("id", "1")
	q.set("name", QueryFormat("name", "Name"))
	query := q.GetQuery(h.Columns("table"))
	fmt.Println(query)

	q = QueryParams{}
	q.ID = QueryFormat("id", "2")
	q.set("*", QueryFormat("*", "Значение"))
	query = q.GetQuery(h.Columns("table"))
	fmt.Println(query)
}
