package crud

import (
	"testing"
)

var tt = NewTableTypes().
	Add("table", "schema.table", true).
	Add("table", "schema.table")

func TestNewTableType_Get(t *testing.T) {

}
