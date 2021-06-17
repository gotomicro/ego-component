package dto

import (
	"encoding/base64"
	"time"

	"github.com/pborman/uuid"
)

type Token struct {
	Token     string `json:"token"`
	AuthAt    int64  `json:"auth_at"`
	ExpiresIn int64  `json:"expires_in"` // Token 多长时间后过期(s)
}

func NewToken(expiresIn int64) Token {
	return Token{
		Token:     generateToken(),
		AuthAt:    time.Now().Unix(),
		ExpiresIn: expiresIn,
	}
}

func generateToken() string {
	return base64.RawURLEncoding.EncodeToString(uuid.NewRandom())
}
