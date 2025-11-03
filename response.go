package crud

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/ewa-go/ewa"
	"github.com/ewa-go/ewa/consts"
)

type Response struct {
	Ok       bool      `json:"ok"`
	State    string    `json:"state"`
	Datetime time.Time `json:"datetime"`
	Data     any       `json:"data,omitempty"`
}

func (rr Response) Send(c *ewa.Context, r *CRUD, state string, status int, data any) (err error) {
	var (
		body        any
		contentType string
		content     []byte
	)
	switch t := data.(type) {
	case error:
		rr.Ok = false
		body = t.Error()
	default:
		rr.Ok = true
		body = data
	}
	switch state {
	case Created, Updated, Deleted:
		rr.State = state
		rr.Datetime = time.Now()
		rr.Data = body
		contentType, content, _ = rr.accept(c.Get(consts.HeaderAccept), r)
	case Read:
		contentType, content, _ = rr.accept(c.Get(consts.HeaderAccept), data)
	}
	return c.Send(status, contentType, content)
}

func (rr Response) accept(header string, data any) (contentType string, content []byte, err error) {
	switch header {
	case consts.MIMEApplicationXML, consts.MIMEApplicationXMLCharsetUTF8:
		content, err = xml.Marshal(data)
		return consts.MIMEApplicationXMLCharsetUTF8, content, err
	case consts.MIMEApplicationJSON, consts.MIMEApplicationJSONCharsetUTF8:
		content, err = json.Marshal(data)
		return consts.MIMEApplicationXMLCharsetUTF8, content, err
	}
	return consts.MIMEApplicationXMLCharsetUTF8, []byte(fmt.Sprintf("%s", data)), nil
}
