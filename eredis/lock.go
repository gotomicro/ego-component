package eredis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/go-redis/redis/v8"
	"io"
	"strconv"
	"sync"
	"time"
)

var (
	luaRefresh = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`)
	luaRelease = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`)
	luaPTTL    = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pttl", KEYS[1]) else return -3 end`)
)

// lockClient wraps a redis client.
type lockClient struct {
	client redis.Cmdable
	tmp    []byte
	tmpMu  sync.Mutex
}

// Obtain tries to obtain a new Lock using a key with the given TTL.
// May return ErrNotObtained if not successful.
func (c *lockClient) Obtain(ctx context.Context, key string, ttl time.Duration, opts ...LockOption) (*Lock, error) {
	// Create a random token
	token, err := c.randomToken()
	if err != nil {
		return nil, err
	}
	opt := &lockOption{}
	for _, o := range opts {
		o(opt)
	}
	if opt.retryStrategy == nil {
		opt.retryStrategy = NoRetry()
	}

	value := token + opt.metadata
	retry := opt.retryStrategy

	deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(ttl))
	defer cancel()

	var timer *time.Timer
	for {
		ok, err := c.obtain(deadlineCtx, key, value, ttl)
		if err != nil {
			return nil, err
		} else if ok {
			return &Lock{client: c, key: key, value: value}, nil
		}

		backoff := retry.NextBackoff()
		if backoff < 1 {
			return nil, ErrNotObtained
		}

		if timer == nil {
			timer = time.NewTimer(backoff)
			defer timer.Stop()
		} else {
			timer.Reset(backoff)
		}

		select {
		case <-deadlineCtx.Done():
			return nil, ErrNotObtained
		case <-timer.C:
		}
	}
}

func (c *lockClient) obtain(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, ttl).Result()
}

func (c *lockClient) randomToken() (string, error) {
	c.tmpMu.Lock()
	defer c.tmpMu.Unlock()

	if len(c.tmp) == 0 {
		c.tmp = make([]byte, 16)
	}

	if _, err := io.ReadFull(rand.Reader, c.tmp); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(c.tmp), nil
}

// Lock represents an obtained, distributed Lock.
type Lock struct {
	client *lockClient
	key    string
	value  string
}

// Key returns the redis key used by the Lock.
func (l *Lock) Key() string {
	return l.key
}

// Token returns the token value set by the Lock.
func (l *Lock) Token() string {
	return l.value[:22]
}

// Metadata returns the metadata of the Lock.
func (l *Lock) Metadata() string {
	return l.value[22:]
}

// TTL returns the remaining time-to-live. Returns 0 if the Lock has expired.
func (l *Lock) TTL(ctx context.Context) (time.Duration, error) {
	res, err := luaPTTL.Run(ctx, l.client.client, []string{l.key}, l.value).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	if num := res.(int64); num > 0 {
		return time.Duration(num) * time.Millisecond, nil
	}
	return 0, nil
}

// Refresh extends the Lock with a new TTL.
// May return ErrNotObtained if refresh is unsuccessful.
func (l *Lock) Refresh(ctx context.Context, ttl time.Duration, opts ...LockOption) error {
	ttlVal := strconv.FormatInt(int64(ttl/time.Millisecond), 10)
	status, err := luaRefresh.Run(ctx, l.client.client, []string{l.key}, l.value, ttlVal).Result()
	if err != nil {
		return err
	} else if status == int64(1) {
		return nil
	}
	return ErrNotObtained
}

// Release manually releases the Lock.
// May return ErrLockNotHeld.
func (l *Lock) Release(ctx context.Context) error {
	res, err := luaRelease.Run(ctx, l.client.client, []string{l.key}, l.value).Result()
	if err == redis.Nil {
		return ErrLockNotHeld
	} else if err != nil {
		return err
	}

	if i, ok := res.(int64); !ok || i != 1 {
		return ErrLockNotHeld
	}
	return nil
}

type LockOption func(c *lockOption)

// Options describe the options for the Lock
type lockOption struct {
	// retryStrategy allows to customise the Lock retry strategy.
	// Default: do not retry
	retryStrategy RetryStrategy

	// metadata string is appended to the Lock token.
	metadata string
}

func WithLockOptionMetadata(md string) LockOption {
	return func(lo *lockOption) {
		lo.metadata = md
	}
}

func WithLockOptionRetryStrategy(retryStrategy RetryStrategy) LockOption {
	return func(lo *lockOption) {
		lo.retryStrategy = retryStrategy
	}
}
