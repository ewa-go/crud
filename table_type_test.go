package crud

import (
	"fmt"
	"testing"
)

var tt = New(nil).TableTypes.
	Add("table", "schema.table", true).
	Add("view", "schema.view")

func TestNewTableType_Get(t *testing.T) {
	fmt.Println(tt)
	fmt.Println(tt.Default())
	ttt := tt.Add("any", "any", true)
	fmt.Println(ttt)
	fmt.Println(ttt.Default())
}
