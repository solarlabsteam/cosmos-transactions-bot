package main

import (
	"fmt"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosDistributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/gogo/protobuf/proto"
)

type MsgWithdrawDelegatorReward struct {
	ValidatorAddress string
	DelegatorAddress string
}

func (msg MsgWithdrawDelegatorReward) Empty() bool {
	return msg.ValidatorAddress == ""
}

func (msg MsgWithdrawDelegatorReward) Serialize(serializer Serializer) string {
	return fmt.Sprintf(`%s
%s %s
%s %s`,
		serializer.StrongSerializer("Withdraw rewards"),
		serializer.StrongSerializer("From: "),
		serializer.LinksSerializer(makeMintscanValidatorLink(msg.ValidatorAddress), msg.ValidatorAddress),
		serializer.StrongSerializer("To: "),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.DelegatorAddress), msg.DelegatorAddress),
	)
}

func ParseMsgWithdrawDelegatorReward(message *cosmosTypes.Any) MsgWithdrawDelegatorReward {
	var parsedMessage cosmosDistributionTypes.MsgWithdrawDelegatorReward
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgWithdrawDelegatorReward")
		return MsgWithdrawDelegatorReward{}
	}

	log.Info().
		Str("from", parsedMessage.ValidatorAddress).
		Str("to", parsedMessage.DelegatorAddress).
		Msg("MsgWithdrawDelegatorReward")

	return MsgWithdrawDelegatorReward{
		ValidatorAddress: parsedMessage.ValidatorAddress,
		DelegatorAddress: parsedMessage.DelegatorAddress,
	}
}

func processMsgSetWithdrawAddress(message *cosmosTypes.Any) string {
	var parsedMessage cosmosDistributionTypes.MsgSetWithdrawAddress
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgSetWithdrawAddress")
	}

	log.Info().
		Str("by", parsedMessage.DelegatorAddress).
		Str("withdraw_address", parsedMessage.WithdrawAddress).
		Msg("MsgSetWithdrawAddress")
	return fmt.Sprintf(`<strong>Set withdraw address</strong>
<strong>By: </strong><a href="%s">%s</a>
<strong>New withdraw address: </strong><a href="%s">%s</a>`,
		makeMintscanAccountLink(parsedMessage.DelegatorAddress),
		parsedMessage.DelegatorAddress,
		makeMintscanAccountLink(parsedMessage.WithdrawAddress),
		parsedMessage.WithdrawAddress,
	)
}

func processMsgWithdrawValidatorCommission(message *cosmosTypes.Any) string {
	var parsedMessage cosmosDistributionTypes.MsgWithdrawValidatorCommission
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgWithdrawValidatorCommission")
	}

	log.Info().
		Str("address", parsedMessage.ValidatorAddress).
		Msg("MsgWithdrawValidatorCommission")
	return fmt.Sprintf(`<strong>Withdraw validator commission</strong>
<strong>Wallet: </strong><a href="%s">%s</a>`,
		makeMintscanValidatorLink(parsedMessage.ValidatorAddress),
		parsedMessage.ValidatorAddress,
	)
}
