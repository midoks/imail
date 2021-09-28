package db

import (
    "errors"
    "fmt"
    "github.com/midoks/imail/internal/config"
    "github.com/midoks/imail/internal/log"
    "gorm.io/driver/mysql"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "time"
)

var db *gorm.DB
var err error

func Init() error {
    switch config.GetString("db.type", "") {
    case "mysql":
        dbUser := config.GetString("db.user", "root")
        dbPasswd := config.GetString("db.password", "root")
        dbHost := config.GetString("db.host", "127.0.0.1")
        dbPort, _ := config.GetInt64("db.port", 3306)

        dbName := config.GetString("db.name", "imail")
        dbCharset := config.GetString("db.charset", "utf8mb4")

        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True", dbUser, dbPasswd, dbHost, dbPort, dbName, dbCharset)
        db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    case "sqlite3":
        dbPath := config.GetString("db.path", "./data/imail.db3")
        db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
            SkipDefaultTransaction: true,
        })

        // synchronous close
        db.Exec("PRAGMA synchronous = OFF; ")
    default:
        log.Errorf("database type not found")
        return errors.New("database type not found")
    }
    if err != nil {
        log.Errorf("init db err,link error:%s", err)
        return err
    }

    log.Info("init db success!")

    sqlDB, err := db.DB()
    // SetMaxIdleConns sets the maximum number of connections in the free connection pool
    sqlDB.SetMaxIdleConns(200)
    // SetMaxOpenConns sets the maximum number of open database connections.
    sqlDB.SetMaxOpenConns(500)
    // SetConnMaxLifetime Sets the maximum time that the connection can be reused.
    sqlDB.SetConnMaxLifetime(time.Hour)

    if err != nil {
        log.Errorf("[DB]:%s", err)
        return err
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
        db.Create(&User{Name: "admin", Password: "21232f297a57a5a743894a0e4a801fc3", Code: "admin", Token: "21232f297a57a5a743894a0e4a801fc3"})
    }

    //退信账户
    var userPostmaster User
    postmaster := db.First(&userPostmaster, "name = ?", "postmaster")
    if postmaster.Error != nil {
        db.Create(&User{Name: "postmaster", Password: "21232f297a57a5a743894a0e4a801fc3", Code: "postmaster", Token: "21232f297a57a5a743894a0e4a801fc2"})
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

    return nil
}

func CheckDb() bool {
    if db != nil {
        return true
    }
    return false
}
