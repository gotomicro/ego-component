package egorm

import (
	"context"
	"errors"

	"github.com/gotomicro/ego/core/elog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/gotomicro/ego-component/egorm/dsn"
)

// PackageName ...
const PackageName = "component.egorm"

var (
	errSlowCommand = errors.New("mysql slow command")
	// ErrRecordNotFound returns a "record not found error". Occurs only when attempting to query the database with a struct; querying with a slice won't return this error
	ErrRecordNotFound = gorm.ErrRecordNotFound
	// ErrInvalidTransaction occurs when you are trying to `Commit` or `Rollback`
	ErrInvalidTransaction = gorm.ErrInvalidTransaction
)

type (
	// DB ...
	DB gorm.DB
	// Dialector ...
	Dialector = gorm.Dialector
	// Model ...
	Model = gorm.Model
	// Field ...
	Field = schema.Field
	// Association ...
	Association = gorm.Association
	// NamingStrategy ...
	NamingStrategy = schema.NamingStrategy
	// Logger ...
	Logger = logger.Interface
)

// Component ...
type Component = gorm.DB

// WithContext ...
func WithContext(ctx context.Context, db *Component) *Component {
	db.Statement.Context = ctx
	return db
}

// newComponent ...
func newComponent(compName string, dsnParser dsn.DSNParser, config *config, elogger *elog.Component) (*Component, error) {
	db, err := gorm.Open(dsnParser.GetDialector(config.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if config.RawDebug {
		db = db.Debug()
	}

	gormDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置默认连接配置
	gormDB.SetMaxIdleConns(config.MaxIdleConns)
	gormDB.SetMaxOpenConns(config.MaxOpenConns)

	if config.ConnMaxLifetime != 0 {
		gormDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	replace := func(processor Processor, callbackName string, interceptors ...Interceptor) {
		handler := processor.Get(callbackName)
		for _, interceptor := range config.interceptors {
			handler = interceptor(compName, config.dsnCfg, callbackName, config, elogger)(handler)
		}

		processor.Replace(callbackName, handler)
	}

	replace(db.Callback().Create(), "gorm:create", config.interceptors...)
	replace(db.Callback().Update(), "gorm:update", config.interceptors...)
	replace(db.Callback().Delete(), "gorm:delete", config.interceptors...)
	replace(db.Callback().Query(), "gorm:query", config.interceptors...)
	// replace(db.Callback().Row(), "gorm:row", config.interceptors...)
	replace(db.Callback().Raw(), "gorm:raw", config.interceptors...)

	return db, nil
}
