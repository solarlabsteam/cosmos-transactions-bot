package main

import (
	"fmt"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosStakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
)

func processMsgDelegate(message *cosmosTypes.Any) string {
	var parsedMessage cosmosStakingTypes.MsgDelegate
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgDelegate")
	}

	log.Info().
		Str("from", parsedMessage.DelegatorAddress).
		Str("to", parsedMessage.ValidatorAddress).
		Str("denom", parsedMessage.Amount.Denom).
		Int64("amount", parsedMessage.Amount.Amount.Int64()).
		Msg("MsgDelegate")
	return fmt.Sprintf(`<strong>Delegate</strong>
<code>%d%s</code>
<strong>From: </strong><a href="%s">%s</a>
<strong>To: </strong><a href="%s">%s</a>`,
		parsedMessage.Amount.Amount.Int64(),
		parsedMessage.Amount.Denom,
		makeMintscanAccountLink(parsedMessage.DelegatorAddress),
		parsedMessage.DelegatorAddress,
		makeMintscanValidatorLink(parsedMessage.ValidatorAddress),
		parsedMessage.ValidatorAddress,
	)
}

func processMsgBeginRedelegate(message *cosmosTypes.Any) string {
	var parsedMessage cosmosStakingTypes.MsgBeginRedelegate
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgBeginRedelegate")
	}

	log.Info().
		Str("by", parsedMessage.DelegatorAddress).
		Str("from", parsedMessage.ValidatorSrcAddress).
		Str("to", parsedMessage.ValidatorDstAddress).
		Str("denom", parsedMessage.Amount.Denom).
		Int64("amount", parsedMessage.Amount.Amount.Int64()).
		Msg("MsgBeginRedelegate")
	return fmt.Sprintf(`<strong>Redelegate</strong>
<code>%d%s</code>
<strong>By: </strong><a href="%s">%s</a>
<strong>From: </strong><a href="%s">%s</a>
<strong>To: </strong><a href="%s">%s</a>`,
		parsedMessage.Amount.Amount.Int64(),
		parsedMessage.Amount.Denom,
		makeMintscanAccountLink(parsedMessage.DelegatorAddress),
		parsedMessage.DelegatorAddress,
		makeMintscanValidatorLink(parsedMessage.ValidatorSrcAddress),
		parsedMessage.ValidatorSrcAddress,
		makeMintscanValidatorLink(parsedMessage.ValidatorDstAddress),
		parsedMessage.ValidatorDstAddress,
	)
}

func processMsgUndelegate(message *cosmosTypes.Any) string {
	var parsedMessage cosmosStakingTypes.MsgUndelegate
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgUndelegate")
	}

	log.Info().
		Str("from", parsedMessage.ValidatorAddress).
		Str("by", parsedMessage.DelegatorAddress).
		Str("denom", parsedMessage.Amount.Denom).
		Int64("amount", parsedMessage.Amount.Amount.Int64()).
		Msg("MsgUndelegate")
	return fmt.Sprintf(`<strong>Undelegate</strong>
<code>%d%s</code>
<strong>From: </strong><a href="%s">%s</a>
<strong>By: </strong><a href="%s">%s</a>`,
		parsedMessage.Amount.Amount.Int64(),
		parsedMessage.Amount.Denom,
		makeMintscanValidatorLink(parsedMessage.ValidatorAddress),
		parsedMessage.ValidatorAddress,
		makeMintscanAccountLink(parsedMessage.DelegatorAddress),
		parsedMessage.DelegatorAddress,
	)
}
