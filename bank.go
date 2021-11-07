package main

import (
	"fmt"
	"strings"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"
)

type MsgSend struct {
	FromAddress string
	ToAddress   string
	Coins       []Coin
}

type Coin struct {
	Amount float64
	Denom  string
}

func (msg MsgSend) Empty() bool {
	return msg.FromAddress == ""
}

func ParseMsgSend(message *cosmosTypes.Any) MsgSend {
	var parsedMessage cosmosBankTypes.MsgSend
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgSend")
		return MsgSend{}
	}

	coins := []Coin{}

	for _, coin := range parsedMessage.Amount {
		log.Info().
			Str("from", parsedMessage.FromAddress).
			Str("to", parsedMessage.ToAddress).
			Str("denom", Denom).
			Float64("amount", float64(coin.Amount.Int64())/DenomCoefficient).
			Msg("MsgSend")

		coins = append(coins, Coin{
			Amount: float64(coin.Amount.Int64()) / DenomCoefficient,
			Denom:  Denom,
		})
	}

	return MsgSend{
		FromAddress: parsedMessage.FromAddress,
		ToAddress:   parsedMessage.ToAddress,
		Coins:       coins,
	}
}

func (msg MsgSend) Serialize(serializer Serializer) string {
	fromLabel, fromLabelFound := labelsConfigManager.getWalletLabel(msg.FromAddress)
	toLabel, toLabelFound := labelsConfigManager.getWalletLabel(msg.ToAddress)

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", serializer.StrongSerializer("Transfer")))

	for _, coin := range msg.Coins {
		sb.WriteString(serializer.getTokensMaybeWithDollarPrice(coin.Amount, coin.Denom) + "\n")
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

	return sb.String()
}
