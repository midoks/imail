package models

import (
	_ "fmt"
	"time"

	"github.com/astaxie/beego/orm"
)

type ImUser struct {
	Id         int
	Username   string
	Password   string
	UpdateTime int64
	CreateTime int64
}

func getTnByUser() string {
	return "im_users"
}

func (u *ImUser) TableName() string {
	return getTnByUser()
}

func (u *ImUser) Update(fields ...string) error {
	u.UpdateTime = time.Now().Unix()
	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
		return err
	}
	return nil
}

func UserGetById(id int) (*ImUser, error) {
	u := new(ImUser)
	err := orm.NewOrm().QueryTable(getTnByUser()).Filter("id", id).One(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UserGetByName(username string) (*ImUser, error) {

	u := new(ImUser)
	err := orm.NewOrm().QueryTable(getTnByUser()).Filter("username", username).One(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UserDelById(id int) (int64, error) {
	return orm.NewOrm().Delete(&ImUser{Id: id})
}
