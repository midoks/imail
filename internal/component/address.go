package component

import (
	"errors"
	"io"
	// "bufio"
	// "fmt"
	// "strings"
)

// An address.
type Address struct {
	// The personal name.
	PersonalName string
	// The SMTP at-domain-list (source route).
	AtDomainList string
	// The mailbox name.
	MailboxName string
	// The host name.
	HostName string
}

type parseError struct {
	error
}

// A literal, as defined in RFC 3501 section 4.3.
type Literal interface {
	io.Reader

	// Len returns the number of bytes of the literal.
	Len() int
}

type (
	// A raw string.
	RawString string
)

func newParseError(text string) error {
	return &parseError{errors.New(text)}
}

// ParseString parses a string, which is either a literal, a quoted string or an
// atom.
func ParseString(f interface{}) (string, error) {
	if s, ok := f.(string); ok {
		return s, nil
	}

	// Useful for tests
	if a, ok := f.(RawString); ok {
		return string(a), nil
	}

	if l, ok := f.(Literal); ok {
		b := make([]byte, l.Len())
		if _, err := io.ReadFull(l, b); err != nil {
			return "", err
		}
		return string(b), nil
	}

	return "", newParseError("expected a string")
}

// Parse an address from fields.
func (addr *Address) Parse(fields []interface{}) error {
	if len(fields) < 4 {
		return errors.New("Address doesn't contain 4 fields")
	}

	if s, err := ParseString(fields[0]); err == nil {
		addr.PersonalName, _ = decodeHeader(s)
	}
	if s, err := ParseString(fields[1]); err == nil {
		addr.AtDomainList, _ = decodeHeader(s)
	}
	if s, err := ParseString(fields[2]); err == nil {
		addr.MailboxName, _ = decodeHeader(s)
	}
	if s, err := ParseString(fields[3]); err == nil {
		addr.HostName, _ = decodeHeader(s)
	}

	return nil
}

// Format an address to fields.
func (addr *Address) Format() []interface{} {
	fields := make([]interface{}, 4)

	if addr.PersonalName != "" {
		fields[0] = encodeHeader(addr.PersonalName)
	}
	if addr.AtDomainList != "" {
		fields[1] = addr.AtDomainList
	}
	if addr.MailboxName != "" {
		fields[2] = addr.MailboxName
	}
	if addr.HostName != "" {
		fields[3] = addr.HostName
	}

	return fields
}
