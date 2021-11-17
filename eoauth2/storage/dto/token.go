package dto

import (
	"encoding/base64"
	"time"

	"github.com/pborman/uuid"
	"github.com/vmihailenco/msgpack"
)

type Token struct {
	Token     string `json:"token" msgpack:"t"`
	AuthAt    int64  `json:"auth_at" msgpack:"at"`
	ExpiresIn int64  `json:"expires_in" msgpack:"ex"` // Token 多长时间后过期(s)
}

func NewToken(expiresIn int64) Token {
	return Token{
		Token:     generateToken(),
		AuthAt:    time.Now().Unix(),
		ExpiresIn: expiresIn,
	}
}

func (t Token) Marshal() ([]byte, error) {
	bytes, err := msgpack.Marshal(t)
	return bytes, err
}

func (t *Token) Unmarshal(content []byte) error {
	return msgpack.Unmarshal(content, t)
}

func generateToken() string {
	return base64.RawURLEncoding.EncodeToString(uuid.NewRandom())
}
