package smtpd

import (
	"fmt"
	// "strings"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	t.Log("hello world")
}

// func TestHelo_1(t *testing.T) {
// 	d, err := DnsQuery("qq.com")
// 	if err == nil {
// 		t.Log("dns.Query ok:" + d)
// 	} else {
// 		t.Error("dns.Query fail:" + err.Error())
// 	}
// }

// func TestRunSendFunc(t *testing.T) {
// 	toEmail := "midoks@163.com"
// 	fromEmail := "midoks@cachecha.com"
// 	toInfo := strings.Split(toEmail, "@")
// 	mxDomain, err := DnsQuery(toInfo[1])
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	fmt.Println(mxDomain)

// 	content := fmt.Sprintf("Data: 24 May 2013 19:00:29\r\nFrom: <%s>\r\nSubject: Hello imail\r\nTo: <%s>\r\n\r\nHi! yes is test. imail ok?!", fromEmail, toEmail)
// 	Delivery(mxDomain, "25", fromEmail, toEmail, content)
// }

func TestRunSendLocal(t *testing.T) {
	toEmail := "midoks@imail.com"
	fromEmail := "midoks@cachecha.com"
	content := fmt.Sprintf("Data: 24 May 2013 19:00:29\r\nFrom: <%s>\r\nSubject: Hello imail\r\nTo: <%s>\r\n\r\nHi! yes is test. imail ok?!", fromEmail, toEmail)
	Delivery("127.0.0.1", "1025", fromEmail, toEmail, content)
}

func Benchmark_SendLocal(b *testing.B) {
	toEmail := "midoks@imail.com"
	fromEmail := "midoks@cachecha.com"
	content := fmt.Sprintf("Data: 24 May 2013 19:00:29\r\nFrom: <%s>\r\nSubject: Hello imail\r\nTo: <%s>\r\n\r\nHi! yes is test. imail ok?!", fromEmail, toEmail)
	for i := 0; i < b.N; i++ { //use b.N for looping
		Delivery("127.0.0.1", "1025", fromEmail, toEmail, content)
	}
}
