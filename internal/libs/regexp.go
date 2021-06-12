package libs

import (
	"regexp"
)

func IsEmailRe(b string) bool {
	var re = regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)
	return re.Match([]byte(b))
}

func IsUrlRe(b string) bool {
	var re = regexp.MustCompile(`[a-zA-z]+://[^\s]*`)
	return re.Match([]byte(b))
}

func IsIpv4Re(b string) bool {
	var re = regexp.MustCompile(`(((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)`)
	return re.Match([]byte(b))
}

func IsCodeRe(b string) bool {
	var re = regexp.MustCompile(`[1-9][\d]5`)
	return re.Match([]byte(b))
}
