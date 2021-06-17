package redisstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gotomicro/ego-component/eoauth2/storage/dto"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
)

const (
	DefaultTokenExpireIn = 24 * 60 * 60

	tokenRefreshLockPrefix = "ssoTokenRefreshLock:%s"
	newTokenKeyPrefix      = "ssoNewToken:%s"
)

type tokenServer struct {
	redis             *eredis.Component
	uidMapParentToken *uidMapParentToken
	parentToken       *parentToken
	subToken          *subToken
}

func initTokenServer(config *config, redis *eredis.Component) *tokenServer {
	return &tokenServer{
		redis:             redis,
		uidMapParentToken: newUidMapParentToken(config, redis),
		parentToken:       newParentToken(config, redis),
		subToken:          newSubToken(config, redis),
	}
}

// setParentToken sso的父节点token
func (t *tokenServer) setParentToken(ctx context.Context, pToken dto.Token, userInfo *dto.User) (err error) {
	// 1 设置uid 到 parent token关系
	err = t.uidMapParentToken.setToken(ctx, userInfo.Uid, "pc", pToken)
	if err != nil {
		return fmt.Errorf("token.setParentToken: create token map failed, err:%w", err)
	}

	// 2 创建父级的token信息
	return t.parentToken.create(ctx, pToken, userInfo)
}

func (t *tokenServer) setToken(ctx context.Context, clientId string, token dto.Token, pToken string) (err error) {
	err = t.parentToken.setToken(ctx, pToken, clientId, token)
	if err != nil {
		return fmt.Errorf("tokenServer.setToken failed, err:%w", err)
	}

	// setTTL new token
	err = t.subToken.create(ctx, token, pToken, clientId)
	return
}

func (t *tokenServer) removeParentToken(ctx context.Context, pToken string) (err error) {
	return t.parentToken.delete(ctx, pToken)
}

// 获取父级token
func (t *tokenServer) getParentToken(uid int64) (tokenInfo dto.Token, err error) {
	return t.uidMapParentToken.getParentToken(context.Background(), uid, "pc")
}

func (t *tokenServer) getToken(clientId string, pToken string) (tokenInfo dto.Token, err error) {
	return t.parentToken.getToken(context.Background(), pToken, clientId)
}

func (t *tokenServer) getUserByParentToken(ctx context.Context, pToken string) (info *dto.User, err error) {
	return t.parentToken.getUser(ctx, pToken)
}

func (t *tokenServer) getUserByToken(ctx context.Context, token string) (info *dto.User, err error) {
	// 通过子系统token，获得父节点token
	pToken, err := t.subToken.getParentToken(ctx, token)
	if err != nil {
		return
	}
	return t.getUserByParentToken(ctx, pToken)
}

func (t *tokenServer) refreshToken(ctx context.Context, clientId string, pToken string) (tk *dto.Token, err error) {
	var genNewToken dto.Token
	// try to get lock
	tokenRefreshLock, err := t.redis.LockClient().Obtain(ctx, redisTokenRefreshLockKey(pToken), 100*time.Millisecond,
		eredis.WithLockOptionRetryStrategy(eredis.LinearBackoffRetry(10*time.Millisecond)))
	if err != nil {
		return nil, err
	}

	defer func() {
		err = tokenRefreshLock.Release(ctx)
		if err != nil {
			elog.Error("tokenServer.genNewToken: release redis lock failed", zap.Error(err),
				zap.String("clientId", clientId), zap.String("pToken", pToken))
		}
	}()

	// try to get new-token from cache
	{
		tk, err = t.getNewTokenFromCache(ctx, pToken)
		if err != nil && !errors.Is(err, eredis.Nil) { // no-empty error
			return nil, err
		} else if err == nil {
			return tk, nil
		} else {
			// empty cache
		}
	}

	// get user info
	//ssoUser, err := t.getUserByParentToken(ctx, pToken)
	//if err != nil {
	//	return nil, err
	//}

	// re-generate token
	{
		//tk = &dto.Token{
		//	Token:     base64.RawURLEncoding.EncodeToString(uuid.NewRandom()),
		//	AuthAt:    time.Now().Unix(),
		//	ExpiresIn: DefaultTokenExpireIn,
		//}

		genNewToken = dto.NewToken(DefaultTokenExpireIn)
		tk = &genNewToken
		err = t.setToken(ctx, clientId, genNewToken, pToken)
		if err != nil {
			return
		}
	}

	// write new-token to cache
	err = t.setNewTokenToCache(ctx, tk, pToken)
	if err != nil {
		elog.Error("tokenServer.genNewToken: setTTL new-token to cache failed", zap.Error(err))
		return tk, nil
	}

	return
}

func (t *tokenServer) getNewTokenFromCache(ctx context.Context, pToken string) (tk *dto.Token, err error) {
	newTokenKey := redisNewTokenKey(pToken)

	newTokenBytes, err := t.redis.GetBytes(ctx, newTokenKey)
	if err != nil {
		return nil, err
	}

	tk = &dto.Token{}
	err = json.Unmarshal(newTokenBytes, tk)
	if err != nil {
		return nil, err
	}

	return
}

func (t *tokenServer) setNewTokenToCache(ctx context.Context, tk *dto.Token, pToken string) (err error) {
	tkBytes, err := json.Marshal(tk)
	if err != nil {
		return
	}

	// write cache
	err = t.redis.SetEX(ctx, redisNewTokenKey(pToken), string(tkBytes), time.Minute)
	if err != nil {
		return
	}

	return
}

//func redisTokenLockKey(uid int64) string {
//	return fmt.Sprintf(tokenLockPrefix, uid)
//}

func redisTokenRefreshLockKey(pToken string) string {
	return fmt.Sprintf(tokenRefreshLockPrefix, pToken)
}

func redisNewTokenKey(pToken string) string {
	return fmt.Sprintf(newTokenKeyPrefix, pToken)
}
