package db

import (
	// "fmt"
	"github.com/midoks/imail/internal/libs"
	"strings"
	_ "time"
)

type User struct {
	Id         int64  `gorm:"primaryKey"`
	Name       string `gorm:"unique;size:50;comment:用户名"`
	Password   string `gorm:"size:32;comment:用户密码"`
	Code       string `gorm:"size:50;comment:编码"`
	Role       int    `gorm:"comment:角色"`
	Token      string `gorm:"unique;size:50;comment:Token"`
	Status     int    `gorm:"comment:状态"`
	UpdateTime int64  `gorm:"autoCreateTime;comment:更新时间"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间"`
}

func (User) TableName() string {
	return "im_users"
}

func LoginWithCode(name string, code string) (bool, int64) {
	list := strings.SplitN(name, "@", 2)

	var user User
	err := db.First(&user, "name = ?", list[0]).Error

	if err != nil {
		return false, 0
	}

	if user.Code == code {
		return true, user.Id
	}

	return false, 0
}

func LoginByUserPassword(name string, password string, rand string) (bool, int64) {

	var user User
	err := db.First(&user, "name = ?", name).Error

	if err != nil {
		return false, 0
	}

	passMd5 := libs.Md5str(user.Password + rand)
	if passMd5 == password {
		return true, user.Id
	}

	return false, 0
}

func UserCheckIsExist(name string) bool {
	var user User
	err := db.First(&user, "name = ?", name).Error
	if err == nil {
		return true
	}
	return false
}

func UserUpdateTokenGetByName(name string, token string) bool {
	db.Model(&User{}).Where("name = ?", name).Update("token", token)
	return true
}

func UserUpdateCodeGetByName(name string, code string) bool {
	db.Model(&User{}).Where("name = ?", name).Update("code", code)
	return true
}

func UserGetByName(name string) (User, error) {
	list := strings.SplitN(name, "@", 2)
	var user User
	err := db.First(&user, "name = ?", list[0]).Error
	if err != nil {
		return user, err
	}
	return user, nil
}
