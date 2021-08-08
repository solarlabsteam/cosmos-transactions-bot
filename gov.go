package main

import (
	"fmt"
	"strings"

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

type MsgSubmitProposal struct {
	Title       string
	Description string
	Proposer    string
}

func (msg MsgSubmitProposal) Empty() bool {
	return msg.Title == ""
}

func ParseMsgSubmitProposal(message *cosmosTypes.Any, block int64) MsgSubmitProposal {
	var parsedMessage cosmosGovTypes.MsgSubmitProposal
	if err := proto.Unmarshal(message.Value, &parsedMessage); err != nil {
		log.Error().Err(err).Msg("Could not parse MsgSubmitProposal")
		return MsgSubmitProposal{}
	}

	log.Info().
		Str("title", parsedMessage.GetContent().GetTitle()).
		Str("description", parsedMessage.GetContent().GetDescription()).
		Str("proposer", parsedMessage.Proposer).
		Msg("MsgWithdrawValidatorCommission")

	return MsgSubmitProposal{
		Title:       parsedMessage.GetContent().GetTitle(),
		Description: parsedMessage.GetContent().GetDescription(),
		Proposer:    parsedMessage.Proposer,
	}
}

func (msg MsgSubmitProposal) Serialize(serializer Serializer) string {
	var sb strings.Builder

	sb.WriteString(serializer.StrongSerializer("New proposal") + "\n")
	sb.WriteString(serializer.LinksSerializer(makeMintscanProposalsLink(), "Mintscan") + "\n")

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("Proposer:"),
		serializer.LinksSerializer(makeMintscanAccountLink(msg.Proposer), msg.Proposer),
	))

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("Title:"),
		serializer.getSingleOrMultilineCodeBlock(msg.Title),
	))

	sb.WriteString(fmt.Sprintf("%s %s\n",
		serializer.StrongSerializer("Description:"),
		serializer.getSingleOrMultilineCodeBlock(msg.Description),
	))

	return sb.String()
}
