package tools

import (
	"regexp"
)

func IsEmailRe(b string) bool {
	var re = regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)
	return re.MatchString(b)
}

func IsUrlRe(b string) bool {
	var re = regexp.MustCompile(`[a-zA-z]+://[^\s]*`)
	return re.MatchString(b)
}

func IsIpv4Re(b string) bool {
	var re = regexp.MustCompile(`(((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)`)
	return re.MatchString(b)
}

func IsCodeRe(b string) bool {
	var re = regexp.MustCompile(`[1-9][\d]5`)
	return re.MatchString(b)
}
