package config

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"reflect"
	"unsafe"
)

var confToml *toml.Tree
var err error

func Load(path string) error {

	confToml, err = toml.LoadFile(path) //load config file

	if err != nil {
		fmt.Println("config init error:", err, confToml)
		return err
	}

	fmt.Println("config file init success!")
	return nil
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
