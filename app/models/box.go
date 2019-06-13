package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"math"
	"strconv"
	"time"
)

type UserMailBox struct {
	Id         int64
	Uid        int64 `comment(用户ID)"`
	Mid        int64 `comment(邮件ID)"`
	Type       int   `comment(类型|0:接收邮件;1:发送邮件)`
	Size       int   `comment(邮件内容大小[byte])`
	UpdateTime int64
	CreateTime int64
}

func BoxTableName() string {
	return "im_user_box"
}

func (u *UserMailBox) TableName() string {
	return BoxTableName()
}

func (u *UserMailBox) Update(fields ...string) error {
	u.UpdateTime = time.Now().Unix()
	if _, err := orm.NewOrm().Update(u, fields...); err != nil {
		return err
	}
	return nil
}

func BoxAdd(uid int64, mid int64, method int, size int) (int64, error) {
	data := new(UserMailBox)
	data.Uid = uid
	data.Mid = mid
	data.Size = size
	data.Type = method

	data.UpdateTime = time.Now().Unix()
	data.CreateTime = time.Now().Unix()
	i, err := orm.NewOrm().Insert(data)
	if err != nil {
		return 0, err
	}
	return i, err
}

func BoxUserTotal(uid int64) (int64, int64) {
	var maps []orm.Params

	o := orm.NewOrm()
	sql := fmt.Sprintf("SELECT count(uid) as count, sum(size) as size FROM `%s` WHERE uid=?", BoxTableName())
	num, err := o.Raw(sql, uid).Values(&maps)
	if err == nil && num > 0 {
		count, err := strconv.ParseInt(maps[0]["count"].(string), 10, 32)
		if err != nil {
			count = 0
		}
		size, err := strconv.ParseInt(maps[0]["size"].(string), 10, 32)
		if err != nil {
			size = 0
		}
		return count, size
	}
	return 0, 0
}

func BoxPop3Pos(uid int64, pos int64) ([]orm.Params, error) {
	var maps []orm.Params

	o := orm.NewOrm()
	sql := fmt.Sprintf("SELECT mid,size FROM `%s` WHERE uid=? order by id limit %d,%d", BoxTableName(), pos-1, 1)
	_, err := o.Raw(sql, uid).Values(&maps)
	return maps, err
}

// Paging List of POP3 Protocol
func BoxPop3List(uid int64, page int, pageSize int) ([]orm.Params, error) {
	var maps []orm.Params

	offset := (page - 1) * pageSize
	o := orm.NewOrm()
	sql := fmt.Sprintf("SELECT mid,size FROM `%s` WHERE uid=? order by id limit %d,%d", BoxTableName(), offset, pageSize)
	_, err := o.Raw(sql, uid).Values(&maps)
	return maps, err
}

// POP3 gets all the data
func BoxPop3All(uid int64) []orm.Params {
	var maps []orm.Params
	count, _ := BoxUserTotal(uid)
	pageSize := 100
	page := int(math.Ceil(float64(count) / float64(pageSize)))

	var num = 1
	for i := 1; i <= page; i++ {
		list, _ := BoxPop3List(uid, i, pageSize)
		for i := 0; i < len(list); i++ {
			list[i]["num"] = strconv.Itoa(num)
			maps = append(maps, list[i])
		}
		num += 1
	}
	// fmt.Println("count:", count, page, maps)
	return maps
}
