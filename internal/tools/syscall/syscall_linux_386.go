package syscall

import (
	"syscall"
)

func Dup2(from int, to int) {
	syscall.Dup2(from, to)
}
