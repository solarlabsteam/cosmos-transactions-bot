package main

import (
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Cache struct {
	Validators map[string]stakingtypes.Validator
}

type CacheManager struct {
	Cache       Cache
	GrpcWrapper GrpcWrapper
}

func NewCacheManager(grpcWrapper *GrpcWrapper) *CacheManager {
	return &CacheManager{
		Cache: Cache{
			Validators: make(map[string]stakingtypes.Validator),
		},
		GrpcWrapper: *grpcWrapper,
	}
}

func (c *CacheManager) getValidatorMaybeFromCache(address string) (stakingtypes.Validator, error) {
	if validator, found := c.Cache.Validators[address]; found {
		log.Trace().Str("address", address).Msg("Getting validator value from cache")
		return validator, nil
	}

	log.Trace().Str("address", address).Msg("No value in cache, querying for validator")

	validator, err := c.GrpcWrapper.getValidator(address)

	if err != nil {
		return stakingtypes.Validator{}, err
	}

	c.Cache.Validators[address] = validator
	return validator, nil
}

func (c *CacheManager) clearCache() {
	log.Trace().Msg("Clearing cache...")
	c.Cache.Validators = make(map[string]stakingtypes.Validator)
}
