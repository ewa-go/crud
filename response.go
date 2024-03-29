package crud

import "time"

type Response struct {
	Id       any       `json:"id" jsonschema:"type=integer"`
	OK       bool      `json:"ok"`
	State    string    `json:"state"`
	Datetime time.Time `json:"datetime"`
	Message  string    `json:"message"`
}

// NewResponse Ответ для обычных таблиц
func NewResponse(id interface{}, state, message string, ok bool) Response {
	return Response{
		Id:       id,
		OK:       ok,
		State:    state,
		Datetime: time.Now(),
		Message:  message,
	}
}

func NewCreated(id interface{}, message string, ok bool) Response {
	return NewResponse(id, Created, message, ok)
}

func NewUpdated(id interface{}, message string, ok bool) Response {
	return NewResponse(id, Updated, message, ok)
}

func NewDeleted(id interface{}, message string, ok bool) Response {
	return NewResponse(id, Deleted, message, ok)
}
