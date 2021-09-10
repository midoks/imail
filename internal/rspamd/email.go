package rspamd

import (
	"fmt"
	"io"
)

// Email represents an abstract type holding the email content and headers to pass to rspamd.
type Email struct {
	message io.Reader
	queueID string
	options Options
}

// Options encapsulate headers the client can pass in requests to rspamd.
type Options struct {
	flag   int
	weight float64
}

// SymbolData encapsulates the data returned for each symbol from Check.
type SymbolData struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	MetricScore float64 `json:"metric_score"`
	Description string  `json:"description"`
}

// NewEmailFromWriterTo creates an Email instance from an io.WriterTo.
func NewEmailFromWriterTo(message io.WriterTo) *Email {
	return &Email{
		message: readerFromWriterTo(message),
		options: Options{},
	}
}

// NewEmailFromWriterTo creates an Email instance from an io.Reader.
func NewEmailFromReader(message io.Reader) *Email {
	return &Email{
		message: message,
		options: Options{},
	}
}

// QueueID attaches a queue-id to an Email, and eventually as a header when sent to rspamd.
// This header helps clients with rspamd logging.
func (e *Email) QueueID(queueID string) *Email {
	e.queueID = queueID
	return e
}

// Flag attaches a flag to an Email, and eventually as a header when sent to rspamd.
// Flag identifies fuzzy storage.
func (e *Email) Flag(flag int) *Email {
	e.options.flag = flag
	return e
}

// Weight attaches a weight to an Email, and eventually as a header when sent to rspamd.
// Weight is added to hashes.
func (e *Email) Weight(weight float64) *Email {
	e.options.weight = weight
	return e
}

func readerFromWriterTo(writerTo io.WriterTo) io.Reader {
	r, w := io.Pipe()

	go func() {
		if _, err := writerTo.WriteTo(w); err != nil {
			_ = w.CloseWithError(fmt.Errorf("writing to pipe: %q", err))
			return
		}

		_ = w.Close() // Always succeeds
	}()

	return r
}
