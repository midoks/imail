package db

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/pkg/errors"
    "gorm.io/driver/mysql"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"

    "github.com/midoks/imail/internal/conf"
    "github.com/midoks/imail/internal/log"
)

var (
    db  *gorm.DB
    err error
)

var Tables = []interface{}{
    new(User), new(Domain), new(Mail),
    new(MailLog), new(MailContent), new(Box),
    new(Class), new(Queue),
}

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
        os.MkdirAll(conf.Web.AppDataPath, os.ModePerm)
        dbPath := conf.Database.Path
        if strings.EqualFold(conf.Database.Path, "data/imail.db3") {
            dbPath = filepath.Dir(conf.Web.AppDataPath) + "/" + conf.Database.Path
        }
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

    // db.AutoMigrate(&User{})
    // db.AutoMigrate(&Domain{})
    // db.AutoMigrate(&Mail{})
    // db.AutoMigrate(&MailLog{})
    // db.AutoMigrate(&MailContent{})
    // db.AutoMigrate(&Box{})
    // db.AutoMigrate(&Class{})
    // db.AutoMigrate(&Queue{})

    for _, table := range Tables {
        if db.Migrator().HasTable(table) {
            continue
        }

        name := strings.TrimPrefix(fmt.Sprintf("%T", table), "*db.")
        err = db.Migrator().AutoMigrate(table)
        if err != nil {
            return errors.Wrapf(err, "auto migrate %q", name)
        }
    }

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

func TablePrefix(tn string) string {
    return fmt.Sprintf("%s%s", conf.Database.Prefix, tn)
}

func CheckDb() bool {
    if db != nil {
        return true
    }
    return false
}
