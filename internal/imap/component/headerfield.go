package component

import (
	"net/textproto"
)

type HeaderField struct {
	b []byte // Raw header field, including whitespace
	k string
	v string
}

func newHeaderField(k, v string, b []byte) *HeaderField {
	return &HeaderField{k: textproto.CanonicalMIMEHeaderKey(k), v: v, b: b}
}
