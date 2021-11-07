package main

import (
	"time"

	gecko "github.com/superoo7/go-gecko/v3"
)

type CoingeckoWrapper struct {
	client     *gecko.Client
	currency   string
	result     float32
	lastUpdate time.Time
}

func NewCoingeckoWrapper(currency string) *CoingeckoWrapper {
	if currency == "" {
		log.Info().Msg("Coingecko currency is not set, not intitializing Coingecko wrapper")
		return &CoingeckoWrapper{}
	}

	var cg = gecko.NewClient(nil)

	return &CoingeckoWrapper{
		client:   cg,
		currency: currency,
	}
}

func (c *CoingeckoWrapper) GetRate() (float32, error) {
	if c.client == nil {
		log.Trace().Msg("Coingecko wrapper not initialized, cannot fetch data.")
		return 0, nil
	}

	if !c.lastUpdate.IsZero() && time.Since(c.lastUpdate).Minutes() > 10 {
		log.Trace().
			Time("now", time.Now()).
			Time("then", c.lastUpdate).
			Dur("diff", time.Since(c.lastUpdate)).
			Msg("Using rate from cache.")
		return c.result, nil
	}

	result, err := c.client.SimpleSinglePrice(c.currency, "usd")
	if err != nil {
		log.Warn().Err(err).Msg("Could not get Coingecko exchange rate")
		return 0, err
	}

	c.result = result.MarketPrice
	c.lastUpdate = time.Now()

	return result.MarketPrice, nil
}
