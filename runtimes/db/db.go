package db

import (
	"fmt"
	"path/filepath"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/logs"
	"tools/runtimes/mq"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

var DB *gorm.DB
var MQDB *gorm.DB
var MEDIADB *gorm.DB
var TaskDB *gorm.DB

var DBINIT = map[string]**gorm.DB{
	config.DBFILE: &DB,
	"mq.db":       &MQDB,
	"media.db":    &MEDIADB,
	"task.db":     &TaskDB,
}
var MqClient *mq.MQ

func init() {
	for dbfile := range DBINIT {
		*DBINIT[dbfile] = mkdb(dbfile)
	}

	MqClient = mq.New(mq.NewGormStore(MQDB), 3)
	MqClient.Start()
	// dbfile := config.FullPath(config.DBFILE)
	// // dbs, err := gorm.Open(sqlite.Open(dbfile), &gorm.Config{})
	// dbs, err := gorm.Open(sqlite.Dialector{
	// 	DriverName: "sqlite", // 改这里
	// 	DSN:        dbfile + "?_busy_timeout=5000&_journal_mode=WAL",
	// }, &gorm.Config{})
	// if err != nil {
	// 	logs.Error(err.Error())
	// 	panic(err.Error())
	// }

	// ddd, err := dbs.DB()
	// if err != nil {
	// 	logs.Error(err.Error())
	// 	panic("db error")
	// }

	// ddd.SetMaxIdleConns(10)
	// ddd.SetMaxOpenConns(100)
	// ddd.SetConnMaxLifetime(time.Hour)
	// ddd.Exec("PRAGMA journal_mode=WAL;")
	// ddd.Exec("PRAGMA synchronous = NORMAL;")

	// funcs.HiddenDir(dbfile)
	// DB = dbs

	// // 建立用户mq消息队列的数据库
	// mqFile := config.FullPath(filepath.Join(config.SYSROOT, ".mq"))
	// mqdb, err := gorm.Open(sqlite.Dialector{
	// 	DriverName: "sqlite", // 改这里
	// 	DSN:        mqFile + "?_busy_timeout=5000&_journal_mode=WAL",
	// }, &gorm.Config{})
	// if err != nil {
	// 	logs.Error(err.Error())
	// 	panic(err.Error())
	// }

	// mqqq, err := mqdb.DB()
	// if err != nil {
	// 	logs.Error(err.Error())
	// 	panic("db error")
	// }

	// mqqq.SetMaxIdleConns(10)
	// mqqq.SetMaxOpenConns(100)
	// mqqq.SetConnMaxLifetime(time.Hour)
	// mqqq.Exec("PRAGMA journal_mode=WAL;")
	// mqqq.Exec("PRAGMA synchronous = NORMAL;")
	// MQDB = mqdb

	// // 媒体文件的数据库
	// mediaFile := config.FullPath(filepath.Join(config.SYSROOT, ".media"))
	// mediadb, err := gorm.Open(sqlite.Dialector{
	// 	DriverName: "sqlite", // 改这里
	// 	DSN:        mediaFile + "?_busy_timeout=5000&_journal_mode=WAL",
	// }, &gorm.Config{})
	// if err != nil {
	// 	logs.Error(err.Error())
	// 	panic(err.Error())
	// }

	// mediaqqq, err := mediadb.DB()
	// if err != nil {
	// 	logs.Error(err.Error())
	// 	panic("db error")
	// }

	// mediaqqq.SetMaxIdleConns(10)
	// mediaqqq.SetMaxOpenConns(100)
	// mediaqqq.SetConnMaxLifetime(time.Hour)
	// mediaqqq.Exec("PRAGMA journal_mode=WAL;")
	// mediaqqq.Exec("PRAGMA synchronous = NORMAL;")
	// MEDIADB = mediadb

	// TaskDB = mkdb("task.db")
}

func mkdb(dbname string) *gorm.DB {
	dbfile := config.FullPath(filepath.Join(config.SYSROOT, dbname))
	dbhandle, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite", // 改这里
		DSN:        dbfile + "?_busy_timeout=5000&_journal_mode=WAL",
	}, &gorm.Config{})
	if err != nil {
		logs.Error(err.Error())
		fmt.Println(dbfile)
		panic(err.Error())
	}

	dbsql, err := dbhandle.DB()
	if err != nil {
		logs.Error(err.Error())
		panic("db error")
	}

	dbsql.SetMaxIdleConns(10)
	dbsql.SetMaxOpenConns(100)
	dbsql.SetConnMaxLifetime(time.Hour)
	dbsql.Exec("PRAGMA journal_mode=WAL;")
	dbsql.Exec("PRAGMA synchronous = NORMAL;")
	return dbhandle
}
