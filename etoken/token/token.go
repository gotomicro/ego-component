package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const tokenKeyPattern = "/token/%d"

type AccessTokenTicket struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

func (con *Container) CreateAccessToken(uid int, startTime int64) (resp AccessTokenTicket, err error) {
	// using the uid as the jwtId
	tokenString, err := con.EncodeAccessToken(uid, uid, startTime)
	if err != nil {
		return
	}

	err = con.client.Set(fmt.Sprintf(con.config.TokenPrefix+tokenKeyPattern, uid), tokenString,
		time.Duration(con.config.AccessTokenExpireInterval)*time.Second).Err()
	if err != nil {
		return AccessTokenTicket{}, fmt.Errorf("set token error %v", err)
	}
	resp.AccessToken = tokenString
	resp.ExpiresIn = con.config.AccessTokenExpireInterval
	return
}

func (con *Container) CheckAccessToken(tokenStr string) bool {
	sc, err := con.DecodeAccessToken(tokenStr)
	if err != nil {
		return false
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	err = con.client.Get(fmt.Sprintf(con.config.TokenPrefix+tokenKeyPattern, uidInt)).Err()
	if err != nil {
		return false
	}
	return true
}

func (con *Container) RefreshAccessToken(tokenStr string, startTime int64) (resp AccessTokenTicket, err error) {
	sc, err := con.DecodeAccessToken(tokenStr)
	if err != nil {
		return
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	return con.CreateAccessToken(uidInt, startTime)
}

func (con *Container) EncodeAccessToken(jwtId int, uid int, startTime int64) (tokenStr string, err error) {
	jwtToken := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["jti"] = jwtId
	claims["iss"] = con.config.AccessTokenIss
	claims["sub"] = uid
	claims["iat"] = startTime
	claims["exp"] = startTime + con.config.AccessTokenExpireInterval
	jwtToken.Claims = claims
	tokenStr, err = jwtToken.SignedString([]byte(con.config.AccessTokenKey))
	if err != nil {
		return
	}
	return
}

func (con *Container) DecodeAccessToken(tokenStr string) (resp map[string]interface{}, err error) {
	tokenParse, err := jwt.Parse(tokenStr, func(jwtToken *jwt.Token) (interface{}, error) {
		return []byte(con.config.AccessTokenKey), nil
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
