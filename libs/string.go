package libs

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/axgle/mahonia"
	"os"
	"strings"
)

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
	req := httplib.Get(url)

	str, err := req.String()
	if err != nil {
		return "", errors.New("资源获取错误!")
	}

	return str, nil
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
