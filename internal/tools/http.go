package tools

import (
	"io/ioutil"
	"net/http"
)

func GetPublicIP() (ip string, err error) {
	// - http://myexternalip.com/raw
	// - http://ip.dhcp.cn/?ip
	resp, err := http.Get("http://ip.dhcp.cn/?ip")
	content, err := ioutil.ReadAll(resp.Body)

	if err == nil {
		return string(content), nil
	}
	return "", err
}
