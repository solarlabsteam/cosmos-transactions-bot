package main

import (
	"fmt"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosStakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
)

type MsgDelegate struct {
	DelegatorAddress string
	ValidatorAddress string
	Denom            string
	Amount           int64
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
		Str("denom", parsedMessage.Amount.Denom).
		Int64("amount", parsedMessage.Amount.Amount.Int64()).
		Msg("MsgDelegate")

	return MsgDelegate{
		DelegatorAddress: parsedMessage.DelegatorAddress,
		ValidatorAddress: parsedMessage.ValidatorAddress,
		Denom:            parsedMessage.Amount.Denom,
		Amount:           parsedMessage.Amount.Amount.Int64(),
	}
}

func (msg MsgDelegate) Serialize(serializer Serializer) string {
	return fmt.Sprintf(`%s
%s
%s %s
%s %s`,
		serializer.StrongSerializer("Delegate"),
		serializer.CodeSerializer(fmt.Sprintf("%d %s", msg.Amount, msg.Denom)),
		serializer.StrongSerializer("From:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.DelegatorAddress), msg.DelegatorAddress),
		serializer.StrongSerializer("To:"),
		serializer.LinksSerializer(makeMintscanValidatorLink(msg.ValidatorAddress), msg.ValidatorAddress),
	)
}

type MsgBeginRedelegate struct {
	DelegatorAddress    string
	ValidatorSrcAddress string
	ValidatorDstAddress string
	Denom               string
	Amount              int64
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
		Str("denom", parsedMessage.Amount.Denom).
		Int64("amount", parsedMessage.Amount.Amount.Int64()).
		Msg("MsgBeginRedelegate")

	return MsgBeginRedelegate{
		DelegatorAddress:    parsedMessage.DelegatorAddress,
		ValidatorSrcAddress: parsedMessage.ValidatorSrcAddress,
		ValidatorDstAddress: parsedMessage.ValidatorDstAddress,
		Denom:               parsedMessage.Amount.Denom,
		Amount:              parsedMessage.Amount.Amount.Int64(),
	}
}

func (msg MsgBeginRedelegate) Serialize(serializer Serializer) string {
	return fmt.Sprintf(`%s
%s
%s %s
%s %s
%s %s`,
		serializer.StrongSerializer("Redelegate"),
		serializer.CodeSerializer(fmt.Sprintf("%d %s", msg.Amount, msg.Denom)),
		serializer.StrongSerializer("By:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.DelegatorAddress), msg.DelegatorAddress),
		serializer.StrongSerializer("From:"),
		serializer.LinksSerializer(makeMintscanValidatorLink(msg.ValidatorSrcAddress), msg.ValidatorSrcAddress),
		serializer.StrongSerializer("To:"),
		serializer.LinksSerializer(makeMintscanValidatorLink(msg.ValidatorDstAddress), msg.ValidatorDstAddress),
	)
}

type MsgUndelegate struct {
	DelegatorAddress string
	ValidatorAddress string
	Denom            string
	Amount           int64
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
		Str("denom", parsedMessage.Amount.Denom).
		Int64("amount", parsedMessage.Amount.Amount.Int64()).
		Msg("MsgUndelegate")

	return MsgUndelegate{
		DelegatorAddress: parsedMessage.DelegatorAddress,
		ValidatorAddress: parsedMessage.ValidatorAddress,
		Denom:            parsedMessage.Amount.Denom,
		Amount:           parsedMessage.Amount.Amount.Int64(),
	}
}

func (msg MsgUndelegate) Serialize(serializer Serializer) string {
	return fmt.Sprintf(`%s
%s
%s %s
%s %s`,
		serializer.StrongSerializer("Undelegate"),
		serializer.CodeSerializer(fmt.Sprintf("%d %s", msg.Amount, msg.Denom)),
		serializer.StrongSerializer("From:"),
		serializer.LinksSerializer(makeMintscanValidatorLink(msg.ValidatorAddress), msg.ValidatorAddress),
		serializer.StrongSerializer("By:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.DelegatorAddress), msg.DelegatorAddress),
	)
}
