package dsn

import (
	"github.com/gotomicro/gorm-driver-dm"
	"gorm.io/gorm"
)

var (
	DefaultDmDSNParser           = &DmDSNParser{}
	_                  DSNParser = (*DmDSNParser)(nil)
)

type DmDSNParser struct {
}

func (m *DmDSNParser) GetDialector(dsn string) gorm.Dialector {
	return dm.Open(dsn)
}

func (m *DmDSNParser) ParseDSN(dsn string) (cfg *DSN, err error) {
	// New config with some default values
	cfg = new(DSN)

	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	// Find the last '/' (since the password or the net addr might contain a '/')
	//foundSlash := false
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '/' {
			//foundSlash = true
			var j int

			// left part is empty if i <= 0
			if i > 0 {
				// [username[:password]@][protocol[(address)]]
				// Find the last '@' in dsn[:i]
				for j = i; j >= 0; j-- {
					if dsn[j] == '@' {
						parseUsernamePassword(cfg, dsn[:j])
						break
					}
				}

				// [protocol[(address)]]
				// Find the first '(' in dsn[j+1:i]
				if err = parseAddrNet(cfg, dsn[j:i]); err != nil {
					return
				}
			}

			// dbname[?param1=value1&...&paramN=valueN]
			// Find the first '?' in dsn[i+1:]
			for j = i + 1; j < len(dsn); j++ {
				if dsn[j] == '?' {
					if err = parseDSNParams(cfg, dsn[j+1:]); err != nil {
						return
					}
					break
				}
			}
			//cfg.DBName = dsn[i+1 : j]

			break
		}
	}
	//if !foundSlash && len(dsn) > 0 {
	//	return nil, errInvalidDSNNoSlash
	//}
	return
}
