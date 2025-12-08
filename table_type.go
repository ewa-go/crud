package crud

type TableType struct {
	Key, Value string
	IsDefault  bool
}

type TableTypes []TableType

// Add Добавить тип таблицы
func (tt TableTypes) Add(key, value string, isDefault ...bool) TableTypes {
	var d bool
	if len(isDefault) > 0 {
		d = isDefault[0]
	}
	return append(tt, TableType{key, value, d})
}

// Default Вернуть тип таблицы по-умолчанию
func (tt TableTypes) Default() string {
	for _, hv := range tt {
		if hv.IsDefault {
			return hv.Value
		}
	}
	return ""
}

// Get Извлечь из заголовков тип таблицы.
// Если заголовок не найден, то возвращается тип таблицы по-умолчанию
func (tt TableTypes) Get(header string) (s TableType) {
	for _, t := range tt {
		if t.Key == header {
			return t
		}
		if t.IsDefault {
			s = t
		}
	}
	return
}
