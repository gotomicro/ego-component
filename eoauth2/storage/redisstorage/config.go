package redisstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gotomicro/ego-component/eoauth2/storage/dto"
	"github.com/gotomicro/ego-component/eredis"
)

//
//type config struct {
//	parentTokenClientId string // 父级token的client id，也就是单点登录系统，默认为ptk
//	/*
//		  key: sso:tkMap:{uid}
//		  value:
//			{parentTokenClientId}: tokenJsonInfo
//			{subTokenClientId1}:   tokenJsonInfo
//			{subTokenClientId2}:   tokenJsonInfo
//			{subTokenClientId3}:   tokenJsonInfo
//	*/
//	parentTokenMapSubTokenKey string // 存储token信息的hash map
//	/*
//		key: sso:ptk:{parentToken}
//		value: {userInfo}
//		ttl: 3600
//	*/
//	parentTokenKey string // 父级token存储用户信息
//	// key token value ptoken
//	/*
//		key: sso:stkMapPtk:{subToken}
//		value: {parentToken}
//		ttl: 3600
//	*/
//	subTokenMapParentTokenKey string // token与父级token的映射关系
//}

type config struct {
	/*
		    hashmap
			key: sso:uid:{uid}
			value:
				{clientType1}: parentTokenJsonInfo
				{clientType2}: parentTokenJsonInfo
	*/
	uidMapParentTokenKey string // 存储token信息的hash map
	/*
		     hashmap
			 key: sso:ptk:{parentToken}
			 value:
				userInfo:              userInfo
				tokenInfo:             tokenInfo
				{subTokenClientId1}:   tokenJsonInfo
				{subTokenClientId2}:   tokenJsonInfo
				{subTokenClientId3}:   tokenJsonInfo
			 ttl: 3600
	*/
	parentTokenMapSubTokenKey string // 存储token信息的hash map
	// key token value ptoken
	/*
		hashmap
		key: sso:stk:{subToken}
		value:
			parentToken: {parentToken}
			clientId:    {subTokenClientId}
			tokenInfo:   {tokenJsonInfo}
		ttl: 3600
	*/
	subTokenMapParentTokenKey string // token与父级token的映射关系
}

func defaultConfig() *config {
	return &config{
		uidMapParentTokenKey: "sso:uid:%d", // uid map parent token type
		//parentTokenClientId:       "ptk",
		parentTokenMapSubTokenKey: "sso:ptk:%d", //  parent token map
		//parentTokenKey:            "sso:ptk:%s",       // parent token
		subTokenMapParentTokenKey: "sso:stk:%s", // sub token map parent token
	}
}

//
//const (
//	//parentTokenClientId = "pToken"
//	//parentTokenMapSubTokenKey   = "sso:tokenHash:%d" // uid
//	//parentTokenKey   = "sso:pToken:%s"    // token                 key token value 用户信息
//	//subTokenMapParentTokenKey      = "sso:token:%s"     // token  string ptoken  key token value ptoken
//
//	/*
//		  key: ssoTokenHash:{uid}
//		  value:
//			{clientId1}: tokenJsonInfo
//			{clientId2}: tokenJsonInfo
//			{clientId3}: tokenJsonInfo
//	*/
//	tokenMapKeyPrefix = "sso:tokenHash:%d" // uid
//
//	/*
//		key: ssoParentToken:{token}
//		value: {userInfo}
//		ttl: 3600
//	*/
//	parentTokenPrefix = "sso:pToken:%s" // token
//
//	/*
//		key: ssoToken:{token}
//		value: {parentToken}
//		ttl: 3600
//	*/
//	tokenKeyPrefix = "sso:token:%s" // token
//)

type subToken struct {
	config             *config
	hashKeyParentToken string
	hashKeyClientId    string
	hashKeyTokenInfo   string
	redis              *eredis.Component
}

func newSubToken(config *config, redis *eredis.Component) *subToken {
	return &subToken{
		config:             config,
		hashKeyParentToken: "pToken",
		hashKeyClientId:    "clientId",
		hashKeyTokenInfo:   "tokenInfo",
		redis:              redis,
	}
}

func (s *subToken) getKey(subToken string) string {
	return fmt.Sprintf(s.config.subTokenMapParentTokenKey, subToken)
}

func (s *subToken) create(ctx context.Context, token dto.Token, parentToken string, clientId string) error {
	err := s.redis.HMSet(ctx, s.getKey(token.Token), map[string]interface{}{
		s.hashKeyParentToken: parentToken,
		s.hashKeyClientId:    clientId,
		s.hashKeyTokenInfo:   token,
	}, time.Duration(token.ExpiresIn)*time.Second)
	if err != nil {
		return fmt.Errorf("token.setToken: setTTL token failed, err:%w", err)
	}
	return nil
}

// 通过子系统token，获得父节点token
func (p *subToken) getParentToken(ctx context.Context, subToken string) (parentToken string, err error) {
	parentToken, err = p.redis.HGet(ctx, p.getKey(subToken), p.hashKeyParentToken)
	if err != nil {
		err = fmt.Errorf("subToken getParentToken error, %w", err)
		return
	}
	return
	//	if err != nil {
	//		if errors.Is(err, redis.Nil) {
	//			return nil, status.Error(codes.NotFound, "token is invalid or  has been expired")
	//		}
	//
	//		err = fmt.Errorf("getUserByToken, redis get parentToken error, %w", err)
	//		return
	//	}
}

type uidMapParentToken struct {
	config *config
	redis  *eredis.Component
}

func newUidMapParentToken(config *config, redis *eredis.Component) *uidMapParentToken {
	return &uidMapParentToken{
		config: config,
		redis:  redis,
	}
}

func (u *uidMapParentToken) getKey(uid int64) string {
	return fmt.Sprintf(u.config.uidMapParentTokenKey, uid)
}

func (u *uidMapParentToken) setToken(ctx context.Context, uid int64, clientType string, pToken dto.Token) error {
	pTokenByte, err := json.Marshal(pToken)
	if err != nil {
		return fmt.Errorf("uidMapParentToken setToken failed, err: %w", err)
	}

	return u.redis.HSet(ctx, u.getKey(uid), clientType, string(pTokenByte))
}

func (u *uidMapParentToken) getParentToken(ctx context.Context, uid int64, clientType string) (resp dto.Token, err error) {
	value, err := u.redis.HGet(ctx, u.getKey(uid), clientType)
	if err != nil {
		err = fmt.Errorf("uidMapParentToken getParentToken failed, err: %w", err)
		return
	}
	err = json.Unmarshal([]byte(value), &resp)
	return
}

//
//type parentToken struct {
//	key   string
//	redis *eredis.Component
//}
//
//func newParentToken(redis *eredis.Component, token string) *parentToken {
//	return &parentToken{
//		key:   fmt.Sprintf(parentTokenPrefix, token),
//
//		redis: redis,
//	}
//}
//
//func (p *parentToken) setTTL(ctx context.Context, userInfo *dto.User, ttl time.Duration) error {
//	userBytes, err := json.Marshal(userInfo)
//	if err != nil {
//		return fmt.Errorf("parentToken: marshal user info failed, err:%w", err)
//	}
//
//	// setTTL new token
//	err = p.redis.Set(ctx, p.key, string(userBytes), ttl)
//	if err != nil {
//		return fmt.Errorf("parentToken: setTTL token failed, err:%w", err)
//	}
//	return nil
//}
//
//func (p *parentToken) delete(ctx context.Context) error {
//	_, err := p.redis.Del(ctx, p.key)
//	if err != nil {
//		return fmt.Errorf("token.removeParentToken: remove token failed, err:%w", err)
//	}
//	return nil
//}
//

type parentToken struct {
	config           *config
	redis            *eredis.Component
	hashKeyCtime     string
	hashKeyUserInfo  string
	hashKeyTokenInfo string
}

func newParentToken(config *config, redis *eredis.Component) *parentToken {
	return &parentToken{
		config:           config,
		redis:            redis,
		hashKeyCtime:     "ctime",
		hashKeyUserInfo:  "userInfo",
		hashKeyTokenInfo: "tokenInfo",
	}
}

func (p *parentToken) getKey(pToken string) string {
	return fmt.Sprintf(p.config.parentTokenMapSubTokenKey, pToken)
}

func (p *parentToken) create(ctx context.Context, token dto.Token, userInfo *dto.User) error {
	err := p.redis.HMSet(ctx, p.getKey(token.Token), map[string]interface{}{
		p.hashKeyCtime:     time.Now().Unix(),
		p.hashKeyUserInfo:  userInfo,
		p.hashKeyTokenInfo: token,
	}, time.Duration(token.ExpiresIn)*time.Second)
	if err != nil {
		return fmt.Errorf("parentToken create failed, err:%w", err)
	}
	return nil
}

func (p *parentToken) delete(ctx context.Context, pToken string) error {
	_, err := p.redis.Del(ctx, p.getKey(pToken))
	if err != nil {
		return fmt.Errorf("token.removeParentToken: remove token failed, err:%w", err)
	}
	return nil
}

func (p *parentToken) getUser(ctx context.Context, pToken string) (userInfo *dto.User, err error) {
	// 根据父节点token，获取用户信息
	userInfoStr, err := p.redis.HGet(ctx, p.getKey(pToken), p.hashKeyUserInfo)
	if err != nil {
		err = fmt.Errorf("parentToken getUser failed, err: %w", err)
		return
	}

	err = json.Unmarshal([]byte(userInfoStr), &userInfo)
	if err != nil {
		err = fmt.Errorf("getUserByToken json unmarshal error, %w", err)
		return
	}
	return
}

func (p *parentToken) setToken(ctx context.Context, pToken string, clientId string, token dto.Token) error {
	// 如果不存在key，报错
	hashKeyCtimeValue, err := p.redis.HGet(ctx, p.getKey(pToken), p.hashKeyCtime)
	if err != nil {
		return fmt.Errorf("parentToken setToken key empty1, err: %w", err)
	}
	if hashKeyCtimeValue == "" {
		return fmt.Errorf("parentToken setToken key empty2, err: %w", err)
	}

	tokenJsonInfo, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("parentToken setToken error, err: %w", err)
	}

	// setTTL token map
	err = p.redis.HSet(ctx, p.getKey(pToken), clientId, string(tokenJsonInfo))
	if err != nil {
		return fmt.Errorf("token.setParentToken: setTTL token map failed, err:%w", err)
	}
	return nil
}

func (h *parentToken) getToken(ctx context.Context, pToken string, clientId string) (tokenInfo dto.Token, err error) {
	tokenValue, err := h.redis.HGet(ctx, h.getKey(pToken), clientId)
	if err != nil {
		err = fmt.Errorf("tokgen get redis hmget string error, %w", err)
		return
	}

	err = json.Unmarshal([]byte(tokenValue), &tokenInfo)
	if err != nil {
		err = fmt.Errorf("redis token info json unmarshal errorr, err: %w", err)
		return
	}
	return
}
