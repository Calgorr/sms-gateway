package balance

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrCacheMiss           = errors.New("balance cache miss")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

var checkAndDeductScript = redis.NewScript(checkAndDeductLua) // loaded from redis_scripts/check_and_deduct.lua via embed

func (s *RedisStore) CheckAndDeduct(ctx context.Context, customerID string, cost int64) (int64, error) {
	key := fmt.Sprintf("balance:%s", customerID)

	result, err := checkAndDeductScript.Run(ctx, s.client, []string{key}, cost).Int64()
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
