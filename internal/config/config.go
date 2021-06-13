package config

import (
	"fmt"
	"github.com/pelletier/go-toml"
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

func GetString(key string) string {
	v := app_config.Get(key).(string)
	// fmt.Println(v)
	return v
}

func GetInt64(key string) int64 {
	v := app_config.Get(key).(int64)
	return v
}

func GetInt(key string) int {
	v := app_config.Get(key).(int64)
	vInt := *(*int)(unsafe.Pointer(&v))
	return vInt
}

func GetBool(key string) bool {
	v := app_config.Get(key).(bool)
	return v
}
