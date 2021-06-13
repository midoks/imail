package dkim

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
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
