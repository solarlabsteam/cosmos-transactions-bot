package main

import (
	"fmt"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosGovTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/gogo/protobuf/proto"
)

type MsgVote struct {
	ProposalId uint64
	Voter      string
	Option     string
}

func (msg MsgVote) Empty() bool {
	return msg.Voter == ""
}

func (msg MsgVote) Serialize(serializer Serializer) string {
	return fmt.Sprintf(`%s
%s %s
%s %d
%s %s`,
		serializer.StrongSerializer("Vote"),
		serializer.StrongSerializer("Voted: "),
		msg.Option,
		serializer.StrongSerializer("Proposal ID: "),
		msg.ProposalId,
		serializer.StrongSerializer("Voter: "),
		msg.Voter,
	)
}

func ParseMsgVote(message *cosmosTypes.Any) MsgVote {
	var parsedMessage cosmosGovTypes.MsgVote
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgVote")
		return MsgVote{}
	}

	log.Info().
		Int64("proposal_id", int64(parsedMessage.ProposalId)).
		Str("voter", parsedMessage.Voter).
		Str("option", parsedMessage.Option.String()).
		Msg("MsgVote")

	return MsgVote{
		ProposalId: parsedMessage.ProposalId,
		Voter:      parsedMessage.Voter,
		Option:     parsedMessage.Option.String(),
	}
}
