package imap

import (
	"bufio"
	"bytes"
	"fmt"
	"net/textproto"
	"strings"
)

// A body structure.
// See RFC 3501 page 74.
type BodyStructure struct {
	// Basic fields

	// The MIME type.
	MimeType string
	// The MIME subtype.
	MimeSubType string
	// The MIME parameters.
	Params map[string]string

	// The Content-Id header.
	Id string
	// The Content-Description header.
	Description string
	// The Content-Encoding header.
	Encoding string
	// The Content-Length header.
	Size uint32

	// Type-specific fields

	// The children parts, if multipart.
	Parts []*BodyStructure
	// The envelope, if message/rfc822.
	Envelope *Envelope
	// The body structure, if message/rfc822.
	BodyStructure *BodyStructure
	// The number of lines, if text or message/rfc822.
	Lines uint32

	// Extension data

	// True if the body structure contains extension data.
	Extended bool

	// The Content-Disposition header field value.
	Disposition string
	// The Content-Disposition header field parameters.
	DispositionParams map[string]string
	// The Content-Language header field, if multipart.
	Language []string
	// The content URI, if multipart.
	Location []string

	// The MD5 checksum.
	MD5 string
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

// trim returns s with leading and trailing spaces and tabs removed.
// It does not assume Unicode or UTF-8.
func trim(s []byte) []byte {
	i := 0
	for i < len(s) && isSpace(s[i]) {
		i++
	}
	n := len(s)
	for n > i && isSpace(s[n-1]) {
		n--
	}
	return s[i:n]
}

func GetHeaderLine(r *bufio.Reader, line []byte) ([]byte, error) {
	for {
		l, more, err := r.ReadLine()

		if err != nil {
			break
		}

		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}

func GetHeader(body string) {
	bufferedBody := bufio.NewReader(strings.NewReader(body))

	for {
		kv, err := GetHeaderLine(bufferedBody, nil)
		if len(kv) == 0 {
			break
		}

		if err != nil {
			break
		}
		i := bytes.IndexByte(kv, ':')
		fmt.Println(":::", i, string(kv))
		if i < 0 {
			break
		}

		key := textproto.CanonicalMIMEHeaderKey(string(trim(kv[:i])))
		fmt.Println("key:", key)

	}

}
