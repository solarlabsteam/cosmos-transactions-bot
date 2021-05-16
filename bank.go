package main

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"

	cosmosBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type MsgSend struct {
	FromAddress string
	ToAddress   string
	Coins       []Coin
}

type Coin struct {
	Amount int64
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
			Str("denom", coin.Denom).
			Int64("amount", coin.Amount.Int64()).
			Msg("MsgSend")

		coins = append(coins, Coin{
			Amount: coin.Amount.Int64(),
			Denom:  coin.Denom,
		})
	}

	return MsgSend{
		FromAddress: parsedMessage.FromAddress,
		ToAddress:   parsedMessage.ToAddress,
		Coins:       coins,
	}
}

func (msg MsgSend) Serialize(serializer Serializer) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", serializer.StrongSerializer("Transfer")))

	for _, coin := range msg.Coins {
		sb.WriteString(serializer.CodeSerializer(fmt.Sprintf("%d %s", coin.Amount, coin.Denom)) + "\n")
	}

	sb.WriteString(fmt.Sprintf(`
%s %s
%s %s`,
		serializer.StrongSerializer("From:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.FromAddress), msg.FromAddress),
		serializer.StrongSerializer("To:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.ToAddress), msg.ToAddress),
	))

	return sb.String()
}
