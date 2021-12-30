package component

import (
	"bufio"
	"fmt"
	"net/textproto"
	"strings"
	// "time"
	"bytes"
)

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
	l []*HeaderField
	m map[string][]*HeaderField
}

// Get gets the first value associated with the given key. If there are no
// values associated with the key, Get returns "".
func (h *Header) Get(k string) string {
	fields := h.m[textproto.CanonicalMIMEHeaderKey(k)]
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1].v
}

func MakeHeaderMap(fs []*HeaderField) map[string][]*HeaderField {
	if len(fs) == 0 {
		return nil
	}

	m := make(map[string][]*HeaderField)
	for i, f := range fs {
		m[f.k] = append(m[f.k], fs[i])
	}
	return m
}

func NewHeader(fs []*HeaderField) Header {
	// Reverse order
	for i := len(fs)/2 - 1; i >= 0; i-- {
		opp := len(fs) - 1 - i
		fs[i], fs[opp] = fs[opp], fs[i]
	}

	// Populate map
	m := MakeHeaderMap(fs)
	return Header{l: fs, m: m}
}

func readLineSlice(r *bufio.Reader, line []byte) ([]byte, error) {
	for {
		l, more, err := r.ReadLine()
		if err != nil {
			return nil, err
		}

		line = append(line, l...)
		if !more {
			break
		}
	}

	return line, nil
}

func hasContinuationLine(r *bufio.Reader) bool {
	c, err := r.ReadByte()
	if err != nil {
		return false // bufio will keep err until next read.
	}
	r.UnreadByte()
	return isSpace(c)
}

func readContinuedLineSlice(r *bufio.Reader) ([]byte, error) {
	// Read the first line.
	line, err := readLineSlice(r, nil)
	if err != nil {
		return nil, err
	}

	if len(line) == 0 { // blank line - no continuation
		return line, nil
	}

	line = append(line, '\r', '\n')

	// Read continuation lines.
	for hasContinuationLine(r) {
		line, err = readLineSlice(r, line)
		if err != nil {
			break // bufio will keep err until next read.
		}

		line = append(line, '\r', '\n')
	}

	return line, nil
}

// ReadHeader reads a MIME header from r. The header is a sequence of possibly
// continued Key: Value lines ending in a blank line.
func ReadHeader(r *bufio.Reader) (Header, error) {
	var fs []*HeaderField

	// The first line cannot start with a leading space.
	if buf, err := r.Peek(1); err == nil && isSpace(buf[0]) {
		line, err := readLineSlice(r, nil)
		if err != nil {
			return NewHeader(fs), err
		}

		return NewHeader(fs), fmt.Errorf("message: malformed MIME header initial line: %v", string(line))
	}

	for {
		kv, err := readContinuedLineSlice(r)
		if len(kv) == 0 {
			return NewHeader(fs), err
		}

		// Key ends at first colon; should not have trailing spaces but they
		// appear in the wild, violating specs, so we remove them if present.
		i := bytes.IndexByte(kv, ':')
		if i < 0 {
			return NewHeader(fs), fmt.Errorf("message: malformed MIME header line: %v", string(kv))
		}

		key := textproto.CanonicalMIMEHeaderKey(string(trim(kv[:i])))

		// As per RFC 7230 field-name is a token, tokens consist of one or more
		// chars. We could return a an error here, but better to be liberal in
		// what we accept, so if we get an empty key, skip it.
		if key == "" {
			continue
		}

		i++ // skip colon
		v := kv[i:]

		value := trimAroundNewlines(v)
		fs = append(fs, newHeaderField(key, value, kv))

		if err != nil {
			return NewHeader(fs), err
		}
	}
}

// ReadHeader reads a MIME header from r. The header is a sequence of possibly
// continued Key: Value lines ending in a blank line.
func ReadHeaderString(r *bufio.Reader) (string, error) {

	var lines []byte
	// The first line cannot start with a leading space.
	if buf, err := r.Peek(1); err == nil && isSpace(buf[0]) {
		line, err := readLineSlice(r, nil)
		if err != nil {
			return string(lines), err
		}
		return string(lines), fmt.Errorf("message: malformed MIME header initial line: %v", string(line))
	}

	for {
		kv, err := readContinuedLineSlice(r)
		if len(kv) == 0 {
			// lines = append(lines, '\r', '\n')
			return string(lines), err
		}

		// Key ends at first colon; should not have trailing spaces but they
		// appear in the wild, violating specs, so we remove them if present.
		i := bytes.IndexByte(kv, ':')
		if i < 0 {
			return string(lines), fmt.Errorf("message: malformed MIME header line: %v", string(kv))
			// return NewHeader(fs), fmt.Errorf("message: malformed MIME header line: %v", string(kv))
		}

		key := textproto.CanonicalMIMEHeaderKey(string(trim(kv[:i])))

		// As per RFC 7230 field-name is a token, tokens consist of one or more
		// chars. We could return a an error here, but better to be liberal in
		// what we accept, so if we get an empty key, skip it.
		if key == "" {
			continue
		}

		if err != nil {
			// lines = append(lines, '\r', '\n')
			return string(lines), err
		}

		lines = append(lines, kv...)
	}
}

// Strip newlines and spaces around newlines.
func trimAroundNewlines(v []byte) string {
	var b strings.Builder
	for {
		i := bytes.IndexByte(v, '\n')
		if i < 0 {
			writeContinued(&b, v)
			break
		}
		writeContinued(&b, v[:i])
		v = v[i+1:]
	}

	return b.String()
}

func writeContinued(b *strings.Builder, l []byte) {
	// Strip trailing \r, if any
	if len(l) > 0 && l[len(l)-1] == '\r' {
		l = l[:len(l)-1]
	}
	l = trim(l)
	if len(l) == 0 {
		return
	}
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.Write(l)
}
