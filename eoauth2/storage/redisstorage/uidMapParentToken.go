package redisstorage

import (
	"github.com/vmihailenco/msgpack"
)

// uidTokenExpire 用户Uid里parent token存储的一些信息
type uidTokenExpire struct {
	Token      string `msgpack:"t"`
	ExpireTime int64  `msgpack:"et"`
}

type uidTokenExpires []uidTokenExpire

func (u uidTokenExpires) Marshal() []byte {
	info, _ := msgpack.Marshal(u)
	return info
}

func (u *uidTokenExpires) Unmarshal(content []byte) error {
	return msgpack.Unmarshal(content, u)
}
