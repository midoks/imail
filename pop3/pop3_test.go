package pop3

import (
	"fmt"
	// "strings"
	"testing"
)

func TestRunSendLocal(t *testing.T) {
	toEmail := "midoks@imail.com"
	fromEmail := "midoks@cachecha.com"
	content := fmt.Sprintf("Data: 24 May 2013 19:00:29\r\nFrom: <%s>\r\nSubject: Hello imail\r\nTo: <%s>\r\n\r\nHi! yes is test. liuxiaoming ok?!", fromEmail, toEmail)
	Delivery("127.0.0.1", "1025", fromEmail, toEmail, content)
}
