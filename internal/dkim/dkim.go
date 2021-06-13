package dkim

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

func makeRsa() ([]byte, []byte, error) {
	privatekey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err == nil {
		var publickey *rsa.PublicKey
		publickey = &privatekey.PublicKey
		Priv := x509.MarshalPKCS1PrivateKey(privatekey)
		Pub, err := x509.MarshalPKIXPublicKey(publickey)

		if err == nil {
			return Priv, Pub, nil
		}
	}
	return []byte{}, []byte{}, err
}

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

func CheckDomainA(domain string) (bool, error) {
	findIp, err := net.LookupIP(domain)
	if err != nil {
		return false, err
	}

	mx, err := net.LookupMX(domain)
	if err != nil {
		return false, err
	}

	if len(mx) < 1 {
		return false, errors.New("not find domain mx!")
	}

	mxHost := fmt.Sprintf("%s", mx[0].Host)
	mxHost = strings.Trim(mxHost, ".")
	if !strings.HasSuffix(mxHost, domain) {
		return false, errors.New("It's not a top-level domain name!")
	}

	ip, err := GetPublicIP()
	if err != nil {
		return false, err
	}

	var isFind = false
	for _, fIp := range findIp {
		if strings.EqualFold(string(fIp), ip) {
			isFind = true
			break
		}
	}

	if !isFind {
		return false, errors.New("IP not configured by domain name!")
	}

	return true, nil
}

func MakeDkimFile() {

}

func MakeDkimConfFile(domain string) {

}

func DKIM() (pri, pub string) {

	Priv, Pub, err := makeRsa()
	pub = ""
	pri = ""
	if err == nil {

		block := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: Priv,
		}

		pri = b64.StdEncoding.EncodeToString(Priv)

		file, err := os.Create("conf/dkim/private.pem")
		fmt.Println(err)

		err = pem.Encode(file, block)
		fmt.Println(err)

		pub = b64.StdEncoding.EncodeToString(Pub)

		blockPub := &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: Pub,
		}
		file, err = os.Create("conf/dkim/public.pem")
		fmt.Println(err)

		err = pem.Encode(file, blockPub)
		fmt.Println(err)

		fmt.Println(pub, pri)
	}
	return
}
