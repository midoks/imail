package component

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
	"time"
)

var testDate, _ = time.Parse(time.RFC1123Z, "Sat, 18 Jun 2016 12:00:00 +0900")

const testHeaderString = "Content-Type: multipart/mixed; boundary=message-boundary\r\n" +
	"Date: Sat, 18 Jun 2016 12:00:00 +0900\r\n" +
	"Date: Sat, 19 Jun 2016 12:00:00 +0900\r\n" +
	"From: Mitsuha Miyamizu <mitsuha.miyamizu@example.org>\r\n" +
	"Reply-To: Mitsuha Miyamizu <mitsuha.miyamizu+replyto@example.org>\r\n" +
	"Message-Id: 42@example.org\r\n" +
	"Subject: Your Name.\r\n" +
	"To: Taki Tachibana <taki.tachibana@example.org>\r\n" +
	"\r\n"

const testHeaderFromToString = "From: Mitsuha Miyamizu <mitsuha.miyamizu@example.org>\r\n" +
	"To: Taki Tachibana <taki.tachibana@example.org>\r\n" +
	"\r\n"

const testHeaderDateString = "Date: Sat, 18 Jun 2016 12:00:00 +0900\r\n" +
	"Date: Sat, 19 Jun 2016 12:00:00 +0900\r\n" +
	"\r\n"

const testHeaderNoFromToString = "Content-Type: multipart/mixed; boundary=message-boundary\r\n" +
	"Date: Sat, 18 Jun 2016 12:00:00 +0900\r\n" +
	"Date: Sat, 19 Jun 2016 12:00:00 +0900\r\n" +
	"Reply-To: Mitsuha Miyamizu <mitsuha.miyamizu+replyto@example.org>\r\n" +
	"Message-Id: 42@example.org\r\n" +
	"Subject: Your Name.\r\n" +
	"\r\n"

const testAltHeaderString = "Content-Type: multipart/alternative; boundary=b2\r\n" +
	"\r\n"

const testTextHeaderString = "Content-Disposition: inline\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n"

const testTextContentTypeString = "Content-Type: text/plain\r\n" +
	"\r\n"

const testTextNoContentTypeString = "Content-Disposition: inline\r\n" +
	"\r\n"

const testTextBodyString = "What's your name?"

const testTextString = testTextHeaderString + testTextBodyString

const testHTMLHeaderString = "Content-Disposition: inline\r\n" +
	"Content-Type: text/html\r\n" +
	"\r\n"

const testHTMLBodyString = "<div>What's <i>your</i> name?</div>"

const testHTMLString = testHTMLHeaderString + testHTMLBodyString

const testAttachmentHeaderString = "Content-Disposition: attachment; filename=note.txt\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n"

const testAttachmentBodyString = "My name is Mitsuha."

const testAttachmentString = testAttachmentHeaderString + testAttachmentBodyString

const testBodyString = "--message-boundary\r\n" +
	testAltHeaderString +
	"\r\n--b2\r\n" +
	testTextString +
	"\r\n--b2\r\n" +
	testHTMLString +
	"\r\n--b2--\r\n" +
	"\r\n--message-boundary\r\n" +
	testAttachmentString +
	"\r\n--message-boundary--\r\n"

const testMailString = testHeaderString + testBodyString

var testBodyStructure = &BodyStructure{
	MimeType:    "multipart",
	MimeSubType: "mixed",
	Params:      map[string]string{"boundary": "message-boundary"},
	Parts: []*BodyStructure{
		{
			MimeType:    "multipart",
			MimeSubType: "alternative",
			Params:      map[string]string{"boundary": "b2"},
			Extended:    true,
			Parts: []*BodyStructure{
				{
					MimeType:          "text",
					MimeSubType:       "plain",
					Params:            map[string]string{},
					Extended:          true,
					Disposition:       "inline",
					DispositionParams: map[string]string{},
				},
				{
					MimeType:          "text",
					MimeSubType:       "html",
					Params:            map[string]string{},
					Extended:          true,
					Disposition:       "inline",
					DispositionParams: map[string]string{},
				},
			},
		},
		{
			MimeType:          "text",
			MimeSubType:       "plain",
			Params:            map[string]string{},
			Extended:          true,
			Disposition:       "attachment",
			DispositionParams: map[string]string{"filename": "note.txt"},
		},
	},
	Extended: true,
}

// go test -v ./internal/component
func TestFetchBodyStructure(t *testing.T) {

	bufferedBody := bufio.NewReader(strings.NewReader(testMailString))
	header, err := ReadHeader(bufferedBody)
	if err != nil {
		t.Fatal("Expected no error while reading mail, got:", err)
	}

	bs, err := FetchBodyStructure(header, bufferedBody, true)
	if err != nil {
		t.Fatal("Expected no error while fetching body structure, got:", err)
	}

	if !reflect.DeepEqual(testBodyStructure, bs) {
		t.Errorf("Expected body structure \n%+v\n but got \n%+v", testBodyStructure, bs)
	}
}
