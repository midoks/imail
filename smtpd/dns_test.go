package smtpd

import (
	"fmt"
	"testing"
)

func TestGetMx_1(t *testing.T) {
	d, err := DnsQuery("163.com")
	fmt.Println(d, err)
	if err == nil {
		t.Log("dns.Query ok:" + d)
	} else {
		t.Error("dns.Query fail:" + err.Error())
	}
}
