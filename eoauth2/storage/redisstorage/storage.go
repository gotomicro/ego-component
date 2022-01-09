package redisstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotomicro/ego-component/egorm"
	"github.com/gotomicro/ego-component/eoauth2/server"
	"github.com/gotomicro/ego-component/eoauth2/storage/dao"
	"github.com/gotomicro/ego-component/eoauth2/storage/dto"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type Storage struct {
	db          *egorm.Component
	logger      *elog.Component
	tokenServer *tokenServer
	config      *config
}

// NewStorage returns a new redis Storage instance.
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

	err = s.addExpireAtData(ctx, tx, data.Code, data.ExpireAt(), data.ParentToken)
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
	prevToken := ""
	authorizeData := &server.AuthorizeData{}

	// 之前的access token
	// 如果是authorize token，那么该数据为空
	// 如果是refresh token，有这个数据
	if data.AccessData != nil {
		prevToken = data.AccessData.AccessToken
	}

	// 如果是authorize token，有这个数据
	// 如果是refresh token，那么该数据为空
	if data.AuthorizeData != nil {
		authorizeData = data.AuthorizeData
	}

	extra := cast.ToString(data.UserData)

	tx := s.db.Begin()

	pToken := ""
	// 这种是在authorize token的时候，会有code信息
	if authorizeData.Code != "" {
		// 根据之前code码，取出parent token信息
		expires, err := dao.ExpiresX(ctx, s.db, egorm.Conds{
			"token": authorizeData.Code,
		})
		if err != nil {
			return fmt.Errorf("pToken not found1, err: %w", err)
		}
		pToken = expires.Ptoken
		// refresh token的时候，没有该信息
		// 1 拿到原先的sub token，看是否有效
		// 2 再从sub token中找到对应parent token，看是否有效
		// 3 刷新token
		// 从load refresh里拿到老的access token信息，查询到ptoken，并处理老token的逻辑
	} else {
		// todo 老的token是需要将过期时间变短
		pToken, err = s.tokenServer.getParentTokenByToken(ctx, prevToken)
		if err != nil {
			return fmt.Errorf("pToken not found2, err: %w", err)
		}
	}
	if pToken == "" {
		return fmt.Errorf("ptoken is empty")
	}

	if data.Client == nil {
		return errors.New("data.Client must not be nil")
	}

	obj := dao.Access{
		Client:       data.Client.GetId(),
		Authorize:    authorizeData.Code,
		Previous:     prevToken,
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

	err = tx.WithContext(ctx).Model(dao.App{}).Where("client_id = ?", data.Client.GetId()).Updates(map[string]interface{}{
		"call_no": gorm.Expr("call_no+?", 1),
	}).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = s.tokenServer.createToken(ctx, data.Client.GetId(), dto.Token{
		Token:     data.AccessToken,
		AuthAt:    time.Now().Unix(),
		ExpiresIn: s.config.parentAccessExpiration,
	}, pToken)
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
	return &result, nil
}

// RemoveAccess revokes or deletes an AccessData.
func (s *Storage) RemoveAccess(ctx context.Context, token string) (err error) {
	err = dao.AccessDeleteX(ctx, s.db, egorm.Conds{"access_token": token})
	if err != nil {
		return
	}
	err = s.removeExpireAtData(ctx, token)
	if err != nil {
		return
	}

	// todo 应该移除子节点token，在这里设置expire sub token
	// 不能删除parent token

	//pToken, err := s.tokenServer.getParentTokenByToken(ctx, token)
	//if err != nil {
	//	return err
	//}

	// 删除redis token
	//s.tokenServer.removeParentToken(ctx, pToken)
	return
}

// RemoveAllAccess 通过token，删除自己的token，以及父token
func (s *Storage) RemoveAllAccess(ctx context.Context, token string) (err error) {
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
	return s.tokenServer.removeParentToken(ctx, pToken)
}

// LoadRefresh retrieves refresh AccessData. Client information MUST be loaded together.
// 原本的load refresh，是使用refresh token来换取新的token，但是在单点登录下，可以简单操作。
// 1 拿到原先的sub token，看是否有效
// 2 再从sub token中找到对应parent token，看是否有效
// 3 刷新token
// 必须要这个信息用于给予access token，告诉oauth2老的token，用于在save access的时候，查询到ptoken，并处理老token的逻辑
// AuthorizeData and AccessData DON'T NEED to be loaded if not easily available.
// Optionally can return error if expired
func (s *Storage) LoadRefresh(ctx context.Context, token string) (*server.AccessData, error) {
	return s.LoadAccess(ctx, token)
}

// RemoveRefresh revokes or deletes refresh AccessData.
func (s *Storage) RemoveRefresh(ctx context.Context, code string) (err error) {
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
func (s *Storage) addExpireAtData(ctx context.Context, tx *gorm.DB, code string, expireAt time.Time, parentToken string) (err error) {
	obj := dao.Expires{
		Token:     code,
		ExpiresAt: expireAt.Unix(),
		Ptoken:    parentToken,
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
func (s *Storage) CreateParentToken(ctx context.Context, pToken dto.Token, uid int64, platform string) (err error) {
	return s.tokenServer.createParentToken(ctx, pToken, uid, platform)
}

// RenewParentToken 续期父级token
func (s *Storage) RenewParentToken(ctx context.Context, pToken dto.Token) (err error) {
	return s.tokenServer.renewParentToken(ctx, pToken)
}

func (s *Storage) GetUidByParentToken(ctx context.Context, token string) (uid int64, err error) {
	return s.tokenServer.getUidByParentToken(ctx, token)
}

func (s *Storage) RemoveParentToken(ctx context.Context, pToken string) (err error) {
	return s.tokenServer.removeParentToken(ctx, pToken)
}

// CreateToken 创建子系统token
func (s *Storage) CreateToken(ctx context.Context, clientId string, token dto.Token, pToken string) (err error) {
	return s.tokenServer.createToken(ctx, clientId, token, pToken)
}

func (s *Storage) GetUidByToken(ctx context.Context, token string) (uid int64, err error) {
	return s.tokenServer.getUidByToken(ctx, token)
}

func (s *Storage) RefreshToken(ctx context.Context, clientId string, pToken string) (tk *dto.Token, err error) {
	return s.tokenServer.refreshToken(ctx, clientId, pToken)
}
