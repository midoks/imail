package smtpd

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
)

func DnsQuery(domain string) (string, error) {

	mx, _ := net.LookupMX(domain)
	mxLen := len(mx)

	if 0 == mxLen {
		return "", errors.New("not find mx record")
	}

	rand := rand.Int() % mxLen
	mxSelect := mx[rand]

	mxHost := fmt.Sprintf("%s", mxSelect.Host)
	mxHost = strings.Trim(mxHost, ".")

	return mxHost, nil
}
