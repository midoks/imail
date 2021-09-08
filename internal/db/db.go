package db

import (
  "fmt"
  "github.com/midoks/imail/internal/config"
  "github.com/midoks/imail/internal/log"
  "gorm.io/driver/mysql"
  "gorm.io/gorm"
  "time"
)

var db *gorm.DB
var err error

func Init() {

  dbUser := config.GetString("db.user", "root")
  dbPasswd := config.GetString("db.password", "root")
  dbHost := config.GetString("db.host", "127.0.0.1")
  dbPort, _ := config.GetInt64("db.port", 3306)

  dbName := config.GetString("db.name", "imail")
  dbCharset := config.GetString("db.charset", "utf8mb4")

  dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True", dbUser, dbPasswd, dbHost, dbPort, dbName, dbCharset)
  db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

  if err != nil {
    log.Errorf("init db err,link error:%s", err)
    return
  }

  log.Info("init db success!")

  sqlDB, sqlErr := db.DB()
  // SetMaxIdleConns 设置空闲连接池中连接的最大数量
  sqlDB.SetMaxIdleConns(10)
  // SetMaxOpenConns 设置打开数据库连接的最大数量。
  sqlDB.SetMaxOpenConns(100)
  // SetConnMaxLifetime 设置了连接可复用的最大时间。
  sqlDB.SetConnMaxLifetime(time.Hour)

  if sqlErr != nil {
    log.Error("[DB]", sqlErr)
    return
  }

  db.AutoMigrate(&User{})
  db.AutoMigrate(&UserLoginVerify{})
  db.AutoMigrate(&Mail{})
  db.AutoMigrate(&Box{})
  db.AutoMigrate(&Class{})
  db.AutoMigrate(&Role{})
  db.AutoMigrate(&Queue{})

  //创建默认账户
  var userAdmin User
  admin := db.First(&userAdmin, "name = ?", "admin")
  if admin.Error != nil {
    db.Create(&User{Name: "admin", Password: "21232f297a57a5a743894a0e4a801fc3", Code: "admin"})
  }

  //退信账户
  var userPostmaster User
  postmaster := db.First(&userPostmaster, "name = ?", "postmaster")
  if postmaster.Error != nil {
    db.Create(&User{Name: "postmaster", Password: "21232f297a57a5a743894a0e4a801fc3", Code: "postmaster"})
  }

  //管理员角色
  var role Role
  ruleResult := db.First(&role, "pid = ?", "0")
  if ruleResult.Error != nil {
    db.Create(&Role{
      Name:   "管理员",
      Pid:    0,
      Status: 1,
    })
  }
}

func CheckDb() bool {
  if db != nil {
    return true
  }
  return false
}
