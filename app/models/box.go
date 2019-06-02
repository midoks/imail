package models

import (
	_ "fmt"
	"time"

	"github.com/astaxie/beego/orm"
)

type UserMailBox struct {
	Id         int64
	Uid        string `json:"name";orm:"unique;size(50);comment(用户名)"`
	Mid        string `json:"password";orm:"unique;size(50);comment(用户密码)"`
	UpdateTime int64
	CreateTime int64
}

func (u *UserMailBox) TableName() string {
	return "im_user_box"
}

func (u *UserMailBox) Update(fields ...string) error {
	u.UpdateTime = time.Now().Unix()
	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
		return err
	}
	return nil
}
