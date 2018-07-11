package smtpd

import (
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
