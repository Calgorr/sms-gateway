package balance

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/calgorr/sms-gateway/config"
)

type RedisStore struct {
	client *redis.Client
}

//go:embed redis_scripts/check_and_deduct.lua
var checkAndDeductLua string

//go:embed redis_scripts/add_or_set_balance.lua
var addOrSetBalanceLua string

var (
	ErrCacheMiss           = errors.New("balance cache miss")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

func NewRedisStore(client *redis.Client) CheckBalanceService {
	return &RedisStore{client: client}
}

var checkAndDeductScript = redis.NewScript(checkAndDeductLua)   // loaded from redis_scripts/check_and_deduct.lua via embed
var addOrSetBalanceScript = redis.NewScript(addOrSetBalanceLua) // loaded from redis_scripts/add_or_set_balance.lua via embed

func (s *RedisStore) CheckAndDeduct(ctx context.Context, customerID string) (int64, error) {
	key := fmt.Sprintf("balance:%s", customerID)

	result, err := checkAndDeductScript.Run(ctx, s.client, []string{key}, config.C.Opts.CostPerSms).Int64()
	if err != nil {
		return 0, fmt.Errorf("redis script exec: %w", err)
	}

	switch {
	case result == -1:
		return 0, ErrCacheMiss
	case result == -2:
		return 0, ErrInsufficientBalance
	default:
		return result, nil
	}
}

func (s *RedisStore) AddOrSetBalance(ctx context.Context, customerID string, amount int64) error {
	key := fmt.Sprintf("balance:%s", customerID)

	result, err := addOrSetBalanceScript.Run(ctx, s.client, []string{key}, amount).Text()
	if err != nil {
		return fmt.Errorf("redis script exec: %w", err)
	}
	if result == "" || result == "-1" {
		return ErrCacheMiss
	}

	return nil
}

func (s *RedisStore) SetBalance(ctx context.Context, customerID string, amount int64) error {
	key := fmt.Sprintf("balance:%s", customerID)

	err := s.client.Set(ctx, key, amount, 0).Err()
	if err != nil {
		return fmt.Errorf("set balance in redis: %w", err)
	}

	return nil
}
