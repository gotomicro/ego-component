package conf

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gotomicro/ego-component/egorm"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/econf/manager"
	"github.com/gotomicro/ego/core/elog"
)

// dataSource file provider.
type dataSource struct {
	changed          chan struct{}
	logger           *elog.Component
	db               *egorm.Component
	table            string // 表名
	keyValue         string // 主键字段值
	keyColumnName    string // 主键字段名
	configColumnName string // value字段名
}

func init() {
	manager.Register("mysql", &dataSource{})
}

// Parse 解析配置
// 全部配置 mysql://ip:port/?username=xxx&password=xxx&database=xxx&table=xxx&keyValue=xxx&keyColumnName=xxx&configColumnName=xxx&configType=xxx&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=1s&readTimeout=3s&writeTimeout=3s
// 简单配置 mysql://ip:port/?username=xxx&password=xxx&database=xxx&table=xxx&keyValue=xxx&keyColumnName=xxx&configColumnName=xxx&configType=xxx
func (fp *dataSource) Parse(path string, watch bool) econf.ConfigType {
	fp.logger = elog.EgoLogger.With(elog.FieldComponent(econf.PackageName))

	urlInfo, err := url.Parse(path)
	if err != nil {
		fp.logger.Panic("new datasource", elog.FieldErr(err))
		return ""
	}
	keyValue := urlInfo.Query().Get("keyValue")
	keyColumnName := urlInfo.Query().Get("keyColumnName")
	configColumnName := urlInfo.Query().Get("configColumnName")
	configType := urlInfo.Query().Get("configType")
	database := urlInfo.Query().Get("database")
	table := urlInfo.Query().Get("table")
	charset := urlInfo.Query().Get("charset")
	collation := urlInfo.Query().Get("collation")
	parseTime := urlInfo.Query().Get("parseTime")
	loc := urlInfo.Query().Get("loc")
	timeout := urlInfo.Query().Get("timeout")
	readTimeout := urlInfo.Query().Get("readTimeout")
	writeTimeout := urlInfo.Query().Get("writeTimeout")

	if keyColumnName == "" {
		fp.logger.Panic("mysql keyColumnName is empty")
	}
	fp.keyColumnName = keyColumnName

	if configColumnName == "" {
		fp.logger.Panic("mysql configColumnName is empty")
	}
	fp.configColumnName = configColumnName

	if keyValue == "" {
		fp.logger.Panic("mysql keyValue is empty")
	}
	fp.keyValue = keyValue
	if table == "" {
		fp.logger.Panic("table is empty")
	}
	fp.table = table

	if configType == "" {
		fp.logger.Panic("config type is empty")
	}

	if database == "" {
		fp.logger.Panic("database is empty")
	}

	if charset == "" {
		charset = "utf8mb4"
	}
	if collation == "" {
		collation = "utf8mb4_general_ci"
	}
	if parseTime == "" {
		parseTime = "True"
	}
	if loc == "" {
		loc = "Local"
	}
	if timeout == "" {
		timeout = "3s"
	}

	if readTimeout == "" {
		readTimeout = "3s"
	}
	if writeTimeout == "" {
		writeTimeout = "3s"
	}

	query := url.Values{}
	query.Set("charset", charset)
	query.Set("collation", collation)
	query.Set("parseTime", parseTime)
	query.Set("loc", loc)
	query.Set("timeout", timeout)
	query.Set("readTimeout", readTimeout)
	query.Set("writeTimeout", writeTimeout)
	mysqlDsn := urlInfo.Query().Get("username") + ":" + urlInfo.Query().Get("password") + "@tcp(" + urlInfo.Host + ")/" + database + "?" + query.Encode()

	fp.db = egorm.DefaultContainer().Build(
		egorm.WithDSN(mysqlDsn),
	)
	return econf.ConfigType(configType)
}

// ReadConfig ...
func (fp *dataSource) ReadConfig() (content []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := map[string]interface{}{}

	err = fp.db.WithContext(ctx).Table(fp.table).Where(fp.keyColumnName+"=?", fp.keyValue).Take(&result).Error
	if err != nil {
		return nil, err
	}

	info, ok := result[fp.configColumnName]
	if !ok {
		return nil, fmt.Errorf("config not found")
	}
	configStr, ok2 := info.(string)
	if !ok2 {
		return nil, fmt.Errorf("config not string")
	}
	return []byte(configStr), nil
}

// Close ...
func (fp *dataSource) Close() error {
	return nil
}

// IsConfigChanged ...
func (fp *dataSource) IsConfigChanged() <-chan struct{} {
	return fp.changed
}

// Watch file and automate update.
func (fp *dataSource) watch() {

}
