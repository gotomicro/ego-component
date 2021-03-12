package egorm

import (
	"fmt"
	"regexp"
	"strings"
)

var postgresSplitRegexp = regexp.MustCompile(`\s+`)

// example :  user=gorm password=gorm dbname=gorm port=9920 host=localhost sslmode=disable
type PostgresDsnParser string

func (p PostgresDsnParser) ParseDSN() (cfg *DSN, err error) {
	cfg = new(DSN)
	res := postgresSplitRegexp.Split(string(p), -1)
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
