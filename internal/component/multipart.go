package component

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	// "io/ioutil"
)

// This constant needs to be at least 76 for this package to work correctly.
// This is because \r\n--separator_of_len_70- would fill the buffer and it
// wouldn't be safe to consume a single byte from it.
const peekBufferSize = 4096

// MultipartReader is an iterator over parts in a MIME multipart body.
// MultipartReader's underlying parser consumes its input as needed. Seeking
// isn't supported.
type MultipartReader struct {
	bufReader *bufio.Reader

	currentPart *Part
	partsRead   int

	nl               []byte // "\r\n" or "\n" (set after seeing first boundary line)
	nlDashBoundary   []byte // nl + "--boundary"
	dashBoundaryDash []byte // "--boundary--"
	dashBoundary     []byte // "--boundary"
}

// stickyErrorReader is an io.Reader which never calls Read on its
// underlying Reader once an error has been seen. (the io.Reader
// interface's contract promises nothing about the return values of
// Read calls after an error, yet this package does do multiple Reads
// after error)
type stickyErrorReader struct {
	r   io.Reader
	err error
}

func (r *stickyErrorReader) Read(p []byte) (n int, _ error) {
	if r.err != nil {
		return 0, r.err
	}
	n, r.err = r.r.Read(p)
	return n, r.err
}

// NewMultipartReader creates a new multipart reader reading from r using the
// given MIME boundary.
//
// The boundary is usually obtained from the "boundary" parameter of
// the message's "Content-Type" header. Use mime.ParseMediaType to
// parse such headers.
func NewMultipartReader(r io.Reader, boundary string) *MultipartReader {
	b := []byte("\r\n--" + boundary + "--")
	return &MultipartReader{
		bufReader:        bufio.NewReaderSize(&stickyErrorReader{r: r}, peekBufferSize),
		nl:               b[:2],
		nlDashBoundary:   b[:len(b)-2],
		dashBoundaryDash: b[2:],
		dashBoundary:     b[2 : len(b)-2],
	}
}

func newPart(mr *MultipartReader) (*Part, error) {
	bp := &Part{mr: mr}
	if err := bp.populateHeaders(); err != nil {
		return nil, err
	}
	bp.r = partReader{bp}
	return bp, nil
}

// NextPart returns the next part in the multipart or an error.
// When there are no more parts, the error io.EOF is returned.
func (r *MultipartReader) NextPart() (*Part, error) {
	if r.currentPart != nil {
		r.currentPart.Close()
	}
	if string(r.dashBoundary) == "--" {
		return nil, fmt.Errorf("multipart: boundary is empty")
	}
	expectNewPart := false

	// line, err := r.bufReader.ReadSlice('\n')
	// fmt.Println("pp:", string(line), err)
	// cc, _ := ioutil.ReadAll(r.bufReader)
	// fmt.Println("cc:ccc", string(cc))
	for {
		line, err := r.bufReader.ReadSlice('\n')

		if err == io.EOF && r.isFinalBoundary(line) {
			// If the buffer ends in "--boundary--" without the
			// trailing "\r\n", ReadSlice will return an error
			// (since it's missing the '\n'), but this is a valid
			// multipart EOF so we need to return io.EOF instead of
			// a fmt-wrapped one.
			return nil, io.EOF
		}
		if err != nil {
			return nil, fmt.Errorf("multipart: NextPart: %v", err)
		}

		if r.isBoundaryDelimiterLine(line) {
			r.partsRead++
			bp, err := newPart(r)
			if err != nil {
				return nil, err
			}
			r.currentPart = bp
			return bp, nil
		}

		if r.isFinalBoundary(line) {
			// Expected EOF
			return nil, io.EOF
		}

		if expectNewPart {
			return nil, fmt.Errorf("multipart: expecting a new Part; got line %q", string(line))
		}

		if r.partsRead == 0 {
			// skip line
			continue
		}

		// Consume the "\n" or "\r\n" separator between the
		// body of the previous part and the boundary line we
		// now expect will follow. (either a new part or the
		// end boundary)
		if bytes.Equal(line, r.nl) {
			expectNewPart = true
			continue
		}

		return nil, fmt.Errorf("multipart: unexpected line in Next(): %q", line)
	}
}

// isFinalBoundary reports whether line is the final boundary line
// indicating that all parts are over.
// It matches `^--boundary--[ \t]*(\r\n)?$`
func (mr *MultipartReader) isFinalBoundary(line []byte) bool {
	if !bytes.HasPrefix(line, mr.dashBoundaryDash) {
		return false
	}
	rest := line[len(mr.dashBoundaryDash):]
	rest = skipLWSPChar(rest)
	return len(rest) == 0 || bytes.Equal(rest, mr.nl)
}

func (mr *MultipartReader) isBoundaryDelimiterLine(line []byte) (ret bool) {
	// https://tools.ietf.org/html/rfc2046#section-5.1
	//   The boundary delimiter line is then defined as a line
	//   consisting entirely of two hyphen characters ("-",
	//   decimal value 45) followed by the boundary parameter
	//   value from the Content-Type header field, optional linear
	//   whitespace, and a terminating CRLF.
	if !bytes.HasPrefix(line, mr.dashBoundary) {
		return false
	}
	rest := line[len(mr.dashBoundary):]
	rest = skipLWSPChar(rest)

	// On the first part, see our lines are ending in \n instead of \r\n
	// and switch into that mode if so. This is a violation of the spec,
	// but occurs in practice.
	if mr.partsRead == 0 && len(rest) == 1 && rest[0] == '\n' {
		mr.nl = mr.nl[1:]
		mr.nlDashBoundary = mr.nlDashBoundary[1:]
	}
	return bytes.Equal(rest, mr.nl)
}

// skipLWSPChar returns b with leading spaces and tabs removed.
// RFC 822 defines:
//    LWSP-char = SPACE / HTAB
func skipLWSPChar(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}

// partReader implements io.Reader by reading raw bytes directly from the
// wrapped *Part, without doing any Transfer-Encoding decoding.
type partReader struct {
	p *Part
}

func (pr partReader) Read(d []byte) (int, error) {
	p := pr.p
	br := p.mr.bufReader

	// Read into buffer until we identify some data to return,
	// or we find a reason to stop (boundary or read error).
	for p.n == 0 && p.err == nil {
		peek, _ := br.Peek(br.Buffered())
		p.n, p.err = scanUntilBoundary(peek, p.mr.dashBoundary, p.mr.nlDashBoundary, p.total, p.readErr)
		if p.n == 0 && p.err == nil {
			// Force buffered I/O to read more into buffer.
			_, p.readErr = br.Peek(len(peek) + 1)
			if p.readErr == io.EOF {
				p.readErr = io.ErrUnexpectedEOF
			}
		}
	}

	// Read out from "data to return" part of buffer.
	if p.n == 0 {
		return 0, p.err
	}
	n := len(d)
	if n > p.n {
		n = p.n
	}
	n, _ = br.Read(d[:n])
	p.total += int64(n)
	p.n -= n
	if p.n == 0 {
		return n, p.err
	}
	return n, nil
}

// scanUntilBoundary scans buf to identify how much of it can be safely
// returned as part of the Part body.
// dashBoundary is "--boundary".
// nlDashBoundary is "\r\n--boundary" or "\n--boundary", depending on what mode we are in.
// The comments below (and the name) assume "\n--boundary", but either is accepted.
// total is the number of bytes read out so far. If total == 0, then a leading "--boundary" is recognized.
// readErr is the read error, if any, that followed reading the bytes in buf.
// scanUntilBoundary returns the number of data bytes from buf that can be
// returned as part of the Part body and also the error to return (if any)
// once those data bytes are done.
func scanUntilBoundary(buf, dashBoundary, nlDashBoundary []byte, total int64, readErr error) (int, error) {
	if total == 0 {
		// At beginning of body, allow dashBoundary.
		if bytes.HasPrefix(buf, dashBoundary) {
			switch matchAfterPrefix(buf, dashBoundary, readErr) {
			case -1:
				return len(dashBoundary), nil
			case 0:
				return 0, nil
			case +1:
				return 0, io.EOF
			}
		}
		if bytes.HasPrefix(dashBoundary, buf) {
			return 0, readErr
		}
	}

	// Search for "\n--boundary".
	if i := bytes.Index(buf, nlDashBoundary); i >= 0 {
		switch matchAfterPrefix(buf[i:], nlDashBoundary, readErr) {
		case -1:
			return i + len(nlDashBoundary), nil
		case 0:
			return i, nil
		case +1:
			return i, io.EOF
		}
	}
	if bytes.HasPrefix(nlDashBoundary, buf) {
		return 0, readErr
	}

	// Otherwise, anything up to the final \n is not part of the boundary
	// and so must be part of the body.
	// Also if the section from the final \n onward is not a prefix of the boundary,
	// it too must be part of the body.
	i := bytes.LastIndexByte(buf, nlDashBoundary[0])
	if i >= 0 && bytes.HasPrefix(nlDashBoundary, buf[i:]) {
		return i, nil
	}
	return len(buf), readErr
}

// matchAfterPrefix checks whether buf should be considered to match the boundary.
// The prefix is "--boundary" or "\r\n--boundary" or "\n--boundary",
// and the caller has verified already that bytes.HasPrefix(buf, prefix) is true.
//
// matchAfterPrefix returns +1 if the buffer does match the boundary,
// meaning the prefix is followed by a dash, space, tab, cr, nl, or end of input.
// It returns -1 if the buffer definitely does NOT match the boundary,
// meaning the prefix is followed by some other character.
// For example, "--foobar" does not match "--foo".
// It returns 0 more input needs to be read to make the decision,
// meaning that len(buf) == len(prefix) and readErr == nil.
func matchAfterPrefix(buf, prefix []byte, readErr error) int {
	if len(buf) == len(prefix) {
		if readErr != nil {
			return +1
		}
		return 0
	}
	c := buf[len(prefix)]
	if c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '-' {
		return +1
	}
	return -1
}
