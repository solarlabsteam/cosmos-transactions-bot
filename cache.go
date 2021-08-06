package main

import (
	"context"

	"google.golang.org/grpc"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Cache struct {
	Validators map[string]stakingtypes.Validator
}

type CacheManager struct {
	Cache    Cache
	grpcConn *grpc.ClientConn
}

func NewCacheManager(conn *grpc.ClientConn) *CacheManager {
	return &CacheManager{
		Cache: Cache{
			Validators: make(map[string]stakingtypes.Validator),
		},
		grpcConn: conn,
	}
}

func (c *CacheManager) getValidatorMaybeFromCache(address string) (stakingtypes.Validator, error) {
	if validator, found := c.Cache.Validators[address]; found {
		log.Trace().Str("address", address).Msg("Getting validator value from cache")
		return validator, nil
	}

	log.Trace().Str("address", address).Msg("No value in cache, querying for validator")

	stakingClient := stakingtypes.NewQueryClient(grpcConn)
	validatorResponse, err := stakingClient.Validator(
		context.Background(),
		&stakingtypes.QueryValidatorRequest{ValidatorAddr: address},
	)

	if err != nil {
		return stakingtypes.Validator{}, err
	}

	c.Cache.Validators[address] = validatorResponse.Validator
	return validatorResponse.Validator, nil
}

func (c *CacheManager) clearCache() {
	log.Trace().Msg("Clearing cache...")
	c.Cache.Validators = make(map[string]stakingtypes.Validator)
}
