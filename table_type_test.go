package crud

import (
	"fmt"
	"testing"
)

var crud = New(nil).SetTableTypeTable("schema.table", true).
	SetTableTypeView("schema.view").SetTableType("any", "any")

func TestNewTableType_Get(t *testing.T) {
	fmt.Println(crud.TableTypes.Get("table"))
	fmt.Println(crud.TableTypes.Get("view"))
	fmt.Println(crud.TableTypes.Get("any"))
	// Если передать пустой заголовок, то вернётся значение по-умолчанию
	fmt.Println(crud.TableTypes.Get(""))
}

func TestTableType_Default(t *testing.T) {
	fmt.Println(crud.TableTypes.Default())
}
