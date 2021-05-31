package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/gotomicro/ego-component/egorm"
	"gorm.io/gorm"
)

type Access struct {
	Id           int    `gorm:"not null;primary_key;AUTO_INCREMENT" json:"id" form:"id"` // FormID
	Client       string `gorm:"not null" json:"client" form:"client"`                    // client
	Authorize    string `gorm:"not null" json:"authorize" form:"authorize"`              // authorize
	Previous     string `gorm:"not null" json:"previous" form:"previous"`                // previous
	AccessToken  string `gorm:"not null" json:"accessToken" form:"accessToken"`          // access_token
	RefreshToken string `gorm:"not null" json:"refreshToken" form:"refreshToken"`        // refresh_token
	ExpiresIn    int    `gorm:"not null" json:"expiresIn" form:"expiresIn"`              // expires_in
	Scope        string `gorm:"not null" json:"scope" form:"scope"`                      // scope
	RedirectUri  string `gorm:"not null" json:"redirectUri" form:"redirectUri"`          // redirect_uri
	Extra        string `gorm:"not null;type:longtext" json:"extra" form:"extra"`        // extra
	Ctime        int64  `gorm:"not null" json:"ctime" form:"ctime"`                      // 创建时间
}

func (t *Access) TableName() string {
	return "access"
}

// AccessCreate insert a new Access into database and returns
// last inserted Id on success.
func AccessCreate(ctx context.Context, db *gorm.DB, data *Access) (err error) {
	data.Ctime = time.Now().Unix()
	if err = db.WithContext(ctx).Create(data).Error; err != nil {
		err = fmt.Errorf("AccessCreate, err: %w", err)
		return
	}
	return
}

// AccessDeleteX Delete的扩展方法，根据Cond删除一条或多条记录。如果有delete_time则软删除，否则硬删除。
func AccessDeleteX(ctx context.Context, db *gorm.DB, conds egorm.Conds) (err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = db.WithContext(ctx).Table("access").Where(sql, binds...).Delete(&Access{}).Error; err != nil {
		err = fmt.Errorf("AccessDeleteX, err: %w", err)
		return
	}
	return
}

// AccessInfoX Info的扩展方法，根据Cond查询单条记录
func AccessInfoX(ctx context.Context, db *gorm.DB, conds egorm.Conds) (resp Access, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = db.WithContext(ctx).Table("access").Where(sql, binds...).First(&resp).Error; err != nil {
		err = fmt.Errorf("AccessInfoX, err: %w", err)
		return
	}
	return
}
