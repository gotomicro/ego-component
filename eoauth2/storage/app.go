package storage

import (
	"context"
	"fmt"

	"github.com/gotomicro/ego-component/egorm"
)

type App struct {
	Aid         int    `gorm:"not null;primary_key;AUTO_INCREMENT" json:"aid" form:"aid"` // 应用id
	ClientId    string `gorm:"not null" json:"clientId" form:"clientId"`                  // 客户端
	Name        string `gorm:"not null" json:"name" form:"name"`                          // 名称
	Secret      string `gorm:"not null" json:"secret" form:"secret"`                      // 秘钥
	RedirectUri string `gorm:"not null" json:"redirectUri" form:"redirectUri"`            // 跳转地址
	Url         string `gorm:"not null" json:"url" form:"url"`                            // 访问地址
	Extra       string `gorm:"not null;type:longtext" json:"extra" form:"extra"`          // 额外信息
	CntCall     int    `gorm:"not null" json:"cntCall" form:"cntCall"`                    // 调用次数
	State       int    `gorm:"not null" json:"state" form:"state"`                        // 状态
	Ctime       int64  `gorm:"not null" json:"ctime" form:"ctime"`                        // 创建时间
	Utime       int64  `gorm:"not null" json:"utime" form:"utime"`                        // 更新时间
	Dtime       int64  `gorm:"not null" json:"dtime" form:"dtime"`                        // 删除时间

}

func (t *App) TableName() string {
	return "app"
}

// AppInfoX Info的扩展方法，根据Cond查询单条记录
func AppInfoX(ctx context.Context, db *egorm.Component, conds egorm.Conds) (resp App, err error) {
	conds["dtime"] = 0
	sql, binds := egorm.BuildQuery(conds)
	if err = db.WithContext(ctx).Table("app").Where(sql, binds...).First(&resp).Error; err != nil {
		err = fmt.Errorf("AccessDeleteX, err: %w", err)
		return
	}
	return
}
