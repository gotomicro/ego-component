package egorm

import (
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"

	"github.com/gotomicro/ego-component/egorm/dsn"
)

// Container ...
type Container struct {
	config    *config
	name      string
	logger    *elog.Component
	dsnParser dsn.DSNParser
}

// DefaultContainer ...
func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

// Load ...
func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}

	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

func (c *Container) setDSNParserIfNotExists(dialect string) error {
	if c.dsnParser != nil {
		return nil
	}
	switch dialect {
	case dialectMysql:
		c.dsnParser = dsn.DefaultMysqlDSNParser
	case dialectPostgres:
		c.dsnParser = dsn.DefaultPostgresDSNParser
	default:
		return errSupportDialect
	}
	return nil
}

// Build 构建组件
func (c *Container) Build(options ...Option) *Component {
	if c.config.Debug {
		options = append(options, WithInterceptor(debugInterceptor))
	}

	if c.config.EnableTraceInterceptor {
		options = append(options, WithInterceptor(traceInterceptor))
	}

	if c.config.EnableMetricInterceptor {
		options = append(options, WithInterceptor(metricInterceptor))
	}

	for _, option := range options {
		option(c)
	}

	var err error
	// todo 设置补齐超时时间
	// timeout 1s
	// readTimeout 5s
	// writeTimeout 5s
	err = c.setDSNParserIfNotExists(c.config.Dialect)
	if err != nil {
		c.logger.Panic("setDSNParserIfNotExists err", elog.String("dialect", c.config.Dialect), elog.FieldErr(err))
	}
	c.config.dsnCfg, err = c.dsnParser.ParseDSN(c.config.DSN)

	if err == nil {
		c.logger.Info("start db", elog.FieldAddr(c.config.dsnCfg.Addr), elog.FieldName(c.config.dsnCfg.DBName))
	} else {
		c.logger.Panic("start db", elog.FieldErr(err))
	}

	c.logger = c.logger.With(elog.FieldAddr(c.config.dsnCfg.Addr))

	component, err := newComponent(c.name, c.dsnParser, c.config, c.logger)
	if err != nil {
		if c.config.OnFail == "panic" {
			c.logger.Panic("open db", elog.FieldErrKind("register err"), elog.FieldErr(err), elog.FieldAddr(c.config.dsnCfg.Addr), elog.FieldValueAny(c.config))
		} else {
			emetric.ClientHandleCounter.Inc(emetric.TypeGorm, c.name, c.name+".ping", c.config.dsnCfg.Addr, "open err")
			c.logger.Error("open db", elog.FieldErrKind("register err"), elog.FieldErr(err), elog.FieldAddr(c.config.dsnCfg.Addr), elog.FieldValueAny(c.config))
			return component
		}
	}

	sqlDB, err := component.DB()
	if err != nil {
		c.logger.Panic("ping db", elog.FieldErrKind("register err"), elog.FieldErr(err), elog.FieldValueAny(c.config))
	}
	if err := sqlDB.Ping(); err != nil {
		c.logger.Panic("ping db", elog.FieldErrKind("register err"), elog.FieldErr(err), elog.FieldValueAny(c.config))
	}

	// store db
	instances.Store(c.name, component)
	return component
}
