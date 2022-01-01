package component

import (
	// "bufio"
	// "bytes"
	"errors"
	"io"
	"mime"
	"strings"
	// "net/textproto"
)

var errNoSuchPart = errors.New("backendutil: no such message body part")

// A body section name.
// See RFC 3501 page 55.
type BodySectionName struct {
	BodyPartName

	// If set to true, do not implicitly set the \Seen flag.
	Peek bool
	// The substring of the section requested. The first value is the position of
	// the first desired octet and the second value is the maximum number of
	// octets desired.
	Partial []int

	value FetchItem
}

// A body part name.
type BodyPartName struct {
	// The specifier of the requested part.
	Specifier PartSpecifier
	// The part path. Parts indexes start at 1.
	Path []int
	// If Specifier is HEADER, contains header fields that will/won't be returned,
	// depending of the value of NotFields.
	Fields []string
	// If set to true, Fields is a blacklist of fields instead of a whitelist.
	NotFields bool
}

// A PartSpecifier specifies which parts of the MIME entity should be returned.
type PartSpecifier string

// Part specifiers described in RFC 3501 page 55.
const (
	// Refers to the entire part, including headers.
	EntireSpecifier PartSpecifier = ""
	// Refers to the header of the part. Must include the final CRLF delimiting
	// the header and the body.
	HeaderSpecifier = "HEADER"
	// Refers to the text body of the part, omitting the header.
	TextSpecifier = "TEXT"
	// Refers to the MIME Internet Message Body header.  Must include the final
	// CRLF delimiting the header and the body.
	MIMESpecifier = "MIME"
)

// A FetchItem is a message data item that can be fetched.
type FetchItem string

func multipartReader(header Header, body io.Reader) *MultipartReader {
	contentType := header.Get("Content-Type")
	// contentTransferEncoding := header.Get("Content-Transfer-Encoding")

	if !strings.HasPrefix(contentType, "multipart/") {
		return nil
	}

	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil
	}

	mpr := NewMultipartReader(body, params["boundary"])
	return mpr
}
