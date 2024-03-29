package crud

import "github.com/ewa-go/ewa"

type HeaderValue struct {
	Key, Value string
	IsDefault  bool
}

type HeaderValues []HeaderValue

func NewHeaderValues(h ...HeaderValue) HeaderValues {
	return h
}

func (h HeaderValues) Add(key, value string, isDefault ...bool) HeaderValues {
	var d bool
	if len(isDefault) > 0 {
		d = isDefault[0]
	}
	return append(h, HeaderValue{key, value, d})
}

func (h HeaderValues) Default() string {
	for _, hv := range h {
		if hv.IsDefault {
			return hv.Value
		}
	}
	return ""
}

func (r Route) GetValueFromHeader(c *ewa.Context, key string) string {
	header := c.Get(key)
	for _, hh := range r.Headers {
		if hh.Key == header {
			header = hh.Value
			break
		}
	}
	if header == "" {
		header = r.Headers.Default()
	}
	return header
}
