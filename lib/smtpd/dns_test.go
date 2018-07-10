package smtpd

import (
	"testing"
)

func TestGetMx_1(t *testing.T) {
	d, err := DnsQuery("qq.com")
	if err == nil {
		t.Log("dns.Query ok:" + d)
	} else {
		t.Error("dns.Query fail:" + err.Error())
	}
}

func TestGetMx_2(t *testing.T) {
	d, err := DnsQuery("bb.com")
	if err == nil {
		t.Log("dns.Query ok:" + d)
	} else {
		t.Error("dns.Query fail:" + err.Error())
	}
}
