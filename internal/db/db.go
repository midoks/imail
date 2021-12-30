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
        dbPath := conf.Database.Path
        os.MkdirAll(conf.WorkDir()+"/data", os.ModePerm)
        fmt.Println("sqlite3 Path:", dbPath)
        db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{SkipDefaultTransaction: true})
        //&gorm.Config{SkipDefaultTransaction: true,}

        // synchronous close
        db.Exec("PRAGMA synchronous = OFF;")
    default:
        log.Errorf("database type not found")
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

    // SetMaxIdleConns sets the maximum number of connections in the free connection pool
    sqlDB.SetMaxIdleConns(conf.Database.MaxIdleConns)
    // SetMaxOpenConns sets the maximum number of open database connections.
    sqlDB.SetMaxOpenConns(conf.Database.MaxOpenConns)
    // SetConnMaxLifetime Sets the maximum time that the connection can be reused.
    sqlDB.SetConnMaxLifetime(time.Hour)

    return sqlDB, nil
}

func Init() error {
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
        dbPath := conf.Database.Path
        os.MkdirAll(conf.WorkDir()+"/data", os.ModePerm)
        fmt.Println("sqlite3 Path:", dbPath)
        db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{SkipDefaultTransaction: true})
        //&gorm.Config{SkipDefaultTransaction: true,}

        // synchronous close
        db.Exec("PRAGMA synchronous = OFF;")
    default:
        log.Errorf("database type not found")
        return errors.New("database type not found")
    }
    if err != nil {
        log.Errorf("init db err,link error:%s", err)
        return err
    }

    sqlDB, err := db.DB()
    // SetMaxIdleConns sets the maximum number of connections in the free connection pool
    sqlDB.SetMaxIdleConns(conf.Database.MaxIdleConns)
    // SetMaxOpenConns sets the maximum number of open database connections.
    sqlDB.SetMaxOpenConns(conf.Database.MaxOpenConns)
    // SetConnMaxLifetime Sets the maximum time that the connection can be reused.
    sqlDB.SetConnMaxLifetime(time.Hour)

    if err != nil {
        log.Errorf("[DB]:%s", err)
        return err
    }

    db.AutoMigrate(&User{})
    db.AutoMigrate(&Domain{})
    db.AutoMigrate(&Mail{})
    db.AutoMigrate(&MailContent{})
    db.AutoMigrate(&Box{})
    db.AutoMigrate(&Class{})
    db.AutoMigrate(&Queue{})

    //创建默认账户
    // var userAdmin User
    // admin := db.First(&userAdmin, "name = ?", "admin")
    // if admin.Error != nil {
    //     db.Create(&User{Name: "admin", Password: "21232f297a57a5a743894a0e4a801fc3", Code: "admin", Token: "21232f297a57a5a743894a0e4a801fc3"})
    // }

    // //退信账户
    // var userPostmaster User
    // postmaster := db.First(&userPostmaster, "name = ?", "postmaster")
    // if postmaster.Error != nil {
    //     db.Create(&User{Name: "postmaster", Password: "21232f297a57a5a743894a0e4a801fc3", Code: "postmaster", Token: "21232f297a57a5a743894a0e4a801fc2"})
    // }

    return nil
}

type Statistic struct {
    Counter struct {
        User      int64
        VaildUser int64
    }
}

func Ping() error {
    sqlDB, _ := db.DB()
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
