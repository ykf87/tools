package db

import (
	"time"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/logs"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

var DB *gorm.DB

func init() {
	dbfile := config.FullPath(config.DBFILE)
	// dbs, err := gorm.Open(sqlite.Open(dbfile), &gorm.Config{})
	dbs, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite", // 改这里
		DSN:        dbfile,
	}, &gorm.Config{})
	if err != nil {
		logs.Error(err.Error())
		panic(err.Error())
	}

	ddd, err := dbs.DB()
	if err != nil {
		logs.Error(err.Error())
		panic("db error")
	}

	ddd.SetMaxIdleConns(10)
	ddd.SetMaxOpenConns(100)
	ddd.SetConnMaxLifetime(time.Hour)

	funcs.HiddenDir(dbfile)

	DB = dbs
}

// func init() {
// 	initDb()
// }
