package libs

import (
	"errors"
	"fmt"
	"github.com/midoks/imail/internal/conf"
	"os"
	"os/exec"
)

func ExecPython(scriptName string, id int64) (string, error) {
	hookEnable, _ := conf.GetBool("hook.enable", false)
	if !hookEnable {
		return "", errors.New("config is disable!")
	}

	cpath, _ := os.Getwd()
	fileName := fmt.Sprintf("%s/conf/hook/%s", cpath, scriptName)
	_, b := IsExists(fileName)
	// fmt.Println(fileName, b)
	if !b {
		return "", errors.New("file is not exist!")
	}

	cmd := exec.Command("python", fileName, fmt.Sprint(id))
	out, err := cmd.CombinedOutput()
	return string(out), err
}
