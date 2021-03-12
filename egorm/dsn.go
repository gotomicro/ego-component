package egorm

// DSN ...
type DSN struct {
	User     string            // Username
	Password string            // Password (requires User)
	Net      string            // Network type
	Addr     string            // Network address (requires Net)
	DBName   string            // Database name
	Params   map[string]string // Connection parameters
}

const (
	DialectMysql    = "mysql"
	DialectPostgres = "postgres"
)

func ParseDSN(dialect string, dsn string) (cfg *DSN, err error) {
	switch dialect {
	case DialectMysql:
		cfg, err = MysqlDSNParser(dsn).ParseDSN()
	case DialectPostgres:
		cfg, err = PostgresDsnParser(dsn).ParseDSN()
	default:
		cfg = new(DSN)
	}
	return
}
