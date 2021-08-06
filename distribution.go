package main

import (
	"fmt"
	"strings"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosDistributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/gogo/protobuf/proto"
)

type MsgWithdrawDelegatorReward struct {
	ValidatorAddress string
	DelegatorAddress string
	Block            int64
}

func (msg MsgWithdrawDelegatorReward) Empty() bool {
	return msg.ValidatorAddress == ""
}

func (msg MsgWithdrawDelegatorReward) Serialize(serializer Serializer) string {
	var sb strings.Builder

	sb.WriteString(serializer.StrongSerializer("Withdraw rewards") + "\n")
	sb.WriteString(serializer.getDelegatorRewardsAtBlock(msg.ValidatorAddress, msg.DelegatorAddress, msg.Block-1))

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

func ParseMsgWithdrawDelegatorReward(message *cosmosTypes.Any, block int64) MsgWithdrawDelegatorReward {
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
		Block:            block,
	}
}

type MsgSetWithdrawAddress struct {
	WithdrawAddress  string
	DelegatorAddress string
}

func (msg MsgSetWithdrawAddress) Empty() bool {
	return msg.DelegatorAddress == ""
}

func ParseMsgSetWithdrawAddress(message *cosmosTypes.Any) MsgSetWithdrawAddress {
	var parsedMessage cosmosDistributionTypes.MsgSetWithdrawAddress
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgSetWithdrawAddress")
		return MsgSetWithdrawAddress{}
	}

	log.Info().
		Str("by", parsedMessage.DelegatorAddress).
		Str("withdraw_address", parsedMessage.WithdrawAddress).
		Msg("MsgSetWithdrawAddress")

	return MsgSetWithdrawAddress{
		DelegatorAddress: parsedMessage.DelegatorAddress,
		WithdrawAddress:  parsedMessage.WithdrawAddress,
	}
}

func (msg MsgSetWithdrawAddress) Serialize(serializer Serializer) string {
	return fmt.Sprintf(`%s
%s %s
%s %s`,
		serializer.StrongSerializer("Set withdraw address"),
		serializer.StrongSerializer("By:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.DelegatorAddress), msg.DelegatorAddress),
		serializer.StrongSerializer("New withdraw address: "),
		serializer.getWalletWithLabel(msg.WithdrawAddress),
	)
}

type MsgWithdrawValidatorCommission struct {
	ValidatorAddress string
	Block            int64
}

func (msg MsgWithdrawValidatorCommission) Empty() bool {
	return msg.ValidatorAddress == ""
}

func ParseMsgWithdrawValidatorCommission(message *cosmosTypes.Any, block int64) MsgWithdrawValidatorCommission {
	var parsedMessage cosmosDistributionTypes.MsgWithdrawValidatorCommission
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgWithdrawValidatorCommission")
		return MsgWithdrawValidatorCommission{}
	}

	log.Info().
		Str("address", parsedMessage.ValidatorAddress).
		Msg("MsgWithdrawValidatorCommission")

	return MsgWithdrawValidatorCommission{
		ValidatorAddress: parsedMessage.ValidatorAddress,
		Block:            block,
	}
}

func (msg MsgWithdrawValidatorCommission) Serialize(serializer Serializer) string {
	var sb strings.Builder

	sb.WriteString(serializer.StrongSerializer("Withdraw validator commission") + "\n")
	sb.WriteString(serializer.getValidatorCommissionAtBlock(msg.ValidatorAddress, msg.Block-1))

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("Validator:"),
		serializer.getValidatorWithName(msg.ValidatorAddress),
	))

	return sb.String()
}
