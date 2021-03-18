package dsn

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	postgresSplitRegexp                = regexp.MustCompile(`\s+`)
	DefaultPostgresDSNParser           = &PostgresDSNParser{}
	_                        DSNParser = (*PostgresDSNParser)(nil)
)

type PostgresDSNParser struct {
}

func (p *PostgresDSNParser) GetDialector(dsn string) gorm.Dialector {
	return postgres.Open(dsn)
}

func (p *PostgresDSNParser) ParseDSN(dsn string) (cfg *DSN, err error) {
	cfg = new(DSN)
	res := postgresSplitRegexp.Split(dsn, -1)
	var host, port string
	for _, kvStr := range res {
		param := strings.SplitN(kvStr, "=", 2)
		if len(param) != 2 {
			continue
		}
		switch param[0] {
		case "user":
			cfg.User = param[1]
		case "password":
			cfg.Password = param[1]
		case "dbname":
			cfg.DBName = param[1]
		case "port":
			port = param[1]
		case "host":
			host = param[1]
		default:
			// lazy init
			if cfg.Params == nil {
				cfg.Params = make(map[string]string)
			}
			cfg.Params[param[0]] = param[1]
		}
	}
	cfg.Addr = fmt.Sprintf("%s:%s", host, port)
	return cfg, nil
}
