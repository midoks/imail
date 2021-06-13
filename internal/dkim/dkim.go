package dkim

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
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

func CheckDomainA(domain string) error {
	findIp, err := net.LookupIP(domain)
	if err != nil {
		return err
	}

	mx, err := net.LookupMX(domain)
	if err != nil {
		return err
	}

	if len(mx) < 1 {
		return errors.New("not find domain mx!")
	}

	mxHost := fmt.Sprintf("%s", mx[0].Host)
	mxHost = strings.Trim(mxHost, ".")
	if !strings.HasSuffix(mxHost, domain) {
		return errors.New("It's not a top-level domain name!")
	}

	ip, err := GetPublicIP()
	if err != nil {
		return err
	}

	var isFind = false
	for _, fIp := range findIp {
		if strings.EqualFold(string(fIp), ip) {
			isFind = true
			break
		}
	}

	if !isFind {
		return errors.New("IP not configured by domain name!")
	}

	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func MakeDkimFile(domain string) error {
	Priv, Pub, err := makeRsa()

	if err != nil {
		return err
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: Priv,
	}

	// pri := b64.StdEncoding.EncodeToString(Priv)
	priFile := fmt.Sprintf("conf/dkim/%s/default.private", domain)
	file, err := os.Create(priFile)
	if err != nil {
		return err
	}

	err = pem.Encode(file, block)
	if err != nil {
		return err
	}

	pub := b64.StdEncoding.EncodeToString(Pub)

	pubContent := fmt.Sprintf("default._domainkey\tIN\tTXT\t( \"v=DKIM1; k=rsa; \"\r\n\t\"p=%s\" )  ; ----- DKIM key default for %s", pub, domain)

	pubFile := fmt.Sprintf("conf/dkim/%s/default.txt", domain)
	file, err = os.Create(pubFile)
	if err != nil {
		return err
	}

	_, err = io.WriteString(file, pubContent)
	if err != nil {
		return err
	}

	return nil
}

func MakeDkimConfFile(domain string) error {
	// if err := CheckDomainA(domain); err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	path := fmt.Sprintf("conf/dkim/%s", domain)
	if _, err := PathExists(path); err == nil {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	return MakeDkimFile(domain)
}
