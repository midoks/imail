package models

import (
	_ "fmt"
	"github.com/astaxie/beego/orm"
	"time"
)

type User struct {
	Id         int64
	Name       string `json:"name";orm:"unique;size(50);comment(用户名)"`
	Password   string `json:"password";orm:"unique;size(50);comment(用户密码)"`
	Status     int
	UpdateTime int64
	CreateTime int64
}

func (u *User) TableName() string {
	return "im_user"
}

func (u *User) Update(fields ...string) error {
	u.UpdateTime = time.Now().Unix()
	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
		return err
	}
	return nil
}

func UserGetById(id int) (*User, error) {
	u := new(User)
	err := orm.NewOrm().QueryTable(u.TableName()).Filter("status", 0).Filter("id", id).One(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UserGetByName(name string) (*User, error) {
	u := new(User)
	err := orm.NewOrm().QueryTable(u.TableName()).Filter("name", name).One(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}
