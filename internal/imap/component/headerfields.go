package component

import (
	"net/textproto"
)

type HeaderFields struct {
	b []byte // Raw header field, including whitespace
	k string
	v string
}

func newHeaderFields(k, v string, b []byte) *HeaderField {
	return &HeaderField{k: textproto.CanonicalMIMEHeaderKey(k), v: v, b: b}
}
