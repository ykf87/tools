package db

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
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
var AppTask *gorm.DB

type ListFinder struct {
	Page  int     `json:"page" form:"page"`
	Limit int     `json:"limit" form:"limit"`
	Q     string  `json:"q" form:"q"`
	By    string  `json:"by" form:"by"`
	Scol  string  `json:"scol" form:"scol"`
	Tags  []int64 `json:"tags" form:"tags"`
	Types []int64 `json:"types" form:"types"`
}

var DBINIT = map[string]**gorm.DB{
	config.DBFILE: &DB,
	"mq.db":       &MQDB,
	"media.db":    &MEDIADB,
	"task.db":     &TaskDB,
	"apptask.db":  &AppTask,
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

type BaseModel struct {
}

func (this *BaseModel) Save(model any, db *gorm.DB) error {
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return errors.New("model must be a non-nil pointer")
	}

	v = v.Elem()
	t := v.Type()

	where := map[string]any{}
	updates := map[string]any{}

	hasPrimaryKey := false
	primaryKeyReady := true // 是否“所有主键都有值”

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		if ft.Anonymous || !fv.CanInterface() {
			continue
		}

		gormTag := ft.Tag.Get("gorm")
		updateTag := ft.Tag.Get("update")

		column := parseColumnName(gormTag, ft.Name)

		// ---- 主键处理 ----
		if strings.Contains(gormTag, "primaryKey") {
			hasPrimaryKey = true

			if isZeroValue(fv) {
				primaryKeyReady = false
			} else {
				where[column] = fv.Interface()
			}
			continue
		}

		// ---- 更新字段规则 ----
		if updateTag == "false" {
			continue
		}

		updates[column] = fv.Interface()
	}

	// 没定义主键 or 主键未就绪 => Create
	if !hasPrimaryKey || !primaryKeyReady {
		return db.Create(model).Error
	}

	// 有主键但无可更新字段
	if len(updates) == 0 {
		return nil
	}

	return db.Model(model).
		Where(where).
		Updates(updates).
		Error
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		zero := reflect.Zero(v.Type())
		return reflect.DeepEqual(v.Interface(), zero.Interface())
	default:
		return v.IsZero()
	}
}
func parseColumnName(gormTag, fieldName string) string {
	parts := strings.SplitSeq(gormTag, ";")
	for p := range parts {
		if col, ok := strings.CutPrefix(p, "column:"); ok {
			return col
		}
	}
	return fieldName
}
