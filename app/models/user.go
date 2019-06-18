package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/midoks/imail/libs"
	"strings"
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

func UserLogin(name string, password string) (bool, int64) {
	list := strings.SplitN(name, "@", 2)
	info, err := UserGetByName(list[0])
	if err != nil {
		return false, 0
	}

	pwd_md5 := libs.Md5str(password)
	fmt.Println("UserLogin", list[0], pwd_md5, info.Password)
	if !strings.EqualFold(pwd_md5, info.Password) {
		return false, 0
	}

	return true, info.Id
}
