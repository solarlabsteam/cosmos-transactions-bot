package main

import (
	"fmt"
	"strings"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosStakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
)

type MsgDelegate struct {
	DelegatorAddress string
	ValidatorAddress string
	Validator        cosmosStakingTypes.Validator
	Denom            string
	Amount           float64
}

func (msg MsgDelegate) Empty() bool {
	return msg.DelegatorAddress == ""
}

func ParseMsgDelegate(message *cosmosTypes.Any) MsgDelegate {
	var parsedMessage cosmosStakingTypes.MsgDelegate
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgDelegate")
		return MsgDelegate{}
	}

	log.Info().
		Str("from", parsedMessage.DelegatorAddress).
		Str("to", parsedMessage.ValidatorAddress).
		Str("denom", Denom).
		Float64("amount", float64(parsedMessage.Amount.Amount.Int64())/DenomCoefficient).
		Msg("MsgDelegate")

	return MsgDelegate{
		DelegatorAddress: parsedMessage.DelegatorAddress,
		ValidatorAddress: parsedMessage.ValidatorAddress,
		Denom:            Denom,
		Amount:           float64(parsedMessage.Amount.Amount.Int64()) / DenomCoefficient,
	}
}

func (msg MsgDelegate) Serialize(serializer Serializer) string {
	var sb strings.Builder
	sb.WriteString(serializer.StrongSerializer("Delegate") + "\n")
	sb.WriteString(serializer.getTokensMaybeWithDollarPrice(msg.Amount, msg.Denom) + "\n")

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("From:"),
		serializer.getWalletWithLabel(msg.DelegatorAddress),
	))

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("To:"),
		serializer.getValidatorWithName(msg.ValidatorAddress),
	))

	return sb.String()
}

type MsgBeginRedelegate struct {
	DelegatorAddress    string
	ValidatorSrcAddress string
	ValidatorDstAddress string
	Denom               string
	Amount              float64
}

func (msg MsgBeginRedelegate) Empty() bool {
	return msg.DelegatorAddress == ""
}

func ParseMsgBeginRedelegate(message *cosmosTypes.Any) MsgBeginRedelegate {
	var parsedMessage cosmosStakingTypes.MsgBeginRedelegate
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgBeginRedelegate")
		return MsgBeginRedelegate{}
	}

	log.Info().
		Str("by", parsedMessage.DelegatorAddress).
		Str("from", parsedMessage.ValidatorSrcAddress).
		Str("to", parsedMessage.ValidatorDstAddress).
		Str("denom", Denom).
		Float64("amount", float64(parsedMessage.Amount.Amount.Int64())/DenomCoefficient).
		Msg("MsgBeginRedelegate")

	return MsgBeginRedelegate{
		DelegatorAddress:    parsedMessage.DelegatorAddress,
		ValidatorSrcAddress: parsedMessage.ValidatorSrcAddress,
		ValidatorDstAddress: parsedMessage.ValidatorDstAddress,
		Denom:               Denom,
		Amount:              float64(parsedMessage.Amount.Amount.Int64()) / DenomCoefficient,
	}
}

func (msg MsgBeginRedelegate) Serialize(serializer Serializer) string {
	var sb strings.Builder
	sb.WriteString(serializer.StrongSerializer("Redelegate") + "\n")
	sb.WriteString(serializer.getTokensMaybeWithDollarPrice(msg.Amount, msg.Denom) + "\n")

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("By:"),
		serializer.getWalletWithLabel(msg.DelegatorAddress),
	))

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("From:"),
		serializer.getValidatorWithName(msg.ValidatorSrcAddress),
	))

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("To:"),
		serializer.getValidatorWithName(msg.ValidatorDstAddress),
	))

	return sb.String()
}

type MsgUndelegate struct {
	DelegatorAddress string
	ValidatorAddress string
	Denom            string
	Amount           float64
}

func (msg MsgUndelegate) Empty() bool {
	return msg.DelegatorAddress == ""
}

func ParseMsgUndelegate(message *cosmosTypes.Any) MsgUndelegate {
	var parsedMessage cosmosStakingTypes.MsgUndelegate
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgUndelegate")
		return MsgUndelegate{}
	}

	log.Info().
		Str("from", parsedMessage.ValidatorAddress).
		Str("by", parsedMessage.DelegatorAddress).
		Str("denom", Denom).
		Float64("amount", float64(parsedMessage.Amount.Amount.Int64())/DenomCoefficient).
		Msg("MsgUndelegate")

	return MsgUndelegate{
		DelegatorAddress: parsedMessage.DelegatorAddress,
		ValidatorAddress: parsedMessage.ValidatorAddress,
		Denom:            Denom,
		Amount:           float64(parsedMessage.Amount.Amount.Int64()) / DenomCoefficient,
	}
}

func (msg MsgUndelegate) Serialize(serializer Serializer) string {
	var sb strings.Builder
	sb.WriteString(serializer.StrongSerializer("Undelegate") + "\n")
	sb.WriteString(serializer.getTokensMaybeWithDollarPrice(msg.Amount, msg.Denom) + "\n")

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("From:"),
		serializer.getValidatorWithName(msg.ValidatorAddress),
	))

	sb.WriteString(fmt.Sprintf("%s %s",
		serializer.StrongSerializer("To:"),
		serializer.getWalletWithLabel(msg.DelegatorAddress),
	))

	return sb.String()
}
