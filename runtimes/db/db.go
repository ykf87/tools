package db

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"tools/runtimes/config"
	"tools/runtimes/logs"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

var DB *SQLiteWriter
var MQDB *SQLiteWriter
var MEDIADB *SQLiteWriter
var TaskDB *SQLiteWriter
var AppTask *SQLiteWriter
var TaskLogDB *SQLiteWriter

type ListFinder struct {
	Page  int     `json:"page" form:"page"`
	Limit int     `json:"limit" form:"limit"`
	Q     string  `json:"q" form:"q"`
	By    string  `json:"by" form:"by"`
	Scol  string  `json:"scol" form:"scol"`
	Tags  []int64 `json:"tags" form:"tags"`
	Types []int64 `json:"types" form:"types"`
}

var DBINIT = map[string]**SQLiteWriter{
	config.DBFILE: &DB,
	"mq.db":       &MQDB,
	"media.db":    &MEDIADB,
	"task.db":     &TaskDB,
	"apptask.db":  &AppTask,
	"tasklog.db":  &TaskLogDB,
}

func init() {
	for dbfile := range DBINIT {
		*DBINIT[dbfile] = mkdb(dbfile)
	}
}

func mkdb(dbname string) *SQLiteWriter {
	dbfile := config.FullPath(filepath.Join(config.SYSROOT, dbname))
	dns := fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000&_wal_autocheckpoint=1000&_foreign_keys=ON", dbfile)
	dbhandle, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite3", // 改这里
		DSN:        dns,
	}, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
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

	dbsql.SetMaxIdleConns(2)
	dbsql.SetMaxOpenConns(10)
	dbsql.SetConnMaxLifetime(0)
	// dbsql.Exec("PRAGMA journal_mode=WAL;")
	// dbsql.Exec("PRAGMA synchronous = NORMAL;")
	//
	return NewSQLiteWriter(dbhandle, 3000)
	// return dbhandle
}

type BaseModel struct {
}

func (this *BaseModel) Save(model any, db *gorm.DB) error {
	if db == nil {
		db = DB.DB()
	}
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
