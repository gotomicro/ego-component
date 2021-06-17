package dao

import (
	"context"
	"fmt"

	"github.com/gotomicro/ego-component/egorm"
	"gorm.io/gorm"
)

type Expires struct {
	Id        int    `gorm:"not null;primary_key;AUTO_INCREMENT" json:"id" form:"id"` // 客户端
	Token     string `gorm:"not null" json:"token" form:"token"`                      // token
	ExpiresAt int64  `gorm:"not null" json:"expiresAt" form:"expiresAt"`              // 过期时间

}

func (t *Expires) TableName() string {
	return "expires"
}

// ExpiresCreate insert a new Expires into database and returns
// last inserted Id on success.
func ExpiresCreate(ctx context.Context, db *gorm.DB, data *Expires) (err error) {
	if err = db.WithContext(ctx).Create(data).Error; err != nil {
		err = fmt.Errorf("ExpiresCreate, err: %w", err)
		return
	}
	return
}

// ExpiresDeleteX Delete的扩展方法，根据Cond删除一条或多条记录。如果有delete_time则软删除，否则硬删除。
func ExpiresDeleteX(ctx context.Context, db *gorm.DB, conds egorm.Conds) (err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = db.WithContext(ctx).Table("expires").Where(sql, binds...).Delete(&Expires{}).Error; err != nil {
		err = fmt.Errorf("ExpiresDeleteX, err: %w", err)
		return
	}
	return
}
