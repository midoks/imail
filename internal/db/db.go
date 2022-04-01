package db

import (
    "database/sql"
    "errors"
    "fmt"
    "os"
    "time"

    "github.com/midoks/imail/internal/conf"
    "github.com/midoks/imail/internal/log"
    "gorm.io/driver/mysql"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

var (
    db  *gorm.DB
    err error
)

func getEngine() (*sql.DB, error) {
    switch conf.Database.Type {
    case "mysql":
        dbUser := conf.Database.User
        dbPwd := conf.Database.Password
        dbHost := conf.Database.Host

        dbName := conf.Database.Name
        dbCharset := conf.Database.Charset

        dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=True", dbUser, dbPwd, dbHost, dbName, dbCharset)
        db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

    case "sqlite3":
        dbPath := conf.Web.AppDataPath + "/" + conf.Database.Path
        os.MkdirAll(conf.Web.AppDataPath, os.ModePerm)
        db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{SkipDefaultTransaction: true, PrepareStmt: true})
        //&gorm.Config{SkipDefaultTransaction: true,}

        // synchronous close
        db.Exec("PRAGMA synchronous = OFF;")
    default:
        return nil, errors.New("database type not found")
    }

    if err != nil {
        log.Errorf("init db err,link error:%s", err)
        return nil, err
    }

    sqlDB, err := db.DB()
    if err != nil {
        log.Errorf("[DB]:%s", err)
        return nil, err
    }

    sqlDB.SetMaxIdleConns(conf.Database.MaxIdleConns)
    sqlDB.SetMaxOpenConns(conf.Database.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(time.Hour)

    return sqlDB, nil
}

func Init() error {
    _, err := getEngine()
    if err != nil {
        return err
    }

    db.AutoMigrate(&User{})
    db.AutoMigrate(&Domain{})
    db.AutoMigrate(&Mail{})
    db.AutoMigrate(&MailLog{})
    db.AutoMigrate(&MailContent{})
    db.AutoMigrate(&Box{})
    db.AutoMigrate(&Class{})
    db.AutoMigrate(&Queue{})

    return nil
}

type Statistic struct {
    Counter struct {
        User      int64
        VaildUser int64
    }
}

func Ping() error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    return sqlDB.Ping()
}

func GetStatistic() (stats Statistic) {

    //user count
    stats.Counter.User = UsersCount()

    //vaild user count
    stats.Counter.VaildUser = UsersVaildCount()
    return stats
}

func CheckDb() bool {
    if db != nil {
        return true
    }
    return false
}
