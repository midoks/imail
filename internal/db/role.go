package db

import (
	// "fmt"
	_ "github.com/midoks/imail/internal/tools"
	// "gorm.io/gorm"
	// "strings"
	// _ "time"
)

type Role struct {
	Id         int64  `gorm:"primaryKey"`
	Name       string `gorm:"unique;size:50;comment:名称"`
	Pid        int64  `gorm:"comment:PID"`
	Status     int    `gorm:"comment:状态"`
	UpdateTime int64  `gorm:"autoCreateTime;comment:更新时间"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间"`
}

func (Role) TableName() string {
	return "im_role"
}
