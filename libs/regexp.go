package libs

import (
	"html"
	"regexp"
	"strings"
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

func TrimHtml(src string) string {
	src = html.UnescapeString(src)

	//将HTML标签全转换成小写
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)
	//去除STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")
	//去除SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")
	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "")
	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "")
	return strings.TrimSpace(src)
}
