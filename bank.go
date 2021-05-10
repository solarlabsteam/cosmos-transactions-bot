package main

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"

	cosmosBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func processMsgSend(message *cosmosTypes.Any) string {
	var sb strings.Builder

	var parsedMessage cosmosBankTypes.MsgSend
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgSend")
	}
	for _, coin := range parsedMessage.Amount {
		log.Info().
			Str("from", parsedMessage.FromAddress).
			Str("to", parsedMessage.ToAddress).
			Str("denom", coin.Denom).
			Int64("amount", coin.Amount.Int64()).
			Msg("MsgSend")
		sb.WriteString(fmt.Sprintf(`<strong>Transfer</strong>
<code>%d %s</code>
<strong>From: </strong><a href="%s">%s</a>
<strong>To: </strong><a href="%s">%s</a>`,
			coin.Amount.Int64(),
			coin.Denom,
			makeMintscanAccountLink(parsedMessage.FromAddress),
			parsedMessage.FromAddress,
			makeMintscanAccountLink(parsedMessage.ToAddress),
			parsedMessage.ToAddress,
		))
	}

	return sb.String()
}
