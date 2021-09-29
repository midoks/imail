package tools

import (
	"testing"
)

func TestWrap(t *testing.T) {

	cases := map[string]string{
		"foobar":         "foobar",
		"foobar quux":    "foobar quux",
		"foobar\r\n":     "foobar\r\n",
		"foobar\r\nquux": "foobar\r\nquux",
		"foobar quux foobar quux foobar quux foobar quux foobar quux foobar quux foobar quux foobar quux":      "foobar quux foobar quux foobar quux foobar quux foobar quux foobar quux foobar\r\n\tquux foobar quux",
		"foobar quux foobar quux foobar quux foobar quux foobar quux foobar\r\n\tquux foobar quux foobar quux": "foobar quux foobar quux foobar quux foobar quux foobar quux foobar\r\n\tquux foobar quux foobar quux",
	}

	for k, v := range cases {
		if string(Wrap([]byte(k))) != v {
			t.Fatal("Didn't wrap correctly.")
		}
	}

}
