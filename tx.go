package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/crypto/tmhash"

	abciTypes "github.com/tendermint/tendermint/abci/types"
)

type Tx struct {
	Hash   string
	Height int64
	Memo   string
}

func (tx Tx) Serialize(serializer Serializer) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(
		"Tx %s at block %s",
		serializer.LinksSerializer(makeMintscanTxLink(tx.Hash), tx.Hash[0:8]),
		serializer.LinksSerializer(makeMintscanBlockLink(tx.Height), strconv.FormatInt(tx.Height, 10)),
	))

	if tx.Memo != "" {
		sb.WriteString(fmt.Sprintf(
			"\n%s %s",
			serializer.StrongSerializer("Memo:"),
			serializer.getSingleOrMultilineCodeBlock(tx.Memo),
		))
	}

	return sb.String()
}

func parseTx(txResult abciTypes.TxResult) Tx {
	Hash := fmt.Sprintf("%X", tmhash.Sum(txResult.Tx))
	Height := txResult.Height

	var tx tx.Tx

	if err := proto.Unmarshal(txResult.Tx, &tx); err != nil {
		log.Error().Err(err).Msg("Could not parse tx")
		return Tx{}
	}

	return Tx{
		Hash:   Hash,
		Height: Height,
		Memo:   tx.GetBody().GetMemo(),
	}
}
