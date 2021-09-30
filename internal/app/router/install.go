package router

import (
	"fmt"

	"github.com/midoks/imail/internal/conf"
)

func GlobalInit(customConf string) error {
	err := conf.Init(customConf)
	fmt.Println(err)
	return nil
}
