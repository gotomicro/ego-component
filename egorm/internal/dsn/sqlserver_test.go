package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqlServerDsnParser_ParseDSN(t *testing.T) {
	dsn := "server=127.0.0.1;user id=user;password=pwd;database=dbName;port=1433;encrypt=disable"
	//dsn := "sqlserver://user:pwd@127.0.0.1:1433?database=dbName"
	sqlServerDSNParser := SqlServerDSNParser{}
	cfg, err := sqlServerDSNParser.ParseDSN(dsn)
	assert.NoError(t, err)
	assert.Equal(t, "user", cfg.User)
	assert.Equal(t, "pwd", cfg.Password)
	assert.Equal(t, "dbName", cfg.DBName)
	assert.Equal(t, "127.0.0.1:1433", cfg.Addr)
	assert.Equal(t, "disable", cfg.Params["encrypt"])
}
