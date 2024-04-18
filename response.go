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

func (r Response) Read(id any, data any, err error) any {
	r.Id = id
	r.State = Read
	r.Datetime = time.Now()
	r.Data = data
	if err != nil {
		r.Message = err.Error()
		r.OK = true
	}
	return r
}

func (r Response) Created(id interface{}, err error) any {
	r.Id = id
	r.State = Created
	r.Datetime = time.Now()
	if err != nil {
		r.Message = err.Error()
		r.OK = true
	}
	return r
}

func (r Response) Updated(id interface{}, err error) any {
	r.Id = id
	r.State = Updated
	r.Datetime = time.Now()
	if err != nil {
		r.Message = err.Error()
		r.OK = true
	}
	return r
}

func (r Response) Deleted(id interface{}, err error) any {
	r.Id = id
	r.State = Deleted
	r.Datetime = time.Now()
	if err != nil {
		r.Message = err.Error()
		r.OK = true
	}
	return r
}
