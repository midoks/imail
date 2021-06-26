package imap

import (
	"crypto/tls"
	"errors"
	"fmt"
	"testing"
)

// go test -run TestLocalImap
// TestLocalImap
func TestLocalImap(t *testing.T) {

	c, err := Dial("127.0.0.1:143")
	if err != nil {
		t.Error("TestLocalImap fail:" + err.Error())
	}
	defer c.Close()

	// if supported, _ := c.Extension("AUTH"); supported {
	// 	t.Fatal("AUTH supported before TLS")
	// }

	if supported, _ := c.Extension("8BITMIME"); !supported {
		t.Fatal("8BITMIME not supported")
	}

}
