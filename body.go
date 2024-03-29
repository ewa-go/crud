package crud

import (
	"encoding/json"
	"errors"
)

type Body struct {
	Data           map[string]interface{}
	Array          []map[string]interface{}
	IsArray        bool
	FieldIDName    string
	Fields         Fields
	ExecuteHandler ExecuteHandler
}

type Field struct {
	Key   string
	Value any
}

type Fields []Field

type ExecuteHandler func(data interface{}) error

func NewFields(key string, value any) Fields {
	return Fields{{key, value}}
}

func (f Fields) SetField(key string, value any) Fields {
	f = append(f, Field{key, value})
	return f
}

func NewBody(fieldIDName string, fields ...Field) *Body {
	return &Body{
		Data:        map[string]interface{}{},
		FieldIDName: fieldIDName,
		Fields:      fields,
	}
}

func (b *Body) SetFields(fields ...Field) *Body {
	b.Fields = append(b.Fields, fields...)
	return b
}

func (b *Body) SetExecuteHandler(h ExecuteHandler) *Body {
	b.ExecuteHandler = h
	return b
}

func (b *Body) Unmarshal(data []byte, isArray bool) error {
	if len(data) == 0 {
		return errors.New("пустые данные")
	}
	b.IsArray = isArray
	if isArray {
		return json.Unmarshal(data, &b.Array)
	}
	return json.Unmarshal(data, &b.Data)
}

func (b *Body) Execute(skipError bool) error {
	if b.IsArray {
		for _, data := range b.Array {
			if err := b.ExecuteHandler(data); err != nil {
				if skipError {
					continue
				}
				return err
			}
		}
		return nil
	}
	return b.ExecuteHandler(b.Data)
}

func (b *Body) SetField(key string, value any) *Body {
	b.Data[key] = value
	return b
}

func (b *Body) GetField(key string) any {
	if value, ok := b.Data[key]; ok {
		return value
	}
	return nil
}

func (b *Body) SetArrayField(i int, key string, value any) *Body {
	b.Array[i][key] = value
	return b
}

func (b *Body) GetArrayField(i int, key string) any {
	if value, ok := b.Array[i][key]; ok {
		return value
	}
	return nil
}

/*func (b *Body) SetAuthor(value interface{}) *Body {
	return b.Set(fieldAuthor, value)
}

func (b *Body) Created() *Body {
	return b.Set(fieldCreated, UTC())
}

func (b *Body) Modified() *Body {
	return b.Set(fieldModified, UTC())
}

func (b *Body) Datetime() *Body {
	return b.Set(fieldDatetime, UTC())
}

func (b *Body) Comment(comment string) *Body {
	return b.Set(fieldComment, comment)
}*/

/*func (b *Body) Id() interface{} {
	if id, ok := b.Data[b.FieldIDName]; ok {
		return id
	}
	return 0
}*/

func (b *Body) ToArrayMap(i int) (data map[string]interface{}) {
	if b == nil {
		return nil
	}
	data = b.Array[i]
	for _, field := range b.Fields {
		if data != nil {
			data[field.Key] = field.Value
		}
	}
	return
}

func (b *Body) ToMap() map[string]interface{} {
	if b == nil {
		return nil
	}
	for _, field := range b.Fields {
		if b.Data != nil {
			b.Data[field.Key] = field.Value
		}
	}
	return b.Data
}
