package libs

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/axgle/mahonia"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"
)

func GetGoEol() string {
	// if "windows" == runtime.GOOS {
	// 	return "\r\n"
	// }
	return "\r\n"
}

func Md5(buf []byte) string {
	hash := md5.New()
	hash.Write(buf)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func Md5str(s string) string {
	return Md5([]byte(s))
}

func CheckStringIsExist(source string, check []string) bool {

	for _, s := range check {
		if strings.EqualFold(source, s) {
			return true
		}
	}
	return false
}

func SizeFormat(size float64) string {
	units := []string{"Byte", "KB", "MB", "GB", "TB"}
	n := 0
	for size > 1024 {
		size /= 1024
		n += 1
	}

	return fmt.Sprintf("%.2f %s", size, units[n])
}

func RandString(len int) string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

func RemoveDuplicatesAndEmpty(a []string) (ret []string) {
	a_len := len(a)
	for i := 0; i < a_len; i++ {
		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}

func GetHttpData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.New("资源获取错误!")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
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

func WriteFile(file string, content string) error {
	return ioutil.WriteFile(file, []byte(content), 0666)
}

func ReadFile(file string) (string, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0600)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	return string(b), err
}

func Base64encode(in string) string {
	encodeString := base64.StdEncoding.EncodeToString([]byte(in))
	return encodeString
}

func Base64decode(in string) (string, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return in, err
	}
	return string(decodeBytes), nil
}

func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

func FilterAddressBody(src string) string {
	s := strings.Split(src, "BODY")
	s = strings.Split(s[0], "SIZE")
	return strings.TrimSpace(s[0])
}

// is_numeric()
func IsNumeric(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
	case float32, float64, complex64, complex128:
		return true
	case string:
		str := val.(string)
		if str == "" {
			return false
		}
		// Trim any whitespace
		str = strings.Trim(str, " \\t\\n\\r\\v\\f")
		if str[0] == '-' || str[0] == '+' {
			if len(str) == 1 {
				return false
			}
			str = str[1:]
		}
		// hex
		if len(str) > 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X') {
			for _, h := range str[2:] {
				if !((h >= '0' && h <= '9') || (h >= 'a' && h <= 'f') || (h >= 'A' && h <= 'F')) {
					return false
				}
			}
			return true
		}
		// 0-9,Point,Scientific
		p, s, l := 0, 0, len(str)
		for i, v := range str {
			if v == '.' { // Point
				if p > 0 || s > 0 || i+1 == l {
					return false
				}
				p = i
			} else if v == 'e' || v == 'E' { // Scientific
				if i == 0 || s > 0 || i+1 == l {
					return false
				}
				s = i
			} else if v < '0' || v > '9' {
				return false
			}
		}
		return true
	}

	return false
}

func CheckStandardMail(src string) bool {
	_, err := mail.ParseAddress(src)
	if err != nil {
		// fmt.Println("mmm:", err)
		return false
	}
	// fmt.Println("mmm:", smail.Address, smail, err)
	if src[0:1] == "<" && src[len(src)-1:] == ">" {
		return true
	}
	return false
}

func GetRealMail(src string) string {
	return src[1 : len(src)-1]
}

func IsExists(path string) (os.FileInfo, bool) {
	f, err := os.Stat(path)
	return f, err == nil || os.IsExist(err)
}
