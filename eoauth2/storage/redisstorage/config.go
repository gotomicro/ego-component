package redisstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gotomicro/ego-component/eoauth2/storage/dto"
	"github.com/gotomicro/ego-component/eredis"
)

type config struct {
	parentAccessExpiration int64 // 父亲节点token

	/*
		    hashmap
			key: sso:uid:{uid}
			expiration: 最大的过期时间
			value:
				expireList:                [{"clientType1|parentToken":"ctime"}]
				expireTime:                最大过期时间
				{clientType1|parentToken}: parentTokenJsonInfo
				{clientType2|parentToken}: parentTokenJsonInfo
	*/
	uidMapParentTokenKey      string // 存储token信息的hash map
	uidMapParentTokenFieldKey string // 存储token信息的hash map的field key  {clientType1|parentToken}
	/*
				     hashmap
					 key: sso:ptk:{parentToken}
		  			 expiration: 最大的过期时间
					 value:
						uid:                   uid
						tokenInfo:             tokenInfo
						expireList:             [{"subTokenClientId1":"ctime"}]
						expireTime:            最大过期时间
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
	//clientType                []string // 支持的客户端类型，web、andorid、ios，用于设置一个客户端，可以登录几个parent token。
}

func defaultConfig() *config {
	return &config{
		uidMapParentTokenKey:      "sso:uid:%d", // uid map parent token type
		uidMapParentTokenFieldKey: "%s|%s",      // uid map parent token type
		parentTokenMapSubTokenKey: "sso:ptk:%s", //  parent token map
		subTokenMapParentTokenKey: "sso:stk:%s", // sub token map parent token
		parentAccessExpiration:    24 * 3600,
		//platform:                []string{"web", "android", "ios"},
	}
}

type subToken struct {
	config             *config
	hashKeyParentToken string
	hashKeyClientId    string
	hashKeyTokenInfo   string
	hashKeyCtime       string
	redis              *eredis.Component
}

func newSubToken(config *config, redis *eredis.Component) *subToken {
	return &subToken{
		config:             config,
		hashKeyCtime:       "_c", // create time
		hashKeyParentToken: "_pt",
		hashKeyClientId:    "_id",
		hashKeyTokenInfo:   "_t",
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
		s.hashKeyCtime:       time.Now().Unix(),
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
	config             *config
	redis              *eredis.Component
	hashExpireTimeList string
	hashExpireTime     string
}

func newUidMapParentToken(config *config, redis *eredis.Component) *uidMapParentToken {
	return &uidMapParentToken{
		config:             config,
		redis:              redis,
		hashExpireTimeList: "_etl", // expire time List
		hashExpireTime:     "_et",  // expire time，最大过期时间，unix时间戳，到了时间就会过期被删除
	}
}

func (u *uidMapParentToken) getKey(uid int64) string {
	return fmt.Sprintf(u.config.uidMapParentTokenKey, uid)
}

func (u *uidMapParentToken) getFieldKey(clientType string, parentToken string) string {
	return fmt.Sprintf(u.config.uidMapParentTokenFieldKey, clientType, parentToken)
}

// 并发操作redis情况不考虑，因为一个用户使用多个终端，并发登录极其少见
// 1 先取出这个key里面的数据
//   expireTimeList:            [{"clientType1|parentToken":"expire的时间戳"}]
//	 expireTime:                最大过期时间

func (u *uidMapParentToken) setToken(ctx context.Context, uid int64, platform string, pToken dto.Token) error {
	fieldKey := u.getFieldKey(platform, pToken.Token)

	expireTime, err := u.getExpireTime(ctx, uid)
	if err != nil {
		return err
	}
	expireTimeList, err := u.getExpireTimeList(ctx, uid)
	if err != nil {
		return err
	}
	nowTime := time.Now().Unix()
	newExpireTimeList := make(uidTokenExpires, 0)
	// 新数据添加到队列前面，这样方便后续清除数据，或者对数据做一些限制
	newExpireTimeList = append(newExpireTimeList, uidTokenExpire{
		Token:      fieldKey,
		ExpireTime: nowTime + pToken.ExpiresIn,
	})

	// 删除过期的数据
	hdelFields := make([]string, 0)
	for _, value := range expireTimeList {
		// 过期时间小于当前时间，那么需要删除
		if value.ExpireTime <= nowTime {
			hdelFields = append(hdelFields, value.Token)
			continue
		}
		newExpireTimeList = append(newExpireTimeList, value)
	}
	if len(hdelFields) > 0 {
		err = u.redis.HDel(ctx, u.getKey(uid), hdelFields...)
		if err != nil {
			return fmt.Errorf("uidMapParentToken setToken HDel expire data failed, error: %w", err)
		}
	}

	err = u.redis.HSet(ctx, u.getKey(uid), u.hashExpireTimeList, newExpireTimeList.Marshal())
	if err != nil {
		return fmt.Errorf("uidMapParentToken setToken HSet expire time failed, error: %w", err)
	}

	// 将parent token信息存入
	pTokenByte, err := pToken.Marshal()
	if err != nil {
		return fmt.Errorf("uidMapParentToken.createToken failed, err: %w", err)
	}

	err = u.redis.HSet(ctx, u.getKey(uid), u.getFieldKey(platform, pToken.Token), pTokenByte)
	if err != nil {
		return fmt.Errorf("uidMapParentToken setToken HSet token info failed, error: %w", err)
	}

	// 如果之前没数据，那么expireTime为0，所以会写入
	// 新的token大于，之前的过期时间，所以需要续期
	if pToken.ExpiresIn+nowTime > expireTime {
		err = u.redis.HSet(ctx, u.getKey(uid), u.hashExpireTime, pToken.ExpiresIn+nowTime)
		if err != nil {
			return fmt.Errorf("uidMapParentToken setToken HSet expire time failed, error: %w", err)
		}

		err = u.redis.Client().Expire(ctx, u.getKey(uid), time.Duration(pToken.ExpiresIn)*time.Second).Err()
		if err != nil {
			return fmt.Errorf("uidMapParentToken setToken expire error %w", err)
		}
	}

	return nil
}

// 获取过期时间，最新的在最前面。
func (u *uidMapParentToken) getExpireTimeList(ctx context.Context, uid int64) (userInfo uidTokenExpires, err error) {
	// 根据父节点token，获取用户信息
	infoBytes, err := u.redis.Client().HGet(ctx, u.getKey(uid), u.hashExpireTimeList).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		err = fmt.Errorf("uidMapParentToken getExpireTimeList failed, err: %w", err)
		return
	}
	if errors.Is(err, redis.Nil) {
		err = nil
		return
	}

	pUserInfo := &userInfo
	err = pUserInfo.Unmarshal(infoBytes)
	if err != nil {
		err = fmt.Errorf("uidMapParentToken getExpireTimeList json unmarshal error, %w", err)
		return
	}
	return
}

// 获取过期时间，快过期的在最前面。
func (u *uidMapParentToken) getExpireTime(ctx context.Context, uid int64) (expireTime int64, err error) {
	// 根据父节点token，获取用户信息
	expireTime, err = u.redis.Client().HGet(ctx, u.getKey(uid), u.hashExpireTime).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		err = fmt.Errorf("uidMapParentToken getExpireTime failed, err: %w", err)
		return
	}
	if errors.Is(err, redis.Nil) {
		err = nil
	}
	return
}

//func (u *uidMapParentToken) getParentToken(ctx context.Context, uid int64, clientType string) (resp dto.Token, err error) {
//	value, err := u.redis.HGet(ctx, u.getKey(uid), clientType)
//	if err != nil {
//		err = fmt.Errorf("uidMapParentToken.getParentToken failed, err: %w", err)
//		return
//	}
//	err = json.Unmarshal([]byte(value), &resp)
//	return
//}

type parentToken struct {
	config             *config
	redis              *eredis.Component
	hashKeyCtime       string
	hashKeyUid         string
	hashKeyPlatform    string
	hashExpireTimeList string
}

func newParentToken(config *config, redis *eredis.Component) *parentToken {
	return &parentToken{
		config:             config,
		redis:              redis,
		hashKeyCtime:       "_c",   // create time
		hashKeyPlatform:    "_p",   // 类型
		hashKeyUid:         "_u",   // uid
		hashExpireTimeList: "_etl", // expire time List

	}
}

func (p *parentToken) getKey(pToken string) string {
	return fmt.Sprintf(p.config.parentTokenMapSubTokenKey, pToken)
}

func (p *parentToken) create(ctx context.Context, pToken dto.Token, platform string, uid int64) error {
	err := p.redis.HMSet(ctx, p.getKey(pToken.Token), map[string]interface{}{
		p.hashKeyCtime:    time.Now().Unix(),
		p.hashKeyUid:      uid,
		p.hashKeyPlatform: platform,
	}, time.Duration(pToken.ExpiresIn)*time.Second)
	if err != nil {
		return fmt.Errorf("parentToken.create failed, err:%w", err)
	}
	return nil
}

func (p *parentToken) renew(ctx context.Context, pToken dto.Token) error {
	err := p.redis.Client().Expire(ctx, p.getKey(pToken.Token), time.Duration(pToken.ExpiresIn)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("parentToken.renew failed, err:%w", err)
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

func (p *parentToken) getUid(ctx context.Context, pToken string) (uid int64, err error) {
	// 根据父节点token，获取用户信息
	uid, err = p.redis.Client().HGet(ctx, p.getKey(pToken), p.hashKeyUid).Int64()
	if err != nil {
		err = fmt.Errorf("parentToken getUid failed, err: %w", err)
		return
	}
	return
}

func (p *parentToken) setToken(ctx context.Context, pToken string, clientId string, token dto.Token) error {
	expireTimeList, err := p.getExpireTimeList(ctx, pToken)
	if err != nil {
		return err
	}
	// 如果不存在key，报错
	_, err = p.redis.HGet(ctx, p.getKey(pToken), p.hashKeyCtime)
	if err != nil {
		return fmt.Errorf("parentToken.createToken get key empty, err: %w", err)
	}

	nowTime := time.Now().Unix()
	newExpireTimeList := make(uidTokenExpires, 0)
	// 新数据添加到队列前面，这样方便后续清除数据，或者对数据做一些限制
	newExpireTimeList = append(newExpireTimeList, uidTokenExpire{
		Token:      clientId,
		ExpireTime: nowTime + token.ExpiresIn,
	})

	// 删除过期的数据
	hdelFields := make([]string, 0)
	for _, value := range expireTimeList {
		// 过期时间小于当前时间，那么需要删除
		if value.ExpireTime <= nowTime {
			hdelFields = append(hdelFields, value.Token)
			continue
		}
		newExpireTimeList = append(newExpireTimeList, value)
	}
	if len(hdelFields) > 0 {
		err = p.redis.HDel(ctx, p.getKey(pToken), hdelFields...)
		if err != nil {
			return fmt.Errorf("uidMapParentToken setToken HDel expire data failed, error: %w", err)
		}
	}

	err = p.redis.HSet(ctx, p.getKey(pToken), p.hashExpireTimeList, newExpireTimeList.Marshal())
	if err != nil {
		return fmt.Errorf("uidMapParentToken setToken HSet expire time failed, error: %w", err)
	}

	tokenJsonInfo, err := token.Marshal()
	if err != nil {
		return fmt.Errorf("parentToken.createToken json marshal failed, err: %w", err)
	}

	err = p.redis.HSet(ctx, p.getKey(pToken), clientId, tokenJsonInfo)
	if err != nil {
		return fmt.Errorf("parentToken.createToken hset failed, err:%w", err)
	}
	return nil
}

// 获取过期时间，最新的在最前面。
func (p *parentToken) getExpireTimeList(ctx context.Context, pToken string) (userInfo uidTokenExpires, err error) {
	// 根据父节点token，获取用户信息
	infoBytes, err := p.redis.Client().HGet(ctx, p.getKey(pToken), p.hashExpireTimeList).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		err = fmt.Errorf("parentToken getExpireTimeList failed, err: %w", err)
		return
	}
	if errors.Is(err, redis.Nil) {
		err = nil
		return
	}

	pUserInfo := &userInfo
	err = pUserInfo.Unmarshal(infoBytes)
	if err != nil {
		err = fmt.Errorf("parentToken getExpireTimeList json unmarshal error, %w", err)
		return
	}
	return
}

func (p *parentToken) getToken(ctx context.Context, pToken string, clientId string) (tokenInfo dto.Token, err error) {
	tokenValue, err := p.redis.HGet(ctx, p.getKey(pToken), clientId)
	if err != nil {
		err = fmt.Errorf("tokgen get redis hmget string error, %w", err)
		return
	}
	pTokenInfo := &tokenInfo
	err = pTokenInfo.Unmarshal([]byte(tokenValue))
	if err != nil {
		err = fmt.Errorf("redis token info json unmarshal errorr, err: %w", err)
		return
	}
	return
}
