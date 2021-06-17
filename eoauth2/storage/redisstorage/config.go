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
		uidMapParentTokenKey:      "sso:uid:%d", // uid map parent token type
		parentTokenMapSubTokenKey: "sso:ptk:%s", //  parent token map
		subTokenMapParentTokenKey: "sso:stk:%s", // sub token map parent token
	}
}

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
	tokenStr, err := token.Marshal()
	if err != nil {
		return err
	}
	err = s.redis.HMSet(ctx, s.getKey(token.Token), map[string]interface{}{
		s.hashKeyParentToken: parentToken,
		s.hashKeyClientId:    clientId,
		s.hashKeyTokenInfo:   tokenStr,
	}, time.Duration(token.ExpiresIn)*time.Second)
	if err != nil {
		return fmt.Errorf("subToken.create token failed, err:%w", err)
	}
	return nil
}

// 通过子系统token，获得父节点token
func (s *subToken) getParentToken(ctx context.Context, subToken string) (parentToken string, err error) {
	parentToken, err = s.redis.HGet(ctx, s.getKey(subToken), s.hashKeyParentToken)
	if err != nil {
		err = fmt.Errorf("subToken.getParentToken failed, %w", err)
		return
	}
	return
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
		return fmt.Errorf("uidMapParentToken.setToken failed, err: %w", err)
	}

	return u.redis.HSet(ctx, u.getKey(uid), clientType, string(pTokenByte))
}

func (u *uidMapParentToken) getParentToken(ctx context.Context, uid int64, clientType string) (resp dto.Token, err error) {
	value, err := u.redis.HGet(ctx, u.getKey(uid), clientType)
	if err != nil {
		err = fmt.Errorf("uidMapParentToken.getParentToken failed, err: %w", err)
		return
	}
	err = json.Unmarshal([]byte(value), &resp)
	return
}

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
	userStr, err := userInfo.Marshal()
	if err != nil {
		return err
	}

	tokenStr, err := token.Marshal()
	if err != nil {
		return err
	}
	err = p.redis.HMSet(ctx, p.getKey(token.Token), map[string]interface{}{
		p.hashKeyCtime:     time.Now().Unix(),
		p.hashKeyUserInfo:  userStr,
		p.hashKeyTokenInfo: tokenStr,
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
	_, err := p.redis.HGet(ctx, p.getKey(pToken), p.hashKeyCtime)
	if err != nil {
		return fmt.Errorf("parentToken.setToken get key empty, err: %w", err)
	}

	tokenJsonInfo, err := token.Marshal()
	if err != nil {
		return fmt.Errorf("parentToken.setToken json marshal failed, err: %w", err)
	}

	// setTTL token map
	err = p.redis.HSet(ctx, p.getKey(pToken), clientId, string(tokenJsonInfo))
	if err != nil {
		return fmt.Errorf("parentToken.setToken hset failed, err:%w", err)
	}
	return nil
}

func (p *parentToken) getToken(ctx context.Context, pToken string, clientId string) (tokenInfo dto.Token, err error) {
	tokenValue, err := p.redis.HGet(ctx, p.getKey(pToken), clientId)
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
