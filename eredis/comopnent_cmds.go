package eredis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const ErrInvalidParams = Err("invalid params")

type Err string

func (e Err) Error() string { return string(e) }

// Ping
func (r *Component) Ping() (string, error) {
	reply, err := r.client.Ping().Result()
	if err != nil {
		return reply, fmt.Errorf("eredis get string error %w", err)
	}
	return reply, err
}

// GetString
func (r *Component) GetString(key string) (string, error) {
	reply, err := r.client.Get(key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis get string error %w", err)
	}
	return reply, err
}

// GetBytes
func (r *Component) GetBytes(key string) ([]byte, error) {
	c, err := r.client.Get(key).Bytes()
	if err != nil {
		return c, fmt.Errorf("eredis get bytes error %w", err)
	}
	return c, nil
}

// MGet ...
func (r *Component) MGetString(keys ...string) ([]string, error) {
	reply, err := r.client.MGet(keys...).Result()
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
func (r *Component) MGetInterface(keys []string) ([]interface{}, error) {
	reply, err := r.client.MGet(keys...).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis mgets error %w", err)
	}
	return reply, nil
}

// Set 设置redis的string
func (r *Component) Set(key string, value interface{}, expire time.Duration) error {
	err := r.client.Set(key, value, expire).Err()
	if err != nil {
		return fmt.Errorf("eredis set error %w", err)
	}
	return nil
}

// HGetAll 从redis获取hash的所有键值对
func (r *Component) HGetAll(key string) (map[string]string, error) {
	reply, err := r.client.HGetAll(key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis hgetall error %w", err)
	}
	return reply, err
}

// HGet 从redis获取hash单个值
func (r *Component) HGet(key string, fields string) (string, error) {
	reply, err := r.client.HGet(key, fields).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis hget error %w", err)
	}
	return reply, err
}

// HMGetMap 批量获取hash值，返回map
func (r *Component) HMGetMap(key string, fields []string) (map[string]string, error) {
	if len(fields) == 0 {
		return make(map[string]string), fmt.Errorf("eredis hmgetmap error %w", ErrInvalidParams)
	}
	reply, err := r.client.HMGet(key, fields...).Result()
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
func (r *Component) HMSet(key string, hash map[string]interface{}, expire time.Duration) error {
	if len(hash) == 0 {
		return fmt.Errorf("eredis hmset error %w", ErrInvalidParams)
	}

	err := r.client.HMSet(key, hash).Err()
	if err != nil {
		return fmt.Errorf("eredis hmset error %w", err)
	}
	if expire > 0 {
		err = r.client.Expire(key, expire).Err()
		if err != nil {
			return fmt.Errorf("eredis hmset expire error %w", err)
		}
	}
	return nil
}

// HSet hset
func (r *Component) HSet(key string, field string, value interface{}) error {
	err := r.client.HSet(key, field, value).Err()
	if err != nil {
		return fmt.Errorf("hset error %w", err)
	}
	return nil
}

// HDel ...
func (r *Component) HDel(key string, field ...string) error {
	err := r.client.HDel(key, field...).Err()
	if err != nil {
		return fmt.Errorf("hdel error %w", err)
	}
	return nil
}

// SetNx 设置redis的string 如果键已存在
func (r *Component) SetNx(key string, value interface{}, expiration time.Duration) (bool, error) {
	result, err := r.client.SetNX(key, value, expiration).Result()
	if err != nil {
		return result, fmt.Errorf("setnx error %w", err)
	}
	return result, nil
}

// Incr redis自增
func (r *Component) Incr(key string) (int64, error) {
	reply, err := r.client.Incr(key).Result()
	if err != nil {
		return reply, fmt.Errorf("incr error %w", err)
	}
	return reply, nil
}

// IncrBy 将 key 所储存的值加上增量 increment 。
func (r *Component) IncrBy(key string, increment int64) (int64, error) {
	reply, err := r.client.IncrBy(key, increment).Result()
	if err != nil {
		return reply, fmt.Errorf("incr by error %w", err)
	}
	return reply, nil
}

// Decr redis自减
func (r *Component) Decr(key string) (int64, error) {
	reply, err := r.client.Decr(key).Result()
	if err != nil {
		return reply, fmt.Errorf("decr by error %w", err)
	}
	return reply, nil
}

// Type ...
func (r *Component) Type(key string) (string, error) {
	reply, err := r.client.Type(key).Result()
	if err != nil {
		return reply, fmt.Errorf("type error %w", err)
	}
	return reply, nil
}

// ZRevRange 倒序获取有序集合的部分数据
func (r *Component) ZRevRange(key string, start, stop int64) ([]string, error) {
	reply, err := r.client.ZRevRange(key, start, stop).Result()
	if err != nil {
		return reply, fmt.Errorf("zrevrange error %w", err)
	}
	return reply, nil
}

// ZRevRangeWithScores ...
func (r *Component) ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	reply, err := r.client.ZRevRangeWithScores(key, start, stop).Result()
	if err != nil {
		return reply, fmt.Errorf("zrevrangewithscores error %w", err)
	}
	return reply, nil
}

// ZRange ...
func (r *Component) ZRange(key string, start, stop int64) ([]string, error) {
	reply, err := r.client.ZRange(key, start, stop).Result()
	if err != nil {
		return reply, fmt.Errorf("zrange error %w", err)
	}
	return reply, nil
}

// ZRevRank ...
func (r *Component) ZRevRank(key string, member string) (int64, error) {
	reply, err := r.client.ZRevRank(key, member).Result()
	if err != nil {
		return reply, fmt.Errorf("zrevrank error %w", err)
	}
	return reply, nil
}

// ZRevRangeByScore ...
func (r *Component) ZRevRangeByScore(key string, opt redis.ZRangeBy) ([]string, error) {
	reply, err := r.client.ZRevRangeByScore(key, opt).Result()
	if err != nil {
		return reply, fmt.Errorf("zrevrangebyscore error %w", err)
	}
	return reply, nil
}

// ZRevRangeByScoreWithScores ...
func (r *Component) ZRevRangeByScoreWithScores(key string, opt redis.ZRangeBy) ([]redis.Z, error) {
	reply, err := r.client.ZRevRangeByScoreWithScores(key, opt).Result()
	if err != nil {
		return reply, fmt.Errorf("zrevrangebyscorewithscores error %w", err)
	}
	return reply, nil
}

// HMGet 批量获取hash值
func (r *Component) HMGetString(key string, fileds []string) ([]string, error) {
	reply, err := r.client.HMGet(key, fileds...).Result()
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

func (r *Component) HMGetInterface(key string, fileds []string) ([]interface{}, error) {
	reply, err := r.client.HMGet(key, fileds...).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis hmgetinterface err %w", err)
	}
	return reply, nil
}

// ZCard 获取有序集合的基数
func (r *Component) ZCard(key string) (int64, error) {
	reply, err := r.client.ZCard(key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis zcard err %w", err)
	}
	return reply, err
}

// ZScore 获取有序集合成员 member 的 score 值
func (r *Component) ZScore(key string, member string) (float64, error) {
	reply, err := r.client.ZScore(key, member).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis zscore err %w", err)
	}
	return reply, nil
}

// ZAdd 将一个或多个 member 元素及其 score 值加入到有序集 key 当中
func (r *Component) ZAdd(key string, members ...redis.Z) (int64, error) {
	reply, err := r.client.ZAdd(key, members...).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis zadd err %w", err)
	}
	return reply, nil
}

// ZCount 返回有序集 key 中， score 值在 min 和 max 之间(默认包括 score 值等于 min 或 max )的成员的数量。
func (r *Component) ZCount(key string, min, max string) (int64, error) {
	reply, err := r.client.ZCount(key, min, max).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis zcount err %w", err)
	}
	return reply, nil
}

// Del redis删除
func (r *Component) Del(key string) (int64, error) {
	result, err := r.client.Del(key).Result()
	if err != nil {
		return result, fmt.Errorf("eredis del err %w", err)
	}
	return result, err
}

// HIncrBy 哈希field自增
func (r *Component) HIncrBy(key string, field string, incr int) (int64, error) {
	result, err := r.client.HIncrBy(key, field, int64(incr)).Result()
	if err != nil {
		return result, fmt.Errorf("eredis hincrby err %w", err)
	}
	return result, nil
}

// Exists 键是否存在
func (r *Component) Exists(key string) (bool, error) {
	result, err := r.client.Exists(key).Result()
	if err != nil {
		return result == 1, fmt.Errorf("eredis err %w", err)
	}
	return result == 1, nil
}

// LPush 将一个或多个值 value 插入到列表 key 的表头
func (r *Component) LPush(key string, values ...interface{}) (int64, error) {
	reply, err := r.client.LPush(key, values...).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis lpush err %w", err)
	}
	return reply, nil
}

// RPush 将一个或多个值 value 插入到列表 key 的表尾(最右边)。
func (r *Component) RPush(key string, values ...interface{}) (int64, error) {
	reply, err := r.client.RPush(key, values...).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis rpush err %w", err)
	}
	return reply, nil
}

// RPop 移除并返回列表 key 的尾元素。
func (r *Component) RPop(key string) (string, error) {
	reply, err := r.client.RPop(key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis rpop err %w", err)
	}
	return reply, nil
}

// LRange 获取列表指定范围内的元素
func (r *Component) LRange(key string, start, stop int64) ([]string, error) {
	reply, err := r.client.LRange(key, start, stop).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis lrange err %w", err)
	}
	return reply, nil
}

// LLen ...
func (r *Component) LLen(key string) (int64, error) {
	reply, err := r.client.LLen(key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis llen err %w", err)
	}
	return reply, nil
}

// LRem ...
func (r *Component) LRem(key string, count int64, value interface{}) (int64, error) {
	reply, err := r.client.LRem(key, count, value).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis lrem err %w", err)
	}
	return reply, nil
}

// LIndex ...
func (r *Component) LIndex(key string, idx int64) (string, error) {
	reply, err := r.client.LIndex(key, idx).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis lindex err %w", err)
	}
	return reply, nil
}

// LTrim ...
func (r *Component) LTrim(key string, start, stop int64) (string, error) {
	reply, err := r.client.LTrim(key, start, stop).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis ltrim err %w", err)
	}
	return reply, nil
}

// ZRemRangeByRank 移除有序集合中给定的排名区间的所有成员
func (r *Component) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	reply, err := r.client.ZRemRangeByRank(key, start, stop).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis zremrangebyrank err %w", err)
	}
	return reply, nil
}

// Expire 设置过期时间
func (r *Component) Expire(key string, expiration time.Duration) (bool, error) {
	reply, err := r.client.Expire(key, expiration).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis expire err %w", err)
	}
	return reply, nil
}

// ZRem 从zset中移除变量
func (r *Component) ZRem(key string, members ...interface{}) (int64, error) {
	reply, err := r.client.ZRem(key, members...).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis zrem err %w", err)
	}
	return reply, nil
}

// SAdd 向set中添加成员
func (r *Component) SAdd(key string, member ...interface{}) (int64, error) {
	reply, err := r.client.SAdd(key, member...).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis sadd err %w", err)
	}
	return reply, nil
}

// SMembers 返回set的全部成员
func (r *Component) SMembers(key string) ([]string, error) {
	reply, err := r.client.SMembers(key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis sadd err %w", err)
	}
	return reply, err
}

// SIsMember ...
func (r *Component) SIsMember(key string, member interface{}) (bool, error) {
	reply, err := r.client.SIsMember(key, member).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis sismember err %w", err)
	}
	return reply, nil
}

// HKeys 获取hash的所有域
func (r *Component) HKeys(key string) ([]string, error) {
	reply, err := r.client.HKeys(key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis hkeys err %w", err)
	}
	return reply, nil
}

// HLen 获取hash的长度
func (r *Component) HLen(key string) (int64, error) {
	reply, err := r.client.HLen(key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis hlen err %w", err)
	}
	return reply, nil
}

// GeoAdd 写入地理位置
func (r *Component) GeoAdd(key string, location *redis.GeoLocation) (int64, error) {
	reply, err := r.client.GeoAdd(key, location).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis geoadd err %w", err)
	}
	return reply, nil
}

// GeoRadius 根据经纬度查询列表
func (r *Component) GeoRadius(key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	reply, err := r.client.GeoRadius(key, longitude, latitude, query).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis geo radius err %w", err)
	}
	return reply, nil

}

// TTL 查询过期时间
func (r *Component) TTL(key string) (time.Duration, error) {
	reply, err := r.client.TTL(key).Result()
	if err != nil {
		return reply, fmt.Errorf("eredis ttl err %w", err)
	}
	return reply, nil
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
