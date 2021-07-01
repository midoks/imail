package db

import (
	// "errors"
	"fmt"
	"time"
)

type Mail struct {
	Id         int64  `gorm:"primaryKey"`
	Uid        int64  `gorm:"comment:用户ID"`
	Type       int    `gorm:"comment:0:发送;1:收到"`
	MailFrom   string `gorm:"size:50;comment:邮件来源"`
	MailTo     string `gorm:"size:50;comment:接收邮件"`
	Content    string `gorm:"comment:邮件内容"`
	Size       int    `gorm:"size:50;comment:邮件内容大小"`
	Status     int    `gorm:"comment:0:准备发送;1:发送成功;2:发送失败;3:已接收"`
	IsRead     int    `gorm:"default:0;comment:是否已读"`
	IsDelete   int    `gorm:"default:0;comment:是否删除"`
	IsFlags    int    `gorm:"default:0;comment:是否星标"`
	IsJunk     int    `gorm:"default:0;comment:是否无用"`
	UpdateTime int64  `gorm:"autoCreateTime;comment:更新时间"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间"`
}

func MailTableName() string {
	return "im_mail"
}

func (Mail) TableName() string {
	return MailTableName()
}

//用户统计邮件信息[POP3]
func MailStatInfoForPop(uid int64) (int64, int64) {
	type Result struct {
		Count int64
		Size  int64
	}
	var result Result
	sql := fmt.Sprintf("SELECT count(uid) as count, sum(size) as size FROM `%s` WHERE uid=?", MailTableName())
	num := db.Raw(sql, uid).Scan(&result)

	if num.Error != nil {
		return 0, 0
	}

	// fmt.Println("MailListForPop:", num, result)
	return result.Count, result.Size
}

func MailListForPop(uid int64) []Mail {

	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? order by create_time desc", MailTableName())
	num := db.Raw(sql, uid).Find(&result)

	fmt.Println("MailListForPop:", num, result)

	return result
}

func MailListPosForPop(uid int64, pos int64) ([]Mail, error) {
	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? order by id limit %d,%d", MailTableName(), pos-1, 1)
	ret := db.Raw(sql, uid).Scan(&result)

	fmt.Println(sql, result)
	if ret.Error != nil {
		return nil, ret.Error
	}
	return result, nil
}

func MailListAllForPop(uid int64) ([]Mail, error) {

	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? order by id", MailTableName())
	ret := db.Raw(sql, uid).Scan(&result)

	fmt.Println(sql, result)
	if ret.Error != nil {
		return nil, ret.Error
	}
	return result, nil
}

func MailPosContentForPop(uid int64, pos int64) (string, int, error) {
	var result []Mail
	sql := fmt.Sprintf("SELECT uid,content,size FROM `%s` WHERE uid=? order by id limit %d,%d", MailTableName(), pos-1, 1)
	ret := db.Raw(sql, uid).Scan(&result)

	if ret.Error != nil {
		return "", 0, ret.Error
	}

	return result[0].Content, result[0].Size, nil
}

// func BoxPosTop(uid int64, pos int64, line int64) ([]Mail, error) {
// 	var result []Mail
// 	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? order by id limit %d,%d", MailTableName(), pos-1, 1)
// 	ret := db.Raw(sql, uid).Scan(&result)

// 	fmt.Println(sql, result)
// 	if ret.Error != nil {
// 		return nil, ret.Error
// 	}
// 	return result, nil
// }

// func BoxPosTop(uid int64, pos int64, line int64) (string, string, error) {
// 	text, size, err := BoxPosContent(uid, pos)

// 	if err != nil {
// 		return "", size, err
// 	}

// 	textSplit := strings.SplitN(text, "\r\n\r\n", 2)
// 	if line == 0 {
// 		return textSplit[0] + "\r\n.\r\n", size, nil
// 	}
// 	return "", size, err
// }

func MailSoftDeleteById(id int64) bool {
	db.Model(&Mail{}).Where("id = ?", id).Update("is_delete", 1)
	return false
}

func MailHardDeleteById(id int64) bool {
	db.Where("id = ?", id).Delete(&Mail{})
	return false
}

func MailSeenById(id int64) bool {
	db.Model(&Mail{}).Where("id = ?", id).Update("is_read", 1)
	return false
}

func MailUnSeenById(id int64) bool {
	db.Model(&Mail{}).Where("id = ?", id).Update("is_read", 0)
	return false
}

func MailSetFlagsById(id int64, status int64) bool {
	db.Model(&Mail{}).Where("id = ?", id).Update("is_flags", status)
	return false
}

func MailSetJunkById(id int64, status int64) bool {
	db.Model(&Mail{}).Where("id = ?", id).Update("is_delete", status)
	return false
}

func MailPush(uid int64, mtype int, mail_from string, mail_to string, content string, status int) (int64, error) {

	user := Mail{
		Uid:      uid,
		Type:     mtype,
		MailFrom: mail_from,
		MailTo:   mail_to,
		Content:  content,
		Size:     len(content),
		Status:   status,
	}

	user.UpdateTime = time.Now().Unix()
	user.CreateTime = time.Now().Unix()
	result := db.Create(&user)

	return result.RowsAffected, result.Error
}
