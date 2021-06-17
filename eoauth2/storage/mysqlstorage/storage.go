package mysqlstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotomicro/ego-component/egorm"
	"github.com/gotomicro/ego-component/eoauth2/server"
	"github.com/gotomicro/ego-component/eoauth2/storage/dao"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/etrace"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type storage struct {
	db     *egorm.Component
	logger *elog.Component
}

// NewStorage returns a new mysql storage instance.
func NewStorage(db *gorm.DB, logger *elog.Component) *storage {
	return &storage{
		db,
		logger,
	}
}

// Clone the storage if needed. For example, using mgo, you can clone the session with session.Clone
// to avoid concurrent access problems.
// This is to avoid cloning the connection at each method access.
// Can return itself if not a problem.
func (s *storage) Clone() server.Storage {
	return s
}

// Close the resources the storage potentially holds (using Clone for example)
func (s *storage) Close() {
}

// GetClient loads the client by id
func (s *storage) GetClient(ctx context.Context, clientId string) (client server.Client, err error) {
	span, ctx := etrace.StartSpanFromContext(ctx, "mysqlStorage.GetClient")
	defer span.Finish()

	app, err := dao.AppInfoX(ctx, s.db, egorm.Conds{"client_id": clientId})
	if err != nil {
		err = fmt.Errorf("mysql storage get client error,err: %w", err)
		return
	}
	c := server.DefaultClient{
		Id:          app.ClientId,
		Secret:      app.Secret,
		RedirectUri: app.RedirectUri,
		UserData:    app.Extra,
	}
	return &c, nil
}

// SaveAuthorize saves authorize data.
func (s *storage) SaveAuthorize(ctx context.Context, data *server.AuthorizeData) (err error) {
	obj := dao.Authorize{
		Client:      data.Client.GetId(),
		Code:        data.Code,
		ExpiresIn:   data.ExpiresIn,
		Scope:       data.Scope,
		RedirectUri: data.RedirectUri,
		State:       data.State,
		Ctime:       data.CreatedAt.Unix(),
		Extra:       cast.ToString(data.UserData),
	}
	span, ctx := etrace.StartSpanFromContext(
		ctx,
		"mysqlStorage.SaveAuthorize",
	)
	defer span.Finish()

	tx := s.db.Begin()
	err = dao.AuthorizeCreate(ctx, tx, &obj)
	if err != nil {
		tx.Rollback()
		return
	}

	err = s.AddExpireAtData(ctx, tx, data.Code, data.ExpireAt())
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}

// LoadAuthorize looks up AuthorizeData by a code.
// Client information MUST be loaded together.
// Optionally can return error if expired.
func (s *storage) LoadAuthorize(ctx context.Context, code string) (*server.AuthorizeData, error) {
	span, ctx := etrace.StartSpanFromContext(
		ctx,
		"mysqlStorage.LoadAuthorize",
	)
	defer span.Finish()

	var data server.AuthorizeData

	info, err := dao.AuthorizeInfoX(ctx, s.db, egorm.Conds{"code": code})
	if err != nil {
		return nil, err
	}

	data = server.AuthorizeData{
		Code:        info.Code,
		ExpiresIn:   info.ExpiresIn,
		Scope:       info.Scope,
		RedirectUri: info.RedirectUri,
		State:       info.State,
		CreatedAt:   time.Unix(info.Ctime, 0),
		UserData:    info.Extra,
	}
	c, err := s.GetClient(ctx, info.Client)
	if err != nil {
		return nil, err
	}

	if data.ExpireAt().Before(time.Now()) {
		return nil, fmt.Errorf("Token expired at %s.", data.ExpireAt().String())
	}

	data.Client = c
	return &data, nil
}

// RemoveAuthorize revokes or deletes the authorization code.
func (s *storage) RemoveAuthorize(ctx context.Context, code string) (err error) {
	span, ctx := etrace.StartSpanFromContext(
		ctx,
		"mysqlStorage.RemoveAuthorize",
	)
	defer span.Finish()

	err = dao.AuthorizeDeleteX(ctx, s.db, egorm.Conds{"code": code})
	if err != nil {
		return
	}

	if err = s.removeExpireAtData(ctx, code); err != nil {
		return err
	}
	return nil
}

// SaveAccess writes AccessData.
// If RefreshToken is not blank, it must save in a way that can be loaded using LoadRefresh.
func (s *storage) SaveAccess(ctx context.Context, data *server.AccessData) (err error) {
	span, ctx := etrace.StartSpanFromContext(
		ctx,
		"mysqlStorage.SaveAccess",
	)
	defer span.Finish()

	prev := ""
	authorizeData := &server.AuthorizeData{}

	if data.AccessData != nil {
		prev = data.AccessData.AccessToken
	}

	if data.AuthorizeData != nil {
		authorizeData = data.AuthorizeData
	}

	extra := cast.ToString(data.UserData)

	tx := s.db.Begin()

	if data.RefreshToken != "" {
		if err := s.saveRefresh(ctx, tx, data.RefreshToken, data.AccessToken); err != nil {
			tx.Rollback()
			return err
		}
	}

	if data.Client == nil {
		return errors.New("data.Client must not be nil")
	}

	obj := dao.Access{
		Client:       data.Client.GetId(),
		Authorize:    authorizeData.Code,
		Previous:     prev,
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		ExpiresIn:    int(data.ExpiresIn),
		Scope:        data.Scope,
		RedirectUri:  data.RedirectUri,
		Ctime:        data.CreatedAt.Unix(),
		Extra:        extra,
	}

	err = dao.AccessCreate(ctx, tx, &obj)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = dao.AppInfoX(ctx, tx, egorm.Conds{
		"client_id": data.Client.GetId(),
	})
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.WithContext(ctx).Model(dao.App{}).Where("client_id = ?", data.Client.GetId()).Updates(egorm.Ups{
		"call_no": gorm.Expr("call_no+?", 1),
	}).Error
	if err != nil {
		tx.Rollback()
		return
	}
	err = s.AddExpireAtData(ctx, tx, data.AccessToken, data.ExpireAt())
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return nil
}

// LoadAccess retrieves access data by token. Client information MUST be loaded together.
// AuthorizeData and AccessData DON'T NEED to be loaded if not easily available.
// Optionally can return error if expired.
func (s *storage) LoadAccess(ctx context.Context, code string) (*server.AccessData, error) {
	span, ctx := etrace.StartSpanFromContext(
		ctx,
		"mysqlStorage.LoadAccess",
	)
	defer span.Finish()

	var result server.AccessData

	info, err := dao.AccessInfoX(ctx, s.db, egorm.Conds{"access_token": code})
	if err != nil {
		return nil, err
	}

	result.AccessToken = info.AccessToken
	result.RefreshToken = info.RefreshToken
	result.ExpiresIn = int32(info.ExpiresIn)
	result.Scope = info.Scope
	result.RedirectUri = info.RedirectUri
	result.CreatedAt = time.Unix(info.Ctime, 0)
	result.UserData = info.Extra
	client, err := s.GetClient(ctx, info.Client)
	if err != nil {
		return nil, err
	}

	result.Client = client
	result.AuthorizeData, _ = s.LoadAuthorize(ctx, info.Authorize)
	prevAccess, _ := s.LoadAccess(ctx, info.Previous)
	result.AccessData = prevAccess
	return &result, nil
}

// RemoveAccess revokes or deletes an AccessData.
func (s *storage) RemoveAccess(ctx context.Context, code string) (err error) {
	span, ctx := etrace.StartSpanFromContext(
		ctx,
		"mysqlStorage.RemoveAccess",
	)
	defer span.Finish()

	err = dao.AccessDeleteX(ctx, s.db, egorm.Conds{"access_token": code})
	if err != nil {
		return
	}
	err = s.removeExpireAtData(ctx, code)
	return
}

// LoadRefresh retrieves refresh AccessData. Client information MUST be loaded together.
// AuthorizeData and AccessData DON'T NEED to be loaded if not easily available.
// Optionally can return error if expired.
func (s *storage) LoadRefresh(ctx context.Context, code string) (*server.AccessData, error) {
	span, ctx := etrace.StartSpanFromContext(ctx, "mysqlStorage.LoadRefresh")
	defer span.Finish()

	info, err := dao.RefreshInfoX(ctx, s.db, egorm.Conds{"token": code})
	if err != nil {
		return nil, err
	}
	return s.LoadAccess(ctx, info.Access)
}

// RemoveRefresh revokes or deletes refresh AccessData.
func (s *storage) RemoveRefresh(ctx context.Context, code string) (err error) {
	span, ctx := etrace.StartSpanFromContext(ctx, "mysqlStorage.RemoveRefresh")
	defer span.Finish()

	err = dao.RefreshDeleteX(ctx, s.db, egorm.Conds{"token": code})
	return
}

// CreateClientWithInformation Makes easy to create a osin.DefaultClient
func (s *storage) CreateClientWithInformation(id string, secret string, redirectURI string, userData interface{}) server.Client {
	return &server.DefaultClient{
		Id:          id,
		Secret:      secret,
		RedirectUri: redirectURI,
		UserData:    userData,
	}
}

func (s *storage) saveRefresh(ctx context.Context, tx *gorm.DB, refresh, access string) (err error) {
	obj := dao.Refresh{
		Token:  refresh,
		Access: access,
	}

	err = dao.RefreshCreate(ctx, tx, &obj)
	return
}

// AddExpireAtData add info in expires table
func (s *storage) AddExpireAtData(ctx context.Context, tx *gorm.DB, code string, expireAt time.Time) (err error) {
	obj := dao.Expires{
		Token:     code,
		ExpiresAt: expireAt.Unix(),
	}
	err = dao.ExpiresCreate(ctx, tx, &obj)
	return
}

// removeExpireAtData remove info in expires table
func (s *storage) removeExpireAtData(ctx context.Context, code string) (err error) {
	err = dao.ExpiresDeleteX(ctx, s.db, egorm.Conds{"token": code})
	return
}
