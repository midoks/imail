package db

import (
	"fmt"
	_ "github.com/midoks/imail/internal/libs"
	// "gorm.io/gorm"
	"strings"
	// "errors"
	// "time"
)

type Box struct {
	Id         int64 `gorm:"primaryKey"`
	Uid        int64 `gorm:"comment:用户ID"`
	Mid        int64 `gorm:"comment:邮件ID"`
	Type       int   `gorm:"comment:类型|0:接收邮件;1:发送邮件"`
	Size       int   `gorm:"comment:邮件内容大小[byte]"`
	UpdateTime int64 `gorm:"autoCreateTime;comment:更新时间"`
	CreateTime int64 `gorm:"autoCreateTime;comment:创建时间"`
}

func BoxTableName() string {
	return "im_user_box"
}
func (Box) TableName() string {
	return BoxTableName()
}

func BoxUserList(uid int64) (int64, int64) {

	var resultBox Box
	sql := fmt.Sprintf("SELECT count(uid) as count, sum(size) as size FROM `%s` WHERE uid=?", BoxTableName())
	num := db.Raw(sql, uid).Find(&resultBox)

	fmt.Println(num, resultBox)

	// count, err := strconv.ParseInt(maps[0]["count"].(string), 10, 64)
	// if err != nil {
	// 	count = 0
	// }

	// if err == nil && num > 0 && count > 0 {

	// 	size, err := strconv.ParseInt(maps[0]["size"].(string), 10, 64)
	// 	if err != nil {
	// 		size = 0
	// 	}
	// 	return count, size
	// }
	return 0, 0
}

func BoxUserTotal(uid int64) (int64, int64) {

	var resultBox Box
	sql := fmt.Sprintf("SELECT count(uid) as count, sum(size) as size FROM `%s` WHERE uid=?", BoxTableName())
	num := db.Raw(sql, uid).Find(&resultBox)

	fmt.Println(num, resultBox)

	// count, err := strconv.ParseInt(maps[0]["count"].(string), 10, 64)
	// if err != nil {
	// 	count = 0
	// }

	// if err == nil && num > 0 && count > 0 {

	// 	size, err := strconv.ParseInt(maps[0]["size"].(string), 10, 64)
	// 	if err != nil {
	// 		size = 0
	// 	}
	// 	return count, size
	// }
	return 0, 0
}

// func BoxAdd(uid int64, mid int64, method int, size int) (int64, error) {

// 	boxData := Box{
// 		Uid:  uid,
// 		Mid:  mid,
// 		Size: size,
// 		Type: method,
// 	}

// 	boxData.UpdateTime = time.Now().Unix()
// 	boxData.CreateTime = time.Now().Unix()
// 	result := db.Create(&boxData)

// 	return 0, errors.New("error")
// }

// 获取分类下的统计数据
func BoxUserMessageCountByClassName(uid int64, className string) (int64, error) {
	fmt.Println("db[BoxUserMessageCountByClassName]", uid, className)

	if strings.EqualFold(className, "INBOX") {
		count, size := MailStatInfoForPop(uid)
		fmt.Println("db[BoxUserMessageCountByClassName]", count, size)
		return count, nil
	}
	// cid, err := ClassGetIdByName(uid, className)
	// if err == nil {
	// 	return BoxUserMessageCountByCid(uid, cid)
	// }
	return 0, nil
}

// func BoxUserMessageCountByCid(uid int64, cid int64) (int64, error) {
// 	var boxData Box
// 	sql := fmt.Sprintf("SELECT count(uid) as count FROM `%s` WHERE uid=? and cid=?", BoxTableName())
// 	result := db.Raw(sql, uid, cid).First(&boxData)
// 	fmt.Println(result)
// 	// if err == nil && num > 0 {
// 	// 	count, err := strconv.ParseInt(maps[0]["count"].(string), 10, 64)
// 	// 	return count, err
// 	// }
// 	return 0, nil
// }

// func BoxPos(uid int64, pos int64) ([]orm.Params, error) {
// 	var boxData Box
// 	sql := fmt.Sprintf("SELECT mid,size FROM `%s` WHERE uid=? order by id limit %d,%d", BoxTableName(), pos-1, 1)
// 	result := db.Raw(sql, uid).First(&boxData)
// 	return boxData, nil
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

// func BoxPosContent(uid int64, pos int64) (string, string, error) {

// 	sql := fmt.Sprintf("SELECT mid,size FROM `%s` WHERE uid=? order by id limit %d,%d", BoxTableName(), pos-1, 1)
// 	_, err := db.Raw(sql, uid).Values(&maps)

// 	if err != nil {
// 		return "", "", err
// 	}
// 	size := maps[0]["size"].(string)
// 	var content []orm.Params
// 	mid := maps[0]["mid"]
// 	sql = fmt.Sprintf("SELECT content FROM `%s` WHERE id=?", MailTableName())
// 	_, err = o.Raw(sql, mid).Values(&content)
// 	text := content[0]["content"].(string)
// 	return text, size, nil
// }

// // Paging List of POP3 Protocol
// func BoxList(uid int64, cid int64, page int, pageSize int) ([]orm.Params, error) {

// 	offset := (page - 1) * pageSize
// 	sql := fmt.Sprintf("SELECT mid,size FROM `%s` WHERE uid=? and cid=? order by id limit %d,%d", BoxTableName(), offset, pageSize)
// 	_, err := db.Raw(sql, uid, cid).Find(&maps)
// 	return maps, err
// }

// // Paging List of POP3 Protocol
func BoxListSE(uid int64, className string, start int64, end int64) ([]Mail, error) {
	var result []Mail

	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE uid=? and id>='%d' and id<='%d'", "im_mail", start, end)
	fmt.Println("BoxListSE..:", sql)
	db.Raw(sql, uid).Find(&result)
	return result, err
}

// // POP3 gets all the data
// func BoxAll(uid int64, cid int64) []orm.Params {

// 	count, _ := BoxUserTotal(uid)
// 	pageSize := 100
// 	page := int(math.Ceil(float64(count) / float64(pageSize)))

// 	var num = 1
// 	for i := 1; i <= page; i++ {
// 		list, _ := BoxList(uid, cid, i, pageSize)
// 		for i := 0; i < len(list); i++ {
// 			list[i]["num"] = strconv.Itoa(num)
// 			maps = append(maps, list[i])
// 		}
// 		num += 1
// 	}
// 	// fmt.Println("count:", count, page, maps)
// 	return maps
// }

// // POP3 gets all the data
// func BoxAllByClassName(uid int64, className string) ([]orm.Params, error) {
// 	cid, err := ClassGetIdByName(uid, className)
// 	if err == nil {
// 		return BoxAll(uid, cid), nil
// 	}
// 	return maps, err
// }
