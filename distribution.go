package main

import (
	"fmt"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosDistributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/gogo/protobuf/proto"
)

func processMsgWithdrawDelegatorReward(message *cosmosTypes.Any) string {
	var parsedMessage cosmosDistributionTypes.MsgWithdrawDelegatorReward
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgWithdrawDelegatorReward")
	}

	log.Info().
		Str("from", parsedMessage.ValidatorAddress).
		Str("to", parsedMessage.DelegatorAddress).
		Msg("MsgWithdrawDelegatorReward")
	return fmt.Sprintf(`<strong>Withdraw rewards</strong>
<strong>From: </strong><a href="%s">%s</a>
<strong>To: </strong><a href="%s">%s</a>`,
		makeMintscanValidatorLink(parsedMessage.ValidatorAddress),
		parsedMessage.ValidatorAddress,
		makeMintscanAccountLink(parsedMessage.DelegatorAddress),
		parsedMessage.DelegatorAddress,
	)
}
