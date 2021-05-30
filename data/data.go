import (
  "gorm.io/driver/mysql"
  "gorm.io/gorm"
)

func init() {
  db, err := gorm.Open(mysql.New(mysql.Config{
    DriverName: "my_mysql_driver",
    DSN:        "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local", // Data Source Name，参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name
  }), &gorm.Config{})
}

