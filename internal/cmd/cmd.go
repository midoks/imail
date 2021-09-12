package cmd

import (
	"errors"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/libs"
	"github.com/midoks/imail/internal/log"
	"github.com/urfave/cli"
	"os"
	"strings"
	"time"
)

func stringFlag(name, value, usage string) cli.StringFlag {
	return cli.StringFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	}
}

func boolFlag(name, usage string) cli.BoolFlag {
	return cli.BoolFlag{
		Name:  name,
		Usage: usage,
	}
}

//nolint:deadcode,unused
func intFlag(name string, value int, usage string) cli.IntFlag {
	return cli.IntFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	}
}

//nolint:deadcode,unused
func durationFlag(name string, value time.Duration, usage string) cli.DurationFlag {
	return cli.DurationFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	}
}

func initConfig(c *cli.Context, defineConf string) (string, error) {
	confFile := ""
	if !strings.EqualFold(defineConf, "") {
		confFile = defineConf
	} else {
		confFile = c.String("config")
		if confFile == "" {
			confFile = "conf/app.conf"
		}
	}

	_, f := libs.IsExists(confFile)

	if !f {
		definedConf, _ := libs.ReadFile("conf/app.defined.conf")
		libs.WriteFile(confFile, definedConf)
	}

	if _, err := os.Stat(confFile); err != nil {
		if os.IsNotExist(err) {
			return confFile, errors.New("imail config is not exist!")
		} else {
			return confFile, err
		}
	}

	err := config.Load(confFile)
	if err != nil {
		log.Infof("imail config file load err:%s", err)
		return confFile, errors.New("imail config file load err")
	}
	return confFile, nil
}
