package models

import (
	_ "fmt"
	"github.com/astaxie/beego/orm"
	"time"
)

type UserMail struct {
	Id         int64
	Mail       string `json:"name";orm:"unique;size(50);comment(用户名)"`
	Content    string `json:"password";orm:"unique;size(50);comment(用户密码)"`
	Status     int
	UpdateTime int64
	CreateTime int64
}

func (u *UserMail) TableName() string {
	return "im_mail"
}

func (u *UserMail) Update(fields ...string) error {
	u.UpdateTime = time.Now().Unix()
	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
		return err
	}
	return nil
}
