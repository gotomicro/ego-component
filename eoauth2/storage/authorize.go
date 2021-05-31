package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/gotomicro/ego-component/egorm"
	"gorm.io/gorm"
)

type Authorize struct {
	Id          int    `gorm:"not null;primary_key;AUTO_INCREMENT" json:"id" form:"id"` // FormID
	Client      string `gorm:"not null" json:"client" form:"client"`                    // 客户端
	Code        string `gorm:"not null" json:"code" form:"code"`                        // 状态码
	ExpiresIn   int32  `gorm:"not null" json:"expiresIn" form:"expiresIn"`              // 过期时间
	Scope       string `gorm:"not null" json:"scope" form:"scope"`                      // 范围
	RedirectUri string `gorm:"not null" json:"redirectUri" form:"redirectUri"`          // 跳转地址
	State       string `gorm:"not null" json:"state" form:"state"`                      // 状态
	Extra       string `gorm:"not null;type:longtext" json:"extra" form:"extra"`        // 额外信息
	Ctime       int64  `gorm:"not null" json:"ctime" form:"ctime"`                      // 创建时间
}

func (t *Authorize) TableName() string {
	return "authorize"
}

// AuthorizeCreate insert a new Authorize into database and returns
// last inserted Id on success.
func AuthorizeCreate(ctx context.Context, db *gorm.DB, data *Authorize) (err error) {
	data.Ctime = time.Now().Unix()
	if err = db.WithContext(ctx).Create(data).Error; err != nil {
		err = fmt.Errorf("AuthorizeCreate, err: %w", err)
		return
	}
	return
}

func AuthorizeDeleteX(ctx context.Context, db *gorm.DB, conds egorm.Conds) (err error) {
	sql, binds := egorm.BuildQuery(conds)

	if err = db.WithContext(ctx).Table("authorize").Where(sql, binds...).Delete(&Authorize{}).Error; err != nil {
		err = fmt.Errorf("AuthorizeDeleteX, err: %w", err)
		return
	}

	return
}

// AuthorizeInfoX Info的扩展方法，根据Cond查询单条记录
func AuthorizeInfoX(ctx context.Context, db *egorm.Component, conds egorm.Conds) (resp Authorize, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = db.WithContext(ctx).Table("authorize").Where(sql, binds...).First(&resp).Error; err != nil {
		err = fmt.Errorf("AuthorizeInfoX, err: %w", err)
		return
	}
	return
}
