package component

import (
	// "bufio"
	// "fmt"
	// "strings"
	"time"
)

// A message envelope, ie. message metadata from its headers.
// See RFC 3501 page 77.
type Envelope struct {
	// The message date.
	Date time.Time
	// The message subject.
	Subject string
	// The From header addresses.
	From []*Address
	// The message senders.
	Sender []*Address
	// The Reply-To header addresses.
	ReplyTo []*Address
	// The To header addresses.
	To []*Address
	// The Cc header addresses.
	Cc []*Address
	// The Bcc header addresses.
	Bcc []*Address
	// The In-Reply-To header. Contains the parent Message-Id.
	InReplyTo string
	// The Message-Id header.
	MessageId string
}

// Parse an address list from fields.
func ParseAddressList(fields []interface{}) (addrs []*Address) {
	addrs = make([]*Address, len(fields))

	for i, f := range fields {
		if addrFields, ok := f.([]interface{}); ok {
			addr := &Address{}
			if err := addr.Parse(addrFields); err == nil {
				addrs[i] = addr
			}
		}
	}

	return
}

// Format an address list to fields.
func FormatAddressList(addrs []*Address) (fields []interface{}) {
	fields = make([]interface{}, len(addrs))

	for i, addr := range addrs {
		fields[i] = addr.Format()
	}

	return
}

type (
	Date             time.Time
	DateTime         time.Time
	envelopeDateTime time.Time
	searchDate       time.Time
)

// Format an envelope to fields.
func (e *Envelope) Format() (fields []interface{}) {
	return []interface{}{
		envelopeDateTime(e.Date),
		encodeHeader(e.Subject),
		FormatAddressList(e.From),
		FormatAddressList(e.Sender),
		FormatAddressList(e.ReplyTo),
		FormatAddressList(e.To),
		FormatAddressList(e.Cc),
		FormatAddressList(e.Bcc),
		e.InReplyTo,
		e.MessageId,
	}
}
