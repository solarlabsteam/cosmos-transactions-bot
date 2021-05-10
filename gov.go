package main

import (
	"fmt"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosGovTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/gogo/protobuf/proto"
)

func processMsgVote(message *cosmosTypes.Any) string {
	var parsedMessage cosmosGovTypes.MsgVote
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgVote")
	}

	log.Info().
		Int64("proposal_id", int64(parsedMessage.ProposalId)).
		Str("voter", parsedMessage.Voter).
		Str("option", parsedMessage.Option.String()).
		Msg("MsgVote")
	return fmt.Sprintf("<strong>Vote</strong>\n<strong> Voted: </strong>%s\n<strong>Proposal ID: </strong>%d\n<strong>Voter: </strong>%s",
		parsedMessage.Option.String(),
		parsedMessage.ProposalId,
		parsedMessage.Voter,
	)
}
