package controllers

import (
// "github.com/astaxie/beego"
// "strconv"
// "strings"
// "time"
)

const (
	MSG_OK  = 0
	MSG_ERR = -1
)

type Index struct {
	Common
}

func (this *Index) Index() {

	this.display()
}
