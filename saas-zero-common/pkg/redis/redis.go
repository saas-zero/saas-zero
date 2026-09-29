package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
	gzredis "github.com/zeromicro/go-zero/core/stores/redis"
)

// ErrNotInitialized 表示 Redis 客户端未完成初始化（nil 指针防护）。
// 生产代码在 serviceContext 已做 fail-closed，此错误仅在测试等场景出现。
var ErrNotInitialized = errors.New("redis client not initialized")

type Conf struct {
	Host string
	Pass string
	Type string `json:",default=node,options=node|cluster"`
	DB   int    `json:",default=0"`
}

type Client struct {
	gz *gzredis.Redis
	gr *goredis.Client
}

func NewClient(c Conf) (*Client, error) {
	if c.DB == 0 {
		rds, err := gzredis.NewRedis(gzredis.RedisConf{
			Host: c.Host,
			Pass: c.Pass,
			Type: c.Type,
		})
		if err != nil {
			return nil, err
		}
		return &Client{gz: rds}, nil
	}
	gr := goredis.NewClient(&goredis.Options{
		Addr:     c.Host,
		Password: c.Pass,
		DB:       c.DB,
	})
	return &Client{gr: gr}, nil
}

func (c *Client) Get(key string) (string, error) {
	if c == nil {
		return "", ErrNotInitialized
	}
	if c.gz != nil {
		return c.gz.Get(key)
	}
	if c.gr != nil {
		return c.gr.Get(context.Background(), key).Result()
	}
	return "", ErrNotInitialized
}

func (c *Client) Setex(key, value string, seconds int) error {
	if c == nil {
		return ErrNotInitialized
	}
	if c.gz != nil {
		return c.gz.Setex(key, value, seconds)
	}
	if c.gr != nil {
		return c.gr.Set(context.Background(), key, value, time.Duration(seconds)*time.Second).Err()
	}
	return ErrNotInitialized
}

func (c *Client) Exists(key string) (bool, error) {
	if c == nil {
		return false, ErrNotInitialized
	}
	if c.gz != nil {
		return c.gz.Exists(key)
	}
	if c.gr != nil {
		n, err := c.gr.Exists(context.Background(), key).Result()
		if err != nil {
			return false, err
		}
		return n > 0, nil
	}
	return false, ErrNotInitialized
}

func (c *Client) Incr(key string) (int64, error) {
	if c == nil {
		return 0, ErrNotInitialized
	}
	if c.gz != nil {
		return c.gz.Incr(key)
	}
	if c.gr != nil {
		return c.gr.Incr(context.Background(), key).Result()
	}
	return 0, ErrNotInitialized
}

func (c *Client) Del(key string) (int, error) {
	if c == nil {
		return 0, ErrNotInitialized
	}
	if c.gz != nil {
		return c.gz.Del(key)
	}
	if c.gr != nil {
		n, err := c.gr.Del(context.Background(), key).Result()
		return int(n), err
	}
	return 0, ErrNotInitialized
}

func (c *Client) Expire(key string, seconds int) error {
	if c == nil {
		return ErrNotInitialized
	}
	if c.gz != nil {
		return c.gz.Expire(key, seconds)
	}
	if c.gr != nil {
		return c.gr.Expire(context.Background(), key, time.Duration(seconds)*time.Second).Err()
	}
	return ErrNotInitialized
}

// SetNX 仅当 key 不存在时设置成功（用于幂等锁）。
// 返回 ok=true 表示本调用获得锁。
func (c *Client) SetNX(key, value string, seconds int) (bool, error) {
	if c == nil {
		return false, ErrNotInitialized
	}
	if c.gz != nil {
		return c.gz.SetnxEx(key, value, seconds)
	}
	if c.gr != nil {
		ok, err := c.gr.SetNX(context.Background(), key, value, time.Duration(seconds)*time.Second).Result()
		return ok, err
	}
	return false, ErrNotInitialized
}
