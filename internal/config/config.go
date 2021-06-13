package config

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"reflect"
	"unsafe"
)

var app_config *toml.Tree
var err error

func Load(path string) error {

	app_config, err = toml.LoadFile(path) //load config file

	if err != nil {
		fmt.Println("config init error:", err, app_config)
		return err
	}

	fmt.Println("config file init success!")
	return nil
}

func GetString(key string, def string) string {
	v := app_config.Get(key)

	if reflect.TypeOf(v) == nil {
		return def
	}

	return v.(string)
}

func GetInt64(key string, def int64) (int64, error) {
	v := app_config.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "int64" {
		return def, errors.New(key + " type is error, expect is int64 type!")
	}
	return v.(int64), nil
}

func GetInt(key string, def int) (int, error) {
	v := app_config.Get(key)

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
	v := app_config.Get(key)

	if reflect.TypeOf(v) == nil {
		return def, nil
	}

	if reflect.TypeOf(v).String() != "bool" {
		return def, errors.New(key + " type is error, expect is bool type!")
	}

	return v.(bool), nil
}
