package db

import (
	"path/filepath"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/logs"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

var DB *gorm.DB
var MQDB *gorm.DB

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
	ddd.Exec("PRAGMA journal_mode=WAL;")

	funcs.HiddenDir(dbfile)
	DB = dbs

	// 建立用户mq消息队列的数据库
	mqFile := config.FullPath(filepath.Join(config.SYSROOT, ".mq"))
	mqdb, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite", // 改这里
		DSN:        mqFile,
	}, &gorm.Config{})
	if err != nil {
		logs.Error(err.Error())
		panic(err.Error())
	}

	mqqq, err := mqdb.DB()
	if err != nil {
		logs.Error(err.Error())
		panic("db error")
	}

	mqqq.SetMaxIdleConns(10)
	mqqq.SetMaxOpenConns(100)
	mqqq.SetConnMaxLifetime(time.Hour)
	mqqq.Exec("PRAGMA journal_mode=WAL;")
	MQDB = mqdb
}

// func init() {
// 	initDb()
// }
