package eredis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Ping
func (r *Component) Ping(ctx context.Context) (string, error) {
	return r.client.Ping(ctx).Result()
}

// Get
func (r *Component) Get(ctx context.Context, key string) (string, error) {
	reply, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis get string error %w", err)
	}
	return reply, err
}

// GETEX
func (r *Component) GetEx(ctx context.Context, key string, expire time.Duration) (string, error) {
	reply, err := r.client.GetEx(ctx, key, expire).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis get string error %w", err)
	}
	return reply, err
}

// GetBytes
func (r *Component) GetBytes(ctx context.Context, key string) ([]byte, error) {
	c, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return c, fmt.Errorf("eredis get bytes error %w", err)
	}
	return c, nil
}

// MGet ...
func (r *Component) MGetString(ctx context.Context, keys ...string) ([]string, error) {
	reply, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return []string{}, fmt.Errorf("eredis mgetstring error %w", err)
	}
	strSlice := make([]string, 0, len(reply))
	for _, v := range reply {
		if v != nil {
			strSlice = append(strSlice, v.(string))
		} else {
			strSlice = append(strSlice, "")
		}
	}
	return strSlice, nil
}

// MGets ...
func (r *Component) MGet(ctx context.Context, keys []string) ([]interface{}, error) {
	return r.client.MGet(ctx, keys...).Result()
}

// Set 设置redis的string
func (r *Component) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return r.client.Set(ctx, key, value, expire).Err()
}

// SetEX ...
func (r *Component) SetEX(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return r.client.SetEX(ctx, key, value, expire).Err()
}

// SetNX ...
func (r *Component) SetNX(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return r.client.SetNX(ctx, key, value, expire).Err()
}

// HGetAll 从redis获取hash的所有键值对
func (r *Component) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HGet 从redis获取hash单个值
func (r *Component) HGet(ctx context.Context, key string, fields string) (string, error) {
	return r.client.HGet(ctx, key, fields).Result()
}

// HMGetMap 批量获取hash值，返回map
func (r *Component) HMGetMap(ctx context.Context, key string, fields []string) (map[string]string, error) {
	if len(fields) == 0 {
		return make(map[string]string), fmt.Errorf("eredis hmgetmap error %w", ErrInvalidParams)
	}
	reply, err := r.client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return make(map[string]string), fmt.Errorf("eredis hmgetmap error %w", err)
	}

	hashRet := make(map[string]string, len(reply))
	var tmpTagID string

	for k, v := range reply {
		tmpTagID = fields[k]
		if v != nil {
			hashRet[tmpTagID] = v.(string)
		} else {
			hashRet[tmpTagID] = ""
		}
	}
	return hashRet, nil
}

// HMSet 设置redis的hash
func (r *Component) HMSet(ctx context.Context, key string, hash map[string]interface{}, expire time.Duration) error {
	if len(hash) == 0 {
		return fmt.Errorf("eredis hmset error %w", ErrInvalidParams)
	}

	err := r.client.HMSet(ctx, key, hash).Err()
	if err != nil {
		return err
	}
	if expire > 0 {
		err = r.client.Expire(ctx, key, expire).Err()
		if err != nil {
			return fmt.Errorf("eredis hmset expire error %w", err)
		}
	}
	return nil
}

// HSet hset
func (r *Component) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

// HDel ...
func (r *Component) HDel(ctx context.Context, key string, field ...string) error {
	return r.client.HDel(ctx, key, field...).Err()
}

// SetNx 设置redis的string 如果键已存在
func (r *Component) SetNx(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Incr redis自增
func (r *Component) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrBy 将 key 所储存的值加上增量 increment 。
func (r *Component) IncrBy(ctx context.Context, key string, increment int64) (int64, error) {
	return r.client.IncrBy(ctx, key, increment).Result()
}

// Decr redis自减
func (r *Component) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// Decr redis自减特定的值
func (r *Component) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return r.client.DecrBy(ctx, key, decrement).Result()
}

// Type ...
func (r *Component) Type(ctx context.Context, key string) (string, error) {
	return r.client.Type(ctx, key).Result()
}

// ZRevRange 倒序获取有序集合的部分数据
func (r *Component) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRevRange(ctx, key, start, stop).Result()
}

// ZRevRangeWithScores ...
func (r *Component) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return r.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
}

// ZRange ...
func (r *Component) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeByScore ...
func (r *Component) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, opt).Result()
}

// ZRangeWithScores ...
func (r *Component) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return r.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRangeByScoreWithScores ...
func (r *Component) ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	return r.client.ZRangeByScoreWithScores(ctx, key, opt).Result()
}

// ZRevRank ...
func (r *Component) ZRevRank(ctx context.Context, key string, member string) (int64, error) {
	return r.client.ZRevRank(ctx, key, member).Result()
}

// ZRevRangeByScore ...
func (r *Component) ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return r.client.ZRevRangeByScore(ctx, key, opt).Result()
}

// ZRevRangeByScoreWithScores ...
func (r *Component) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	return r.client.ZRevRangeByScoreWithScores(ctx, key, opt).Result()
}

// HMGet 批量获取hash值
func (r *Component) HMGetString(ctx context.Context, key string, fileds []string) ([]string, error) {
	reply, err := r.client.HMGet(ctx, key, fileds...).Result()
	if err != nil {
		return []string{}, fmt.Errorf("hmgetstring err %w", err)
	}
	strSlice := make([]string, 0, len(reply))
	for _, v := range reply {
		if v != nil {
			strSlice = append(strSlice, v.(string))
		} else {
			strSlice = append(strSlice, "")
		}
	}
	return strSlice, nil
}

func (r *Component) HMGet(ctx context.Context, key string, fileds []string) ([]interface{}, error) {
	return r.client.HMGet(ctx, key, fileds...).Result()
}

// ZCard 获取有序集合的基数
func (r *Component) ZCard(ctx context.Context, key string) (int64, error) {
	return r.client.ZCard(ctx, key).Result()
}

// ZScore 获取有序集合成员 member 的 score 值
func (r *Component) ZScore(ctx context.Context, key string, member string) (float64, error) {
	return r.client.ZScore(ctx, key, member).Result()
}

// ZAdd 将一个或多个 member 元素及其 score 值加入到有序集 key 当中
func (r *Component) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	return r.client.ZAdd(ctx, key, members...).Result()
}

// ZCount 返回有序集 key 中， score 值在 min 和 max 之间(默认包括 score 值等于 min 或 max )的成员的数量。
func (r *Component) ZCount(ctx context.Context, key string, min, max string) (int64, error) {
	return r.client.ZCount(ctx, key, min, max).Result()
}

// Del redis删除
func (r *Component) Del(ctx context.Context, key string) (int64, error) {
	return r.client.Del(ctx, key).Result()
}

// HIncrBy 哈希field自增
func (r *Component) HIncrBy(ctx context.Context, key string, field string, incr int) (int64, error) {
	return r.client.HIncrBy(ctx, key, field, int64(incr)).Result()
}

// Exists 键是否存在
func (r *Component) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return result == 1, err
	}
	return result == 1, nil
}

// LPush 将一个或多个值 value 插入到列表 key 的表头
func (r *Component) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.client.LPush(ctx, key, values...).Result()
}

// RPush 将一个或多个值 value 插入到列表 key 的表尾(最右边)。
func (r *Component) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.client.RPush(ctx, key, values...).Result()
}

// RPop 移除并返回列表 key 的尾元素。
func (r *Component) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LRange 获取列表指定范围内的元素
func (r *Component) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

// LLen ...
func (r *Component) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// LRem ...
func (r *Component) LRem(ctx context.Context, key string, count int64, value interface{}) (int64, error) {
	return r.client.LRem(ctx, key, count, value).Result()
}

// LIndex ...
func (r *Component) LIndex(ctx context.Context, key string, idx int64) (string, error) {
	return r.client.LIndex(ctx, key, idx).Result()
}

// LTrim ...
func (r *Component) LTrim(ctx context.Context, key string, start, stop int64) (string, error) {
	return r.client.LTrim(ctx, key, start, stop).Result()
}

// ZRemRangeByRank 移除有序集合中给定的排名区间的所有成员
func (r *Component) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	return r.client.ZRemRangeByRank(ctx, key, start, stop).Result()
}

// Expire 设置过期时间
func (r *Component) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return r.client.Expire(ctx, key, expiration).Result()
}

// ZRem 从zset中移除变量
func (r *Component) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.client.ZRem(ctx, key, members...).Result()
}

// SAdd 向set中添加成员
func (r *Component) SAdd(ctx context.Context, key string, member ...interface{}) (int64, error) {
	return r.client.SAdd(ctx, key, member...).Result()
}

// SMembers 返回set的全部成员
func (r *Component) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SIsMember ...
func (r *Component) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// SRem ...
func (r *Component) SRem(ctx context.Context, key string, member interface{}) (int64, error) {
	return r.client.SRem(ctx, key, member).Result()
}

// HKeys 获取hash的所有域
func (r *Component) HKeys(ctx context.Context, key string) ([]string, error) {
	return r.client.HKeys(ctx, key).Result()
}

// HLen 获取hash的长度
func (r *Component) HLen(ctx context.Context, key string) (int64, error) {
	return r.client.HLen(ctx, key).Result()
}

// GeoAdd 写入地理位置
func (r *Component) GeoAdd(ctx context.Context, key string, location *redis.GeoLocation) (int64, error) {
	return r.client.GeoAdd(ctx, key, location).Result()
}

// GeoRadius 根据经纬度查询列表
func (r *Component) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	return r.client.GeoRadius(ctx, key, longitude, latitude, query).Result()
}

// TTL 查询过期时间
func (r *Component) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Close closes the cluster client, releasing any open resources.
//
// It is rare to Close a ClusterClient, as the ClusterClient is meant
// to be long-lived and shared between many goroutines.
func (r *Component) Close() (err error) {
	err = nil
	if r.client != nil {
		if r.Cluster() != nil {
			err = r.Cluster().Close()
			if err != nil {
				err = fmt.Errorf("cluster close err %w", err)
			}
		}

		if r.Stub() != nil {
			err = r.Stub().Close()
			if err != nil {
				err = fmt.Errorf("stub close err %w", err)
			}
		}
	}
	return err
}
