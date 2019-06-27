package imap

import (
// "bufio"
// "fmt"
// "strings"
// "time"
)

type headerField struct {
	b []byte // Raw header field, including whitespace
	k string
	v string
}

// A Header represents the key-value pairs in a message header.
//
// The header representation is idempotent: if the header can be read and
// written, the result will be exactly the same as the original (including
// whitespace). This is required for e.g. DKIM.
//
// Mutating the header is restricted: the only two allowed operations are
// inserting a new header field at the top and deleting a header field. This is
// again necessary for DKIM.
type Header struct {
	// Fields are in reverse order so that inserting a new field at the top is
	// cheap.
	l []*headerField
	m map[string][]*headerField
}
