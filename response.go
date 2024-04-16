package crud

import "time"

type Response struct {
	Id       any       `json:"id" jsonschema:"type=integer"`
	OK       bool      `json:"ok"`
	State    string    `json:"state"`
	Datetime time.Time `json:"datetime"`
	Message  string    `json:"message"`
	Data     any       `json:"data,omitempty"`
}

func (r Response) Read(id any, message string, data any, ok bool) any {
	r.Id = id
	r.State = Read
	r.Datetime = time.Now()
	r.Message = message
	r.OK = ok
	r.Data = data
	return r
}

func (r Response) Created(id interface{}, message string, ok bool) any {
	r.Id = id
	r.State = Created
	r.Datetime = time.Now()
	r.Message = message
	r.OK = ok
	return r
}

func (r Response) Updated(id interface{}, message string, ok bool) any {
	r.Id = id
	r.State = Updated
	r.Datetime = time.Now()
	r.Message = message
	r.OK = ok
	return r
}

func (r Response) Deleted(id interface{}, message string, ok bool) any {
	r.Id = id
	r.State = Deleted
	r.Datetime = time.Now()
	r.Message = message
	r.OK = ok
	return r
}
