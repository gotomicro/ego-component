package etoken

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/gotomicro/ego/core/elog"
)

const tokenKeyPattern = "/token/%d"

type Component struct {
	config *config
	client *redis.Client
	logger *elog.Component
}

func newComponent(cfg *config, client *redis.Client, logger *elog.Component) *Component {
	return &Component{
		config: cfg,
		client: client,
		logger: logger,
	}
}

type AccessTokenTicket struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

func (c *Component) CreateAccessToken(uid int, startTime int64) (resp AccessTokenTicket, err error) {
	// using the uid as the jwtId
	tokenString, err := c.EncodeAccessToken(uid, uid, startTime)
	if err != nil {
		return
	}

	err = c.client.Set(context.Background(), fmt.Sprintf(c.config.TokenPrefix+tokenKeyPattern, uid), tokenString,
		time.Duration(c.config.AccessTokenExpireInterval)*time.Second).Err()
	if err != nil {
		return AccessTokenTicket{}, fmt.Errorf("set token error %v", err)
	}
	resp.AccessToken = tokenString
	resp.ExpiresIn = c.config.AccessTokenExpireInterval
	return
}

func (c *Component) CheckAccessToken(tokenStr string) bool {
	sc, err := c.DecodeAccessToken(tokenStr)
	if err != nil {
		return false
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	err = c.client.Get(context.Background(), fmt.Sprintf(c.config.TokenPrefix+tokenKeyPattern, uidInt)).Err()
	if err != nil {
		return false
	}
	return true
}

func (c *Component) RefreshAccessToken(tokenStr string, startTime int64) (resp AccessTokenTicket, err error) {
	sc, err := c.DecodeAccessToken(tokenStr)
	if err != nil {
		return
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	return c.CreateAccessToken(uidInt, startTime)
}

func (c *Component) EncodeAccessToken(jwtId int, uid int, startTime int64) (tokenStr string, err error) {
	jwtToken := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["jti"] = jwtId
	claims["iss"] = c.config.AccessTokenIss
	claims["sub"] = uid
	claims["iat"] = startTime
	claims["exp"] = startTime + c.config.AccessTokenExpireInterval
	jwtToken.Claims = claims
	tokenStr, err = jwtToken.SignedString([]byte(c.config.AccessTokenKey))
	if err != nil {
		return
	}
	return
}

func (c *Component) DecodeAccessToken(tokenStr string) (resp map[string]interface{}, err error) {
	tokenParse, err := jwt.Parse(tokenStr, func(jwtToken *jwt.Token) (interface{}, error) {
		return []byte(c.config.AccessTokenKey), nil
	})
	if err != nil {
		return
	}
	var flag bool
	resp, flag = tokenParse.Claims.(jwt.MapClaims)
	if !flag {
		err = errors.New("assert error")
		return
	}
	return
}
