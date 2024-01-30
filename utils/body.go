package utils

import (
	"encoding/json"
	"errors"
)

type Body struct {
	Data           map[string]interface{}
	Array          []map[string]interface{}
	IsArray        bool
	FieldIDName    string
	Fields         []string
	ExecuteHandler ExecuteHandler
}

type ExecuteHandler func(data interface{}) error

const (
	fieldComment  = "comment"
	fieldAuthor   = "author"
	fieldCreated  = "created"
	fieldModified = "modified"
	fieldDatetime = "datetime"
)

func NewBody(fieldIDName string, fields []string) *Body {
	return &Body{
		Data:        map[string]interface{}{},
		FieldIDName: fieldIDName,
		Fields:      fields,
	}
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

func (b *Body) Set(key string, value interface{}) *Body {
	b.Data[key] = value
	return b
}

func (b *Body) Get(key string) interface{} {
	if value, ok := b.Data[key]; ok {
		return value
	}
	return nil
}

func (b *Body) SetAuthor(value interface{}) *Body {
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
}

func (b *Body) Id() interface{} {
	if id, ok := b.Data[b.FieldIDName]; ok {
		return id
	}
	return 0
}

func (b *Body) ToMap() map[string]interface{} {
	return b.Data
}
