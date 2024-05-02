package crud

type TableType struct {
	Key, Value string
	IsDefault  bool
}

type TableTypes []TableType

// Add Добавить тип таблицы
func (h TableTypes) Add(key, value string, isDefault ...bool) TableTypes {
	var d bool
	if len(isDefault) > 0 {
		d = isDefault[0]
	}
	return append(h, TableType{key, value, d})
}

// Default Вернуть тип таблицы по-умолчанию
func (h TableTypes) Default() string {
	for _, hv := range h {
		if hv.IsDefault {
			return hv.Value
		}
	}
	return ""
}

// Get Извлечь из заголовков тип таблицы.
// Если заголовок не найден, то возвращается тип таблицы по-умолчанию
func (h TableTypes) Get(header string) (s string) {
	for _, hh := range h {
		if hh.Key == header {
			return hh.Value
		}
		if hh.IsDefault {
			s = hh.Value
		}
	}
	return
}
