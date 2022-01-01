package component

import (
	// "bufio"
	// "bytes"
	// "errors"
	// "fmt"
	"io"
	"io/ioutil"
	// "mime"
	// "strings"
	// "net/textproto"
)

// A Part represents a single part in a multipart body.
type Part struct {
	Header Header

	mr *MultipartReader

	disposition       string
	dispositionParams map[string]string
	Content           string

	// r is either a reader directly reading from mr
	r io.Reader

	n       int   // known data bytes waiting in mr.bufReader
	total   int64 // total data bytes read already
	err     error // error to return when n == 0
	readErr error // read error observed from mr.bufReader
}

// Read reads the body of a part, after its headers and before the
// next part (if any) begins.
func (p *Part) Read(d []byte) (n int, err error) {
	return p.r.Read(d)
}

func (bp *Part) populateHeaders() error {
	header, err := ReadHeader(bp.mr.bufReader)
	if err == nil {
		bp.Header = header
	}
	return err
}

func (p *Part) Close() error {
	io.Copy(ioutil.Discard, p)
	return nil
}
