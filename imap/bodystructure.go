package imap

import (
	"bufio"
	"fmt"
	"strings"
)

type BodyStructure struct {
}

func GetHeaderLine(r *bufio.Reader, line []byte) ([]byte, error) {
	for {
		l, more, err := r.ReadLine()

		if err != nil {
			break
		}
		fmt.Println("GetHeader:", string(l))
		fmt.Println("GetHeader--end")
		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}

func GetHeader(body string) {
	bufferedBody := bufio.NewReader(strings.NewReader(body))

	for {
		kv, err := GetHeaderLine(bufferedBody, nil)
		// fmt.Println(kv, err)
		if err != nil {
			break
		}

		if len(kv) == 0 {
			break
		}
	}

}
