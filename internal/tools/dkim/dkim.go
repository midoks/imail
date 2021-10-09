package dkim

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/midoks/imail/internal/tools"
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

func CheckDomainA(domain string) error {
	findIp, err := net.LookupIP(domain)
	if err != nil {
		return err
	}

	ip, err := tools.GetPublicIP()
	if err != nil {
		return err
	}

	var isFind = false
	for _, fIp := range findIp {
		if strings.EqualFold(fIp.String(), ip) {
			isFind = true
			break
		}
	}

	if !isFind {
		return errors.New("IP not configured by domain name!")
	}

	return nil
}

func MakeDkimFile(path, domain string) (string, error) {
	priFile := fmt.Sprintf("%s/dkim/%s/default.private", path, domain)
	defalutTextFile := fmt.Sprintf("%s/dkim/%s/default.txt", path, domain)
	defalutValFile := fmt.Sprintf("%s/dkim/%s/default.val", path, domain)

	if tools.IsExist(priFile) {
		pubContent, _ := tools.ReadFile(defalutTextFile)
		return pubContent, nil
	}

	Priv, Pub, err := makeRsa()
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: Priv,
	}

	// pri := b64.StdEncoding.EncodeToString(Priv)
	file, err := os.Create(priFile)
	if err != nil {
		return "", err
	}

	err = pem.Encode(file, block)
	if err != nil {
		return "", err
	}

	pub := b64.StdEncoding.EncodeToString(Pub)
	pubContent := fmt.Sprintf("default._domainkey\tIN\tTXT\t(\r\nv=DKIM1;k=rsa;p=%s\r\n)\r\n----- DKIM key default for %s", pub, domain)

	err = tools.WriteFile(defalutTextFile, pubContent)
	err = tools.WriteFile(defalutValFile, fmt.Sprintf("v=DKIM1;k=rsa;p=%s", pub))

	return pubContent, err
}

func MakeDkimConfFile(path, domain string) (string, error) {
	pDir := fmt.Sprintf("%s/dkim", path)
	if b := tools.IsExist(pDir); !b {
		err := os.MkdirAll(pDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	pathDir := fmt.Sprintf("%s/dkim/%s", path, domain)
	if b := tools.IsExist(pathDir); !b {
		err := os.MkdirAll(pathDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return MakeDkimFile(path, domain)
}

func GetDomainDkimVal(path, domain string) (string, error) {
	_, _ = MakeDkimConfFile(path, domain)
	defalutValFile := fmt.Sprintf("%s/dkim/%s/default.val", path, domain)
	pubContentRecord, err := tools.ReadFile(defalutValFile)
	return pubContentRecord, err
}
