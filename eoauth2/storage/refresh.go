package storage

import (
	"context"
	"fmt"

	"github.com/gotomicro/ego-component/egorm"
	"gorm.io/gorm"
)

type Refresh struct {
	Id     int    `gorm:"not null;primary_key;AUTO_INCREMENT" json:"id" form:"id"` // FormID
	Token  string `gorm:"not null" json:"token" form:"token"`                      // token
	Access string `gorm:"not null" json:"access" form:"access"`                    // access
}

func (t *Refresh) TableName() string {
	return "refresh"
}

func RefreshCreate(ctx context.Context, db *gorm.DB, data *Refresh) (err error) {
	if err = db.WithContext(ctx).Create(data).Error; err != nil {
		err = fmt.Errorf("RefreshCreate, err: %w", err)
		return
	}
	return
}

// RefreshDeleteX Delete的扩展方法，根据Cond删除一条或多条记录。如果有delete_time则软删除，否则硬删除。
func RefreshDeleteX(ctx context.Context, db *gorm.DB, conds egorm.Conds) (err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = db.WithContext(ctx).Table("refresh").Where(sql, binds...).Delete(&Refresh{}).Error; err != nil {
		err = fmt.Errorf("RefreshDeleteX, err: %w", err)
		return
	}
	return
}

// RefreshInfoX Info的扩展方法，根据Cond查询单条记录
func RefreshInfoX(ctx context.Context, db *gorm.DB, conds egorm.Conds) (resp Refresh, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = db.WithContext(ctx).Table("refresh").Where(sql, binds...).First(&resp).Error; err != nil {
		err = fmt.Errorf("RefreshInfoX, err: %w", err)
		return
	}
	return
}
