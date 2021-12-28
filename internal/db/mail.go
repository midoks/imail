package db

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/midoks/imail/internal/tools/mail"
)

type Mail struct {
	Id                int64  `gorm:"primaryKey"`
	Uid               int64  `gorm:"comment:用户ID"`
	Type              int    `gorm:"comment:0:发送;1:接到"`
	MailFrom          string `gorm:"size:50;comment:邮件来源"`
	MailFromInContent string `gorm:"text;comment:邮件来源"`
	MailTo            string `gorm:"size:50;comment:接收邮件"`
	Subject           string `gorm:"size:250;comment:标题"`
	Content           string `gorm:"comment:邮件内容"`
	Size              int    `gorm:"size:50;comment:邮件内容大小"`
	Status            int    `gorm:"comment:0:准备发送;1:发送成功;2:发送失败;3:已接收"`

	IsRead   bool `gorm:"default:0;comment:是否已读"`
	IsDelete bool `gorm:"default:0;comment:是否删除"`
	IsFlags  bool `gorm:"default:0;comment:是否星标"`
	IsJunk   bool `gorm:"default:0;comment:是否无用"`
	IsDraft  bool `gorm:"default:0;comment:是否草稿"`
	IsCheck  bool `gorm:"default:0;comment:是否通过检查"`

	Created     time.Time `gorm:"autoCreateTime;comment:创建时间"`
	CreatedUnix int64     `gorm:"autoCreateTime;comment:创建时间"`
	Updated     time.Time `gorm:"autoCreateTime;comment:更新时间"`
	UpdatedUnix int64     `gorm:"autoCreateTime;comment:更新时间"`
}

const (
	MailSearchOptionsTypeSend = iota
	MailSearchOptionsTypeInbox
	MailSearchOptionsTypeDeleted
	MailSearchOptionsTypeFlags
	MailSearchOptionsTypeJunk
	MailSearchOptionsTypeUnread
)

type DIRType int32

const (
	DIR_DELETED DIRType = 0
	DIR_JUNK    DIRType = 1
	DIR_READ    DIRType = 2
	DIR_FLAGS   DIRType = 3
)

func MailTableName() string {
	return "im_mail"
}

func (*Mail) TableName() string {
	return MailTableName()
}

func MailCount() int64 {
	var count int64
	db.Model(&Mail{}).Count(&count)
	return count
}

func MailCountWithOpts(opts *MailSearchOptions) int64 {
	var count int64
	dbm := db.Model(&Mail{})
	dbm = MailSearchByNameCond(opts, dbm)
	dbm.Where("uid=?", opts.Uid).Count(&count)
	return count
}

func MailList(page, pageSize int, opts *MailSearchOptions) ([]*Mail, error) {
	mail := make([]*Mail, 0, pageSize)
	dbm := db.Limit(pageSize).Offset((page - 1) * pageSize).Order("id desc")
	dbm = MailSearchByNameCond(opts, dbm)

	err := dbm.Where("uid=?", opts.Uid).Find(&mail)
	return mail, err.Error
}

type MailSearchOptions struct {
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int
	Type     int
	Uid      int64
}

func MailSearchByNameCond(opts *MailSearchOptions, dbm *gorm.DB) *gorm.DB {
	if opts.Type == MailSearchOptionsTypeSend {
		dbm = dbm.Where("type = ?", 0).
			Where("is_junk = ?", 0).
			Where("is_flags = ?", 0)
	}

	if opts.Type == MailSearchOptionsTypeInbox {
		dbm = dbm.Where("type = ?", 1).
			Where("is_junk = ?", 0)
	}

	if opts.Type == MailSearchOptionsTypeDeleted {
		dbm = dbm.Where("is_delete = ?", 1)
	}

	if opts.Type == MailSearchOptionsTypeJunk {
		dbm = dbm.Where("is_junk = ?", 1)
	}

	if opts.Type == MailSearchOptionsTypeFlags {
		dbm = dbm.Where("is_flags = ?", 1)
	}

	return dbm
}

func MailSearchByName(opts *MailSearchOptions) (user []*Mail, _ int64, _ error) {
	if len(opts.Keyword) == 0 {
		return user, 0, nil
	}

	opts.Keyword = strings.ToLower(opts.Keyword)

	if opts.PageSize <= 0 || opts.PageSize > 20 {
		opts.PageSize = 10
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	searchQuery := "%" + opts.Keyword + "%"
	email := make([]*Mail, 0, opts.PageSize)

	dbm := db.Model(&Mail{}).Where("LOWER(subject) LIKE ?", searchQuery)
	dbm = MailSearchByNameCond(opts, dbm)
	err := dbm.Where("uid=?", opts.Uid).Find(&email)
	return email, MailCountWithOpts(opts), err.Error
}

func MailStatInfoForImap(uid int64, mtype int64) (int64, int64) {
	return MailStatInfo(uid, mtype)
}

func MailStatInfoForPop(uid int64) (int64, int64) {
	return MailStatInfo(uid, 0)
}

func MailStatInfo(uid int64, mtype int64) (int64, int64) {
	type Result struct {
		Count int64
		Size  int64
	}
	var result Result
	sql := fmt.Sprintf("SELECT count(uid) as count, sum(size) as size FROM `%s` WHERE uid=? and type=%d", MailTableName(), mtype)
	num := db.Raw(sql, uid).Scan(&result)

	if num.Error != nil {
		return 0, 0
	}

	return result.Count, result.Size
}

func MailListForPop(uid int64) []Mail {

	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? and type=1 order by created_unix desc", MailTableName())
	_ = db.Raw(sql, uid).Find(&result)

	return result
}

func MailListForImap(uid int64) []Mail {

	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? order by created_unix desc", MailTableName())
	_ = db.Raw(sql, uid).Find(&result)

	return result
}

func MailSendListForStatus(status int64, limit int64) []Mail {
	var result []Mail
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE status=%d and type=0 order by created_unix limit %d", MailTableName(), status, limit)
	db.Raw(sql).Find(&result)
	return result
}

func MailListPosForPop(uid int64, pos int64) ([]Mail, error) {
	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? and type=1 order by id limit %d,%d", MailTableName(), pos-1, 1)
	ret := db.Raw(sql, uid).Scan(&result)

	// fmt.Println(sql, result)
	if ret.Error != nil {
		return nil, ret.Error
	}
	return result, nil
}

func MailListForRspamd(limit int64) []Mail {
	var result []Mail
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE type=1 and is_check=0 order by id desc limit %d", MailTableName(), limit)
	db.Raw(sql).Find(&result)
	return result
}

func MailListAllForPop(uid int64) ([]Mail, error) {

	var result []Mail
	sql := fmt.Sprintf("SELECT id,size FROM `%s` WHERE uid=? and type=1 order by id limit 100", MailTableName())
	ret := db.Raw(sql, uid).Scan(&result)
	// fmt.Println(sql, result)
	if ret.Error != nil {
		return nil, ret.Error
	}
	return result, nil
}

func MailDeletedListAllForImap(uid int64) ([]Mail, error) {

	var result []Mail
	sql := fmt.Sprintf("SELECT id FROM `%s` WHERE uid=? and is_delete=1 order by id limit 10", MailTableName())
	ret := db.Raw(sql, uid).Scan(&result)
	if ret.Error != nil {
		return nil, ret.Error
	}
	return result, nil
}

func MailPosContentForPop(uid int64, pos int64) (string, int, error) {
	var result []Mail
	sql := fmt.Sprintf("SELECT uid,content,size FROM `%s` WHERE uid=? and type=1 order by id limit %d,%d", MailTableName(), pos-1, 1)
	ret := db.Raw(sql, uid).Scan(&result)

	if ret.Error != nil {
		return "", 0, ret.Error
	}

	return result[0].Content, result[0].Size, nil
}

func MailDeleteById(id int64, status int64) bool {

	var result []Mail
	sql := fmt.Sprintf("SELECT id FROM `%s` WHERE is_delete=1 and id='%d' order by id limit 1", MailTableName(), id)
	ret := db.Raw(sql).Scan(&result)
	if ret.Error == nil {
		if len(result) > 0 && status == 1 {
			MailHardDeleteById(id)
			return true
		}
	}

	db.Model(&Mail{}).Where("id = ?", id).Update("is_delete", status)
	return true
}

//用过ID获取邮件的全部信息
func MailById(id int64) (Mail, error) {
	var m Mail
	result := db.Model(&Mail{}).Where("id=?", id).Take(&m)
	return m, result.Error
}

func MailSoftDeleteById(id int64) bool {
	db.Model(&Mail{}).Where("id = ?", id).Update("is_delete", 1)
	return true
}

func MailSoftDeleteByIds(ids []int64) bool {
	err := db.Model(&Mail{}).Where("id IN  ?", ids).Update("is_delete", 1).Error
	if err != nil {
		return false
	}
	return true
}

func MailHardDeleteById(id int64) bool {
	err := db.Where("id = ? and is_delete=1", id).Delete(&Mail{}).Error
	if err != nil {
		return false
	}
	return true
}

func MailSeenById(id int64) bool {
	ids := []int64{id}
	return MailSeenByIds(ids)
}

func MailSeenByIds(ids []int64) bool {
	err := db.Model(&Mail{}).Where("id IN  ?", ids).Update("is_read", 1).Error
	if err != nil {
		return false
	}
	return true
}

func MailUnSeenById(id int64) bool {
	err := db.Model(&Mail{}).Where("id = ?", id).Update("is_read", 0).Error
	if err != nil {
		return false
	}
	return true
}

func MailUnSeenByIds(ids []int64) bool {
	err := db.Model(&Mail{}).Where("id IN  ?", ids).Update("is_read", 0).Error
	if err != nil {
		return false
	}
	return true
}

func MailSetFlagsById(id int64, status int64) bool {
	err := db.Model(&Mail{}).Where("id = ?", id).Update("is_flags", status).Error
	if err != nil {
		return false
	}
	return true
}

func MailSetJunkById(id int64, status int64) bool {
	// fmt.Println("MailSetJunkById", id, status)
	db.Model(&Mail{}).Where("id = ?", id).Update("is_junk", status)
	return true
}

func MailSetJunkByIds(ids []int64, status int64) bool {
	err := db.Model(&Mail{}).Where("id IN  ?", ids).Update("is_junk", status).Error
	if err != nil {
		return false
	}
	return true
}

func MailSetIsCheckById(id int64, status int64) bool {
	db.Model(&Mail{}).Where("id = ?", id).Update("is_check", status)
	return true
}

func MailSetStatusById(id int64, status int64) bool {
	db.Model(&Mail{}).Where("id = ?", id).Update("status", status)
	return true
}

func MailPushSend(uid int64, mail_from string, mail_to string, content string) (int64, error) {
	return MailPush(uid, 0, mail_from, mail_to, content, 0)
}

func MailPushReceive(uid int64, mail_from string, mail_to string, content string) (int64, error) {
	return MailPush(uid, 1, mail_from, mail_to, content, 3)
}

func MailPush(uid int64, mtype int, mail_from string, mail_to string, content string, status int) (int64, error) {
	if uid == 0 {
		return 0, errors.New("user id is error!")
	}

	tx := db.Begin()

	subject := mail.GetMailSubject(content)
	mail_from_in_content := mail.GetMailFromInContent(content)

	m := Mail{
		Uid:               uid,
		Type:              mtype,
		MailFrom:          mail_from,
		MailFromInContent: mail_from_in_content,
		MailTo:            mail_to,
		Content:           content,
		Subject:           subject,
		Size:              len(content),
		Status:            status,
	}

	m.UpdatedUnix = time.Now().Unix()
	m.CreatedUnix = time.Now().Unix()
	result := db.Create(&m)

	if result.Error != nil {
		tx.Rollback()
	}

	tx.Commit()
	return m.Id, result.Error
}
