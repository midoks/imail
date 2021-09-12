package libs

import (
	"fmt"
	"os"
	"os/exec"
)

func ExecPython(scriptName string, id int64) (string, error) {

	cpath, _ := os.Getwd()
	fileName := fmt.Sprintf("%s/hook/%s", cpath, scriptName)
	_, b := IsExists(fileName)
	// fmt.Println(fileName, b)
	if !b {
		return
	}

	cmd := exec.Command("python", fileName, string(id))
	out, err := cmd.CombinedOutput()
	return out, err
}
