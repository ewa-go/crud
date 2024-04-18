package crud

import (
	"fmt"
	"testing"
)

func TestNewTableType_Get(t *testing.T) {
	crud := New(nil).SetTableType("table", "schema.table", true).
		SetTableType("view", "schema.view")
	fmt.Println(crud.TableTypes)
	fmt.Println(crud.TableTypes.Default())
	crud.SetTableType("any", "any", true)
	fmt.Println(crud.TableTypes)
	fmt.Println(crud.TableTypes.Default())
}
