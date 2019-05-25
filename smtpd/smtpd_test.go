package smtpd

import (
	"fmt"
	"strings"
	"testing"
)

func TestHelo_1(t *testing.T) {
	d, err := DnsQuery("qq.com")
	if err == nil {
		t.Log("dns.Query ok:" + d)
	} else {
		t.Error("dns.Query fail:" + err.Error())
	}
}

func TestRunSend(t *testing.T) {
	toEmail := "627293072@qq.com"
	fromEmail := "midoks@cachecha.com"
	toInfo := strings.Split(toEmail, "@")
	mxDomain, err := DnsQuery(toInfo[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(mxDomain)

	content := fmt.Sprintf("Data: 24 May 2013 19:00:29\nFrom: <%s>\nSubject: Hello imail\nTo: <%s>\n\nHi! yes is test. liuxiaoming ok?!", fromEmail, toEmail)
	SendMail(mxDomain, fromEmail, toEmail, content)
}

func TestRunSend2(t *testing.T) {
}
