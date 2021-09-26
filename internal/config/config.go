package config

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"net"
	"reflect"
	"strings"
	"unsafe"
)

var confToml *toml.Tree
var err error

var IsLoadedVar bool

var (
	App struct {
		Version string
	}
)

func Load(path string) error {

	confToml, err = toml.LoadFile(path) //load config file

	if err != nil {
		panic("config init error")
		IsLoadedVar = false
		return err
	}
	IsLoadedVar = true
	http_enable, err := GetBool("http.enable", false)
	if err == nil && http_enable {
		ipWhiteList := strings.Split(GetString("http.ip_white", "*"), ",")
		if InSliceString("*", ipWhiteList) {
			return nil
		}
		for _, ip := range ipWhiteList {
			if strings.Contains(ip, "/") {
				_, _, err = net.ParseCIDR(ip)
				if err != nil {
					return err
				}
			} else {
				if net.ParseIP(ip) == nil {
					return errors.New(fmt.Sprint(ip, ", Invalid whitelist"))
				}
			}

		}

	}
	return nil
}

func LoadString(content string) error {
	confToml, err = toml.Load(content) //load config string
	return err
}

func InSliceString(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func IsLoaded() bool {
	return IsLoadedVar
}

func GetString(key string, def string) string {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def
	}

	return v.(string)
}

func GetInt64(key string, def int64) (int64, error) {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "int64" {
		return def, errors.New(key + " type is error, expect is int64 type!")
	}
	return v.(int64), nil
}

func GetInt(key string, def int) (int, error) {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "int64" {
		return def, errors.New(key + " type is error, expect is int type!")
	}

	vv := v.(int64)

	vInt := *(*int)(unsafe.Pointer(&vv))
	return vInt, nil
}

func GetFloat64(key string, def float64) (float64, error) {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "float64" {
		return def, errors.New(key + " type is error, expect is float64 type!")
	}
	return v.(float64), nil
}

func GetBool(key string, def bool) (bool, error) {
	v := confToml.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "bool" {
		return def, errors.New(key + " type is error, expect is bool type!")
	}

	return v.(bool), nil
}
