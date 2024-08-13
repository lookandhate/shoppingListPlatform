package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

type handler func(ctx context.Context, conn redis.Conn) error
type Client struct {
	pool              *redis.Pool
	connectionTimeout time.Duration
}

func NewClient(pool *redis.Pool, connectionTimeout time.Duration) *Client {
	return &Client{
		pool:              pool,
		connectionTimeout: connectionTimeout,
	}
}

func (c *Client) HashSet(ctx context.Context, key string, values interface{}) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		args := redis.Args{key}.AddFlat(values)
		_, txErr := conn.Do("HSET", args...)
		return txErr
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, txErr := conn.Do("SET", redis.Args{key}.Add(value)...)
		return txErr
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) HGetAll(ctx context.Context, key string) ([]interface{}, error) {
	var results []interface{}
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var txErr error
		results, txErr = redis.Values(conn.Do("HGETALL", key))
		return txErr
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (c *Client) Get(ctx context.Context, key string) (interface{}, error) {
	var result interface{}
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var txErr error
		result, txErr = conn.Do("GET", key)
		return txErr
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("EXPIRE", key, expiration)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Del(ctx context.Context, key string) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("DEL", key)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("PING")
		return err
	})
	if err != nil {
		return err
	}

	return nil
}
func (c *Client) execute(ctx context.Context, handler handler) error {
	connect, err := c.getConnect(ctx)
	if err != nil {
		return err
	}

	defer func() {
		txErr := connect.Close()
		if txErr != nil {
			log.Printf("Failed to close redis connection")
		}
	}()

	err = handler(ctx, connect)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) getConnect(ctx context.Context) (redis.Conn, error) {
	getConnTimeoutCtx, cancel := context.WithTimeout(ctx, c.connectionTimeout)
	defer cancel()

	conn, err := c.pool.GetContext(getConnTimeoutCtx)
	if err != nil {
		_ = conn.Close()
		fmt.Printf("\n%v error when GetContext\n", err)
		return nil, err
	}

	return conn, nil
}
