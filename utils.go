package main

import (
	"fmt"
)

func makeMintscanLink(suffix string) string {
	return fmt.Sprintf("https://mintscan.io/%s/%s", MintscanProject, suffix)
}

func makeMintscanTxLink(hash string) string {
	return makeMintscanLink(fmt.Sprintf("txs/%s", hash))
}

func makeMintscanBlockLink(block int64) string {
	return makeMintscanLink(fmt.Sprintf("blocks/%d", block))
}

func makeMintscanAccountLink(account string) string {
	return makeMintscanLink(fmt.Sprintf("account/%s", account))
}

func makeMintscanValidatorLink(validator string) string {
	return makeMintscanLink(fmt.Sprintf("validators/%s", validator))
}

func makeMintscanProposalsLink() string {
	return makeMintscanLink("validators")
}
