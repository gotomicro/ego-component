package cache

import "time"

// Cache interface
type Cache interface {
	GetString(key string) (string, error)
	Set(key string, value interface{}, expire time.Duration) error
	Exists(key string) (bool, error)
	Del(key string) (int64, error)
}
