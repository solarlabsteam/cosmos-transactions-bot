package main

import (
	"fmt"
	"strings"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	ibcTypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	ibcChannelTypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	"github.com/gogo/protobuf/proto"
)

type MsgIbcTransfer struct {
	FromAddress string
	ToAddress   string
	SrcPort     string
	SrcChannel  string
	Amount      float64
	Denom       string
}

func (msg MsgIbcTransfer) Empty() bool {
	return msg.FromAddress == ""
}

func (msg MsgIbcTransfer) Serialize(serializer Serializer) string {
	fromLabel, fromLabelFound := labelsConfigManager.getWalletLabel(msg.FromAddress)
	toLabel, toLabelFound := labelsConfigManager.getWalletLabel(msg.ToAddress)

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", serializer.StrongSerializer("IBC transfer")))

	if msg.Denom == BaseDenom {
		// for native tokens
		sb.WriteString(serializer.getTokensMaybeWithDollarPrice(msg.Amount/DenomCoefficient, Denom) + "\n")
	} else {
		// for non-native tokens, like ibc/xxxxxx
		sb.WriteString(serializer.getTokensFormatted(msg.Amount, msg.Denom) + "\n")
	}

	sb.WriteString(fmt.Sprintf(`%s %s`,
		serializer.StrongSerializer("From:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.FromAddress), msg.FromAddress),
	))

	if fromLabelFound {
		sb.WriteString(fmt.Sprintf(
			" (%s)",
			serializer.CodeSerializer(fromLabel),
		))
	}

	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf(`%s %s`,
		serializer.StrongSerializer("To:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.ToAddress), msg.ToAddress),
	))

	if toLabelFound {
		sb.WriteString(fmt.Sprintf(
			" (%s)",
			serializer.CodeSerializer(toLabel),
		))
	}

	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf(`%s %s`,
		serializer.StrongSerializer("IBC channel:"),
		msg.SrcChannel,
	))

	return sb.String()
}

func ParseMsgIbcTransfer(message *cosmosTypes.Any) MsgIbcTransfer {
	var parsedMessage ibcTypes.MsgTransfer
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgIbcTransfer")
		return MsgIbcTransfer{}
	}

	log.Info().
		Str("from", parsedMessage.Sender).
		Str("to", parsedMessage.Receiver).
		Float64("amount", parsedMessage.Token.Amount.ToDec().MustFloat64()).
		Str("denom", parsedMessage.Token.Denom).
		Msg("MsgIbcTransfer")

	return MsgIbcTransfer{
		FromAddress: parsedMessage.Sender,
		ToAddress:   parsedMessage.Receiver,
		SrcPort:     parsedMessage.SourcePort,
		SrcChannel:  parsedMessage.SourceChannel,
		Amount:      parsedMessage.Token.Amount.ToDec().MustFloat64(),
		Denom:       parsedMessage.Token.Denom,
	}
}

type MsgIbcRecvPacket struct {
	Signer      string
	FromAddress string
	ToAddress   string
	SrcPort     string
	SrcChannel  string
	DstPort     string
	DstChannel  string
	Amount      float64
	Denom       string
}

func (msg MsgIbcRecvPacket) Empty() bool {
	return msg.Signer == ""
}

func (msg MsgIbcRecvPacket) Serialize(serializer Serializer) string {
	fromLabel, fromLabelFound := labelsConfigManager.getWalletLabel(msg.FromAddress)
	toLabel, toLabelFound := labelsConfigManager.getWalletLabel(msg.ToAddress)

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", serializer.StrongSerializer("IBC receive packet")))

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("Signer:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.Signer), msg.Signer),
	))

	if msg.Denom != "" {
		if msg.Denom == BaseDenom {
			// for native tokens
			sb.WriteString(serializer.getTokensMaybeWithDollarPrice(msg.Amount/DenomCoefficient, Denom) + "\n")
		} else {
			// for non-native tokens, like ibc/xxxxxx
			sb.WriteString(serializer.getTokensFormatted(msg.Amount, msg.Denom) + "\n")
		}
	}

	if msg.FromAddress != "" {
		sb.WriteString(fmt.Sprintf(`%s %s`,
			serializer.StrongSerializer("From:"),
			serializer.LinksSerializer(makeMintscanAccountLink(msg.FromAddress), msg.FromAddress),
		))

		if fromLabelFound {
			sb.WriteString(fmt.Sprintf(
				" (%s)",
				serializer.CodeSerializer(fromLabel),
			))
		}

		sb.WriteString("\n")
	}

	if msg.ToAddress != "" {
		sb.WriteString(fmt.Sprintf(`%s %s`,
			serializer.StrongSerializer("To:"),
			serializer.LinksSerializer(makeMintscanAccountLink(msg.ToAddress), msg.ToAddress),
		))

		if toLabelFound {
			sb.WriteString(fmt.Sprintf(
				" (%s)",
				serializer.CodeSerializer(toLabel),
			))
		}

		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf(`%s %s -> %s`,
		serializer.StrongSerializer("IBC channel:"),
		msg.SrcChannel,
		msg.DstChannel,
	))

	return sb.String()
}

func ParseMsgIbcRecvPacket(message *cosmosTypes.Any) MsgIbcRecvPacket {
	var parsedMessage ibcChannelTypes.MsgRecvPacket
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgIbcRecvPacket")
		return MsgIbcRecvPacket{}
	}

	result := MsgIbcRecvPacket{
		Signer:     parsedMessage.Signer,
		SrcPort:    parsedMessage.Packet.SourcePort,
		SrcChannel: parsedMessage.Packet.SourceChannel,
		DstPort:    parsedMessage.Packet.DestinationPort,
		DstChannel: parsedMessage.Packet.DestinationChannel,
	}

	var data ibcTypes.FungibleTokenPacketData
	if err := ibcTypes.ModuleCdc.UnmarshalJSON(parsedMessage.Packet.Data, &data); err != nil {
		log.Warn().Err(err).Msg("Could not parse MsgIbcRecvPacket data")
		return result
	} else {
		result.FromAddress = data.Sender
		result.ToAddress = data.Receiver
		result.Amount = float64(data.Amount)
		result.Denom = data.Denom
	}

	log.Info().
		Str("signer", parsedMessage.Signer).
		Str("from", data.Sender).
		Str("to", data.Receiver).
		Uint64("amount", data.Amount).
		Str("denom", data.Denom).
		Msg("MsgIbcRecvPacket")

	return result
}
