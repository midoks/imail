package db

import (
	"strings"
	"time"
)

type MailLog struct {
	Id          int64  `gorm:"primaryKey"`
	Type        string `gorm:"index;comment:type"`
	Content     string `gorm:"comment:content"`
	CreatedUnix int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedUnix int64  `gorm:"autoCreateTime;comment:更新时间"`
}

type LogSearchOptions struct {
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int
	TplName  string
}

func (*MailLog) TableName() string {
	return TablePrefix("log")
}

func LogList(page, pageSize int) ([]MailLog, error) {
	log := make([]MailLog, 0, pageSize)
	err := db.Limit(pageSize).Offset((page - 1) * pageSize).Order("id desc").Find(&log)
	return log, err.Error
}

func LogCount() int64 {
	var count int64
	db.Model(&MailLog{}).Count(&count)
	return count
}

func LogSearchByName(opts *LogSearchOptions) (user []MailLog, _ int64, _ error) {
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
	log := make([]MailLog, 0, opts.PageSize)

	err := db.Model(&MailLog{}).
		Where("LOWER(content) LIKE ?", searchQuery).
		Or("LOWER(type) LIKE ?", searchQuery).
		Find(&log)
	return log, LogCount(), err.Error
}

func LogAdd(ty, content string) error {

	m := MailLog{}
	m.Type = ty
	m.Content = content
	m.UpdatedUnix = time.Now().Unix()
	m.CreatedUnix = time.Now().Unix()
	result := db.Save(&m)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func LogDeleteById(id int64) error {
	var d MailLog
	return db.Where("id = ?", id).Delete(&d).Error
}

func LogClear() error {
	err := db.Exec("truncate table `im_log`")
	if err.Error != nil {
		return err.Error
	}
	return nil
}
