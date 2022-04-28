package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Msg interface {
	Serialize(Serializer Serializer) string
	Empty() bool
}

type Serializer struct {
	LinksSerializer         func(string, string) string
	StrongSerializer        func(string) string
	CodeSerializer          func(string) string
	MultilineCodeSerializer func(string) string
	CacheManager            *CacheManager
}

type Report struct {
	Tx   Tx
	Msgs []Msg
}

func (r Report) Empty() bool {
	return r.Tx.Hash == "" || len(r.Msgs) == 0
}

type Reporter interface {
	Serialize(Report) string
	Init()
	Enabled() bool
	SendReport(Report) error
	Name() string
	Serializer() Serializer
}

func (s Serializer) getWalletWithLabel(address string) string {
	label, labelFound := labelsConfigManager.getWalletLabel(address)

	var sb strings.Builder

	sb.WriteString(s.LinksSerializer(makeMintscanAccountLink(address), address))

	if labelFound {
		sb.WriteString(fmt.Sprintf(" (%s)", s.CodeSerializer(label)))
	}

	return sb.String()
}

func (s Serializer) getSingleOrMultilineCodeBlock(block string) string {
	if strings.Contains(block, "\n") {
		return "\n" + s.MultilineCodeSerializer(block)
	}

	return s.CodeSerializer(block)
}

func (s Serializer) getValidatorWithName(address string) string {
	var sb strings.Builder

	sb.WriteString(s.LinksSerializer(makeMintscanValidatorLink(address), address))

	if validator, err := s.CacheManager.getValidatorMaybeFromCache(address); err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Could not load delegate validator info")
	} else {
		sb.WriteString(fmt.Sprintf(" (%s)", s.CodeSerializer(validator.Description.Moniker)))
	}

	return sb.String()
}

func (s Serializer) getTokensMaybeWithDollarPrice(amount float64, denom string) string {
	rate, err := s.CacheManager.getRate()

	if err != nil || rate == 0 {
		return s.getTokensFormatted(amount, denom)
	}

	return s.CodeSerializer(Printer.Sprintf(
		"%.6f %s ($%.3f)",
		amount,
		denom,
		rate*amount,
	))
}

func (s Serializer) getTokensFormatted(amount float64, denom string) string {
	return s.CodeSerializer(Printer.Sprintf(
		"%.6f %s",
		amount,
		denom,
	))
}

func (s Serializer) getValidatorCommissionAtBlock(address string, block int64) string {
	var sb strings.Builder

	if response, err := s.CacheManager.GrpcWrapper.getValidatorCommissionAtBlock(address, block); err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Could not load validator commission info")
	} else {
		for _, coin := range response {
			if value, err := strconv.ParseFloat(coin.Amount.String(), 64); err != nil {
				log.Error().
					Err(err).
					Msg("Could not parse balance")
			} else {
				sb.WriteString(s.CodeSerializer(Printer.Sprintf(
					"%.6f %s",
					float64(value)/DenomCoefficient,
					Denom,
				)) + "\n")
			}
		}
	}

	return sb.String()
}

func (s Serializer) getDelegatorRewardsAtBlock(validator string, delegator string, block int64) string {
	var sb strings.Builder

	if response, err := s.CacheManager.GrpcWrapper.getDelegatorRewardsAtBlock(validator, delegator, block); err != nil {
		log.Warn().Err(err).
			Str("validator", validator).
			Str("delegator", delegator).
			Msg("Could not load delegator rewards info")
	} else {
		for _, coin := range response {
			if value, err := strconv.ParseFloat(coin.Amount.String(), 64); err != nil {
				log.Error().
					Err(err).
					Msg("Could not parse balance")
			} else {
				sb.WriteString(s.getTokensMaybeWithDollarPrice(value/DenomCoefficient, Denom) + "\n")
			}
		}
	}

	return sb.String()
}
