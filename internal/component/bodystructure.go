package component

import (
	// "bufio"
	// "bytes"
	"fmt"
	"io"
	"mime"
	// "net/textproto"
	// "io/ioutil"
	"reflect"
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
	Size int

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

func (bs *BodyStructure) Format() (fields []interface{}) {
	if strings.EqualFold(bs.MimeType, "multipart") {
		for _, part := range bs.Parts {
			fields = append(fields, part.Format())
		}

		fields = append(fields, bs.MimeSubType)

		if bs.Extended {
			extended := make([]interface{}, 3)

			if bs.Params != nil {
				extended[0] = formatHeaderParamList(bs.Params)
			}
			if bs.Disposition != "" {
				extended[1] = []interface{}{
					encodeHeader(bs.Disposition),
					formatHeaderParamList(bs.DispositionParams),
				}
			}
			if bs.Language != nil {
				extended[2] = FormatStringList(bs.Language)
			}
			if bs.Location != nil {
				// extended[3] = FormatStringList(bs.Location)
			}

			fields = append(fields, extended...)
		}
	} else {
		fields = make([]interface{}, 7)
		fields[0] = bs.MimeType
		fields[1] = bs.MimeSubType
		fields[2] = formatHeaderParamList(bs.Params)

		if bs.Id != "" {
			fields[3] = bs.Id
		}
		if bs.Description != "" {
			fields[4] = encodeHeader(bs.Description)
		}
		if bs.Encoding != "" {
			fields[5] = bs.Encoding
		}

		fields[6] = bs.Size

		// Type-specific fields
		if strings.EqualFold(bs.MimeType, "message") && strings.EqualFold(bs.MimeSubType, "rfc822") {
			var env interface{}
			if bs.Envelope != nil {
				env = bs.Envelope.Format()
			}

			var bsbs interface{}
			if bs.BodyStructure != nil {
				bsbs = bs.BodyStructure.Format()
			}

			fields = append(fields, env, bsbs, bs.Lines)
		}
		if strings.EqualFold(bs.MimeType, "text") {
			fields = append(fields, bs.Lines)
		}

		// Extension data
		if bs.Extended {
			extended := make([]interface{}, 3)

			if bs.MD5 != "" {
				extended[0] = bs.MD5
			}
			if bs.Disposition != "" {
				extended[1] = []interface{}{
					encodeHeader(bs.Disposition),
					formatHeaderParamList(bs.DispositionParams),
				}
			}
			if bs.Language != nil {
				extended[2] = FormatStringList(bs.Language)
			}
			if bs.Location != nil {
				// extended[3] = FormatStringList(bs.Location)
			}

			fields = append(fields, extended...)
		}
	}
	return
}

func mergeFields(fields []interface{}) string {
	result := "("
	for i := 0; i < len(fields); i++ {

		switch fields[i].(type) {
		case nil:
			result += "NIL "
		case string:
			result += fmt.Sprintf("\"%s\" ", fields[i])
		case uint32:
			result += fmt.Sprintf("%d ", fields[i])
		case int:
			result += fmt.Sprintf("%d ", fields[i])
		case interface{}:
			result += fmt.Sprintf("%s ", mergeFields(fields[i].([]interface{})))
		default:
			fmt.Println("type::", reflect.TypeOf(fields[i]))
		}
	}
	result = fmt.Sprintf("%s)", result)
	return result
}

func (bs *BodyStructure) ToString() string {
	t := bs.Format()
	res := mergeFields(t)
	return res
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

// Convert a string list to a field list.
func FormatStringList(list []string) (fields []interface{}) {
	fields = make([]interface{}, len(list))
	for i, v := range list {
		fields[i] = v
	}
	return
}

// CharsetReader, if non-nil, defines a function to generate charset-conversion
// readers, converting from the provided charset into UTF-8. Charsets are always
// lower-case. utf-8 and us-ascii charsets are handled by default. One of the
// the CharsetReader's result values must be non-nil.
//
// Importing github.com/emersion/go-message/charset will set CharsetReader to
// a function that handles most common charsets. Alternatively, CharsetReader
// can be set to e.g. golang.org/x/net/html/charset.NewReaderLabel.
var CharsetReader func(charset string, input io.Reader) (io.Reader, error)

// charsetReader calls CharsetReader if non-nil.
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	charset = strings.ToLower(charset)
	// "ascii" is not in the spec but is common
	if charset == "utf-8" || charset == "us-ascii" || charset == "ascii" {
		return input, nil
	}
	if CharsetReader != nil {
		return CharsetReader(charset, input)
	}
	return input, fmt.Errorf("message: unhandled charset %q", charset)
}

// decodeHeader decodes an internationalized header field. If it fails, it
// returns the input string and the error.
func decodeHeader(s string) (string, error) {
	wordDecoder := mime.WordDecoder{CharsetReader: charsetReader}
	dec, err := wordDecoder.DecodeHeader(s)
	if err != nil {
		return s, err
	}
	return dec, nil
}

func encodeHeader(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}

func FormatParamList(params map[string]string) []interface{} {
	var fields []interface{}
	for key, value := range params {
		fields = append(fields, key, value)
	}

	return fields
}

func formatHeaderParamList(params map[string]string) []interface{} {
	encoded := make(map[string]string)
	for k, v := range params {
		encoded[k] = encodeHeader(v)
	}
	return FormatParamList(encoded)
}

// FetchBodyStructure computes a message's body structure from its content.
func FetchBodyStructure(header Header, body io.Reader, extended bool) (*BodyStructure, error) {
	bs := new(BodyStructure)

	mediaType, mediaParams, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err == nil {
		typeParts := strings.SplitN(mediaType, "/", 2)
		bs.MimeType = typeParts[0]
		if len(typeParts) == 2 {
			bs.MimeSubType = typeParts[1]
		}
		bs.Params = mediaParams
	} else {
		bs.MimeType = "text"
		bs.MimeSubType = "plain"
	}

	bs.Id = header.Get("Content-Id")
	bs.Description = header.Get("Content-Description")
	bs.Encoding = header.Get("Content-Transfer-Encoding")

	// multipartReader(header, body)
	if mr := multipartReader(header, body); mr != nil {
		var parts []*BodyStructure
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			pbs, err := FetchBodyStructure(p.Header, p, extended)
			if err != nil {
				return nil, err
			}
			parts = append(parts, pbs)
		}
		bs.Parts = parts
	}

	// DO: bs.Size
	// bodyStr, err := ioutil.ReadAll(body)
	// if err != nil {
	// 	return bs, err
	// }
	// bs.Size = len(string(bodyStr))

	// TODO: bs.Envelope, bs.BodyStructure
	// TODO: bs.Lines

	if extended {
		bs.Extended = true

		bs.Disposition, bs.DispositionParams, _ = mime.ParseMediaType(header.Get("Content-Disposition"))

		// TODO: bs.Language, bs.Location
		// TODO: bs.MD5
		bs.MD5 = ""
	}
	return bs, nil
}
