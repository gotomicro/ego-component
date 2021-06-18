package redisstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gotomicro/ego-component/egorm"
	"github.com/gotomicro/ego-component/eoauth2/server"
	"github.com/gotomicro/ego-component/eoauth2/storage/dao"
	"github.com/gotomicro/ego-component/eoauth2/storage/dto"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/etrace"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type Storage struct {
	db          *egorm.Component
	logger      *elog.Component
	tokenServer *tokenServer
	config      *config
}

// NewStorage returns a new mysql Storage instance.
func NewStorage(db *egorm.Component, redis *eredis.Component, logger *elog.Component, options ...Option) *Storage {
	container := &Storage{
		db:     db,
		logger: logger,
		config: defaultConfig(),
	}
	for _, option := range options {
		option(container)
	}
	tSrv := initTokenServer(container.config, redis)
	container.tokenServer = tSrv
	return container
}

// Clone the Storage if needed. For example, using mgo, you can clone the session with session.Clone
// to avoid concurrent access problems.
// This is to avoid cloning the connection at each method access.
// Can return itself if not a problem.
func (s *Storage) Clone() server.Storage {
	return s
}

// Close the resources the Storage potentially holds (using Clone for example)
func (s *Storage) Close() {
}

// GetClient loads the client by id
func (s *Storage) GetClient(ctx context.Context, clientId string) (client server.Client, err error) {
	span, ctx := etrace.StartSpanFromContext(ctx, "redisStorage.GetClient")
	defer span.Finish()

	app, err := dao.AppInfoX(ctx, s.db, egorm.Conds{"client_id": clientId})
	if err != nil {
		return
	}
	c := server.DefaultClient{
		Id:          app.ClientId,
		Secret:      app.Secret,
		RedirectUri: app.RedirectUri,
	}
	return &c, nil
}

// SaveAuthorize saves authorize data.
func (s *Storage) SaveAuthorize(ctx context.Context, data *server.AuthorizeData) (err error) {
	span, ctx := etrace.StartSpanFromContext(ctx, "redisStorage.SaveAuthorize")
	defer span.Finish()

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
	tx := s.db.Begin()
	err = dao.AuthorizeCreate(ctx, tx, &obj)
	if err != nil {
		tx.Rollback()
		return
	}

	err = s.addExpireAtData(ctx, tx, data.Code, data.ExpireAt())
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
func (s *Storage) LoadAuthorize(ctx context.Context, code string) (*server.AuthorizeData, error) {
	span, ctx := etrace.StartSpanFromContext(context.Background(), "redisStorage.LoadAuthorize")
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
func (s *Storage) RemoveAuthorize(ctx context.Context, code string) (err error) {
	span, ctx := etrace.StartSpanFromContext(context.Background(), "redisStorage.RemoveAuthorize")
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
func (s *Storage) SaveAccess(ctx context.Context, data *server.AccessData) (err error) {
	prev := ""
	authorizeData := &server.AuthorizeData{}

	if data.AccessData != nil {
		prev = data.AccessData.AccessToken
	}

	if data.AuthorizeData != nil {
		authorizeData = data.AuthorizeData
	}

	span, ctx := etrace.StartSpanFromContext(
		ctx,
		"redisStorage.SaveAccess",
	)
	defer span.Finish()

	extra := cast.ToString(data.UserData)

	var ssoUser *dto.User
	err = json.Unmarshal([]byte(extra), &ssoUser)
	if err != nil {
		return fmt.Errorf("解析登录用户json数据失败, err: %w", err)
	}

	tx := s.db.Begin()

	// 1 获取父级token，也可以认为是refresh token
	pToken, err := s.tokenServer.getParentToken(ssoUser.Uid)
	if err != nil {
		return err
	}

	data.RefreshToken = pToken.Token

	// 创建parent token和sub token关系
	//if err = s.saveRefresh(ctx, tx, pToken.Token, data.AccessToken); err != nil {
	//	tx.Rollback()
	//	return err
	//}

	//if data.RefreshToken != "" {
	//	if err := s.saveRefresh(ctx, tx, data.RefreshToken, data.AccessToken); err != nil {
	//		tx.Rollback()
	//		return err
	//	}
	//}

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

	err = s.addExpireAtData(ctx, tx, data.AccessToken, data.ExpireAt())
	if err != nil {
		tx.Rollback()
		return
	}

	err = s.tokenServer.createToken(ctx, data.Client.GetId(), dto.Token{
		Token:     data.AccessToken,
		AuthAt:    time.Now().Unix(),
		ExpiresIn: DefaultTokenExpireIn,
	}, pToken.Token)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("设置redis token失败, err:%w", err)
	}

	tx.Commit()
	return nil
}

// LoadAccess retrieves access data by token. Client information MUST be loaded together.
// AuthorizeData and AccessData DON'T NEED to be loaded if not easily available.
// Optionally can return error if expired.
func (s *Storage) LoadAccess(ctx context.Context, token string) (*server.AccessData, error) {
	span, ctx := etrace.StartSpanFromContext(ctx, "redisStorage.LoadAccess")
	defer span.Finish()

	var result server.AccessData

	info, err := dao.AccessInfoX(ctx, s.db, egorm.Conds{"access_token": token})
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
func (s *Storage) RemoveAccess(ctx context.Context, token string) (err error) {
	span, ctx := etrace.StartSpanFromContext(ctx, "redisStorage.RemoveAccess")
	defer span.Finish()

	err = dao.AccessDeleteX(ctx, s.db, egorm.Conds{"access_token": token})
	if err != nil {
		return
	}
	err = s.removeExpireAtData(ctx, token)
	if err != nil {
		return
	}
	pToken, err := s.tokenServer.getParentTokenByToken(ctx, token)
	if err != nil {
		return err
	}

	// 删除redis token
	s.tokenServer.removeParentToken(ctx, pToken)
	return
}

// LoadRefresh retrieves refresh AccessData. Client information MUST be loaded together.
// AuthorizeData and AccessData DON'T NEED to be loaded if not easily available.
// Optionally can return error if expired
func (s *Storage) LoadRefresh(ctx context.Context, token string) (*server.AccessData, error) {
	return nil, fmt.Errorf("not implement")
	// 这里的refresh token，实际上是parent token
	//span, ctx := etrace.StartSpanFromContext(context.Background(), "redisStorage.LoadRefresh")
	//defer span.Finish()
	//
	//info, err := dao.RefreshInfoX(ctx, s.db, egorm.Conds{"token": token})
	//if err != nil {
	//	return nil, err
	//}
	//accessInfo, err := dao.AccessInfoX(ctx, s.db, egorm.Conds{"access_token": info.Access})
	//if err != nil {
	//	return nil, err
	//}
	//var result server.AccessData
	//
	//result.AccessToken = accessInfo.AccessToken
	//result.RefreshToken = token
	//result.ExpiresIn = int32(accessInfo.ExpiresIn)
	//result.Scope = accessInfo.Scope
	//result.RedirectUri = accessInfo.RedirectUri
	//result.CreatedAt = time.Unix(accessInfo.Ctime, 0)
	//result.UserData = accessInfo.Extra
	//client, err := s.GetClient(ctx, accessInfo.Client)
	//if err != nil {
	//	return nil, err
	//}
	//
	//result.Client = client
	//
	//tk, err := s.tokenServer.refreshToken(ctx, client.GetId(), token)
	//if err != nil {
	//	return nil, err
	//}
	//
	//var result server.AccessData
	//result.Client = client
	//result.AccessToken = tk.Token
	//result.CreatedAt = time.Unix(tk.AuthAt, 0)
	//result.ExpiresIn = int32(tk.ExpiresIn)
	//return result, 0
}

// RemoveRefresh revokes or deletes refresh AccessData.
func (s *Storage) RemoveRefresh(ctx context.Context, code string) (err error) {
	span, ctx := etrace.StartSpanFromContext(context.Background(), "redisStorage.RemoveRefresh")
	defer span.Finish()

	err = dao.RefreshDeleteX(ctx, s.db, egorm.Conds{"token": code})
	return
}

// CreateClientWithInformation Makes easy to create a osin.DefaultClient
func (s *Storage) CreateClientWithInformation(id string, secret string, redirectURI string, userData interface{}) server.Client {
	return &server.DefaultClient{
		Id:          id,
		Secret:      secret,
		RedirectUri: redirectURI,
		UserData:    userData,
	}
}

func (s *Storage) saveRefresh(ctx context.Context, tx *gorm.DB, refresh, access string) (err error) {
	obj := dao.Refresh{
		Token:  refresh,
		Access: access,
	}

	err = dao.RefreshCreate(ctx, tx, &obj)
	return
}

// addExpireAtData add info in expires table
func (s *Storage) addExpireAtData(ctx context.Context, tx *gorm.DB, code string, expireAt time.Time) (err error) {
	obj := dao.Expires{
		Token:     code,
		ExpiresAt: expireAt.Unix(),
	}
	err = dao.ExpiresCreate(ctx, tx, &obj)
	return
}

// removeExpireAtData remove info in expires table
func (s *Storage) removeExpireAtData(ctx context.Context, code string) (err error) {
	err = dao.ExpiresDeleteX(ctx, s.db, egorm.Conds{"token": code})
	return
}

// CreateParentToken 创建父级token
func (s *Storage) CreateParentToken(ctx context.Context, pToken dto.Token, userInfo *dto.User) (err error) {
	return s.tokenServer.createParentToken(ctx, pToken, userInfo)
}

// RenewParentToken 续期父级token
func (s *Storage) RenewParentToken(ctx context.Context, pToken dto.Token) (err error) {
	return s.tokenServer.renewParentToken(ctx, pToken)
}

func (s *Storage) GetUserByParentToken(ctx context.Context, token string) (info *dto.User, err error) {
	return s.tokenServer.getUserByParentToken(ctx, token)
}

func (s *Storage) RemoveParentToken(ctx context.Context, pToken string) (err error) {
	return s.tokenServer.removeParentToken(ctx, pToken)
}

// CreateToken 创建子系统token
func (s *Storage) CreateToken(ctx context.Context, clientId string, token dto.Token, pToken string) (err error) {
	return s.tokenServer.createToken(ctx, clientId, token, pToken)
}

func (s *Storage) GetUserByToken(ctx context.Context, token string) (info *dto.User, err error) {
	return s.tokenServer.getUserByToken(ctx, token)
}

func (s *Storage) RefreshToken(ctx context.Context, clientId string, pToken string) (tk *dto.Token, err error) {
	return s.tokenServer.refreshToken(ctx, clientId, pToken)
}
