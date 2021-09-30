package mail

import (
	"errors"
	"fmt"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/tools"
	"os/exec"
)

func ExecPython(scriptName string, id int64) (string, error) {
	if !conf.Hook.Enable {
		return "", errors.New("config is disable!")
	}

	fileName := fmt.Sprintf("%s/conf/hook/%s", conf.WorkDir(), scriptName)
	b := tools.IsExist(fileName)
	// fmt.Println(fileName, b)
	if !b {
		return "", errors.New("file is not exist!")
	}

	cmd := exec.Command("python", fileName, fmt.Sprint(id))
	out, err := cmd.CombinedOutput()
	return string(out), err
}
