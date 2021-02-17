package etoken

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/gotomicro/ego/core/elog"
)

const tokenKeyPattern = "/token/%d"

type Component struct {
	Config *Config
	client                    *redis.Client
	logger                    *elog.Component
}

func newComponent(cfg *Config, client *redis.Client, logger *elog.Component) *Component {
	return &Component{
		Config: cfg,
		client: client,
		logger: logger,
	}
}

type AccessTokenTicket struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

func (con *Component) CreateAccessToken(uid int, startTime int64) (resp AccessTokenTicket, err error) {
	// using the uid as the jwtId
	tokenString, err := con.EncodeAccessToken(uid, uid, startTime)
	if err != nil {
		return
	}

	err = con.client.Set(fmt.Sprintf(con.Config.TokenPrefix+tokenKeyPattern, uid), tokenString,
		time.Duration(con.Config.AccessTokenExpireInterval)*time.Second).Err()
	if err != nil {
		return AccessTokenTicket{}, fmt.Errorf("set token error %v", err)
	}
	resp.AccessToken = tokenString
	resp.ExpiresIn = con.Config.AccessTokenExpireInterval
	return
}

func (con *Component) CheckAccessToken(tokenStr string) bool {
	sc, err := con.DecodeAccessToken(tokenStr)
	if err != nil {
		return false
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	err = con.client.Get(fmt.Sprintf(con.Config.TokenPrefix+tokenKeyPattern, uidInt)).Err()
	if err != nil {
		return false
	}
	return true
}

func (con *Component) RefreshAccessToken(tokenStr string, startTime int64) (resp AccessTokenTicket, err error) {
	sc, err := con.DecodeAccessToken(tokenStr)
	if err != nil {
		return
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	return con.CreateAccessToken(uidInt, startTime)
}

func (con *Component) EncodeAccessToken(jwtId int, uid int, startTime int64) (tokenStr string, err error) {
	jwtToken := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["jti"] = jwtId
	claims["iss"] = con.Config.AccessTokenIss
	claims["sub"] = uid
	claims["iat"] = startTime
	claims["exp"] = startTime + con.Config.AccessTokenExpireInterval
	jwtToken.Claims = claims
	tokenStr, err = jwtToken.SignedString([]byte(con.Config.AccessTokenKey))
	if err != nil {
		return
	}
	return
}

func (con *Component) DecodeAccessToken(tokenStr string) (resp map[string]interface{}, err error) {
	tokenParse, err := jwt.Parse(tokenStr, func(jwtToken *jwt.Token) (interface{}, error) {
		return []byte(con.Config.AccessTokenKey), nil
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
