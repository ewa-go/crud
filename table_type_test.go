package crud

import (
	"fmt"
	"testing"
)

func TestNewTableType_Get(t *testing.T) {
	crud := New(nil).SetHeader("table", "schema.table", true).
		SetHeader("view", "schema.view")
	fmt.Println(crud.TableTypes)
	fmt.Println(crud.TableTypes.Default())
	crud.SetHeader("any", "any", true)
	fmt.Println(crud.TableTypes)
	fmt.Println(crud.TableTypes.Default())
}
