package dsn

import (
	"fmt"
	"github.com/gotomicro/ego-component/egorm/manager"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"strings"
)

var (
	_ manager.DSNParser = (*SqlServerDSNParser)(nil)
)

type SqlServerDSNParser struct {
}

func init() {
	manager.Register(&SqlServerDSNParser{})
}

func (p *SqlServerDSNParser) Scheme() string {
	return "mssql"
}

func (p *SqlServerDSNParser) GetDialector(dsn string) gorm.Dialector {
	return sqlserver.Open(dsn)
}

func (p *SqlServerDSNParser) ParseDSN(dsn string) (cfg *manager.DSN, err error) {
	cfg = new(manager.DSN)
	prefixStr := dsn[0:12]
	var host, port string
	if prefixStr == "sqlserver://" {
		host, port = parseNewDNS(cfg, dsn[12:])
	} else {
		res := strings.Split(dsn, ";")
		for _, kvStr := range res {
			param := strings.SplitN(kvStr, "=", 2)
			if len(param) != 2 {
				continue
			}
			switch param[0] {
			case "user id":
				cfg.User = param[1]
			case "password":
				cfg.Password = param[1]
			case "database":
				cfg.DBName = param[1]
			case "port":
				port = param[1]
			case "server":
				host = param[1]
			default:
				// lazy init
				if cfg.Params == nil {
					cfg.Params = make(map[string]string)
				}
				cfg.Params[param[0]] = param[1]
			}
		}
	}
	cfg.Addr = fmt.Sprintf("%s:%s", host, port)
	return cfg, nil
}

func parseNewDNS(cfg *manager.DSN, dsn string) (string, string) {
	param := strings.SplitN(dsn, "?", 2)
	if len(param) != 2 {
		return "", ""
	}
	//get db info
	paramDB := strings.SplitN(param[1], "=", 2)
	if paramDB[0] == "database" {
		cfg.DBName = paramDB[1]
	}

	//get user&host
	paramUserAndHost := strings.SplitN(param[0], "@", 2)

	paramUser := strings.SplitN(paramUserAndHost[0], ":", 2)
	if len(paramUser) == 2 {
		cfg.User = paramUser[0]
		cfg.Password = paramUser[1]
	}
	paramAddr := strings.SplitN(paramUserAndHost[1], ":", 2)
	if len(paramAddr) == 2 {
		cfg.Addr = fmt.Sprintf("%s:%s", paramAddr[0], paramAddr[1])
		return paramAddr[0], paramAddr[1]
	}
	return "", ""
}
