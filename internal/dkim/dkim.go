package dkim

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	// "strings"
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

func CheckDomainA(domain string) {
	aIp, _ := net.LookupIP(domain)
	fmt.Println("aIP:", aIp)

	mx, _ := net.LookupMX(domain)
	fmt.Println("amx", mx)

	ip, err := GetPublicIP()

	fmt.Println(ip, err)
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
