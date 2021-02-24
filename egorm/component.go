package egorm

import (
	"errors"
	"github.com/gotomicro/ego/core/elog"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const PackageName = "component.egorm"

var (
	errSlowCommand = errors.New("mysql slow command")
	// IsRecordNotFoundError ...
	IsRecordNotFoundError = gorm.IsRecordNotFoundError
	// ErrRecordNotFound returns a "record not found error". Occurs only when attempting to query the database with a struct; querying with a slice won't return this error
	ErrRecordNotFound = gorm.ErrRecordNotFound
	// ErrInvalidSQL occurs when you attempt a query with invalid SQL
	ErrInvalidSQL = gorm.ErrInvalidSQL
	// ErrInvalidTransaction occurs when you are trying to `Commit` or `Rollback`
	ErrInvalidTransaction = gorm.ErrInvalidTransaction
	// ErrCantStartTransaction can't start transaction when you are trying to start one with `Begin`
	ErrCantStartTransaction = gorm.ErrCantStartTransaction
	// ErrUnaddressable unaddressable value
	ErrUnaddressable = gorm.ErrUnaddressable
)

// SQLCommon ...
type (
	// SQLCommon alias of gorm.SQLCommon
	SQLCommon = gorm.SQLCommon
	// Callback alias of gorm.Callback
	Callback = gorm.Callback
	// CallbackProcessor alias of gorm.CallbackProcessor
	CallbackProcessor = gorm.CallbackProcessor
	// Dialect alias of gorm.Dialect
	Dialect = gorm.Dialect
	// Scope ...
	Scope = gorm.Scope
	// Model ...
	Model = gorm.Model
	// ModelStruct ...
	ModelStruct = gorm.ModelStruct
	// Field ...
	Field = gorm.Field
	// FieldStruct ...
	StructField = gorm.StructField
	// RowQueryResult ...
	RowQueryResult = gorm.RowQueryResult
	// RowsQueryResult ...
	RowsQueryResult = gorm.RowsQueryResult
	// Association ...
	Association = gorm.Association
	// Errors ...
	Errors = gorm.Errors
	// logger ...
	Logger = gorm.Logger
)

// Component ...
type Component = gorm.DB

// newComponent ...
func newComponent(compName string, config *config, logger *elog.Component) (*Component, error) {
	gormDB, err := gorm.Open(config.Dialect, config.DSN)
	if err != nil {
		return nil, err
	}

	if config.RawDebug {
		gormDB = gormDB.Debug()
	}

	// 设置默认连接配置
	gormDB.DB().SetMaxIdleConns(config.MaxIdleConns)
	gormDB.DB().SetMaxOpenConns(config.MaxOpenConns)

	if config.ConnMaxLifetime != 0 {
		gormDB.DB().SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	replace := func(processor func() *gorm.CallbackProcessor, callbackName string, interceptors ...Interceptor) {
		old := processor().Get(callbackName)
		var handler = old
		for _, interceptor := range interceptors {
			handler = interceptor(compName, config.dsnCfg, callbackName, config, logger)(handler)
		}
		processor().Replace(callbackName, handler)
	}

	replace(
		gormDB.Callback().Delete,
		"gorm:delete",
		config.interceptors...,
	)
	replace(
		gormDB.Callback().Update,
		"gorm:update",
		config.interceptors...,
	)
	replace(
		gormDB.Callback().Create,
		"gorm:create",
		config.interceptors...,
	)
	replace(
		gormDB.Callback().Query,
		"gorm:query",
		config.interceptors...,
	)
	replace(
		gormDB.Callback().RowQuery,
		"gorm:row_query",
		config.interceptors...,
	)

	return gormDB, nil
}
