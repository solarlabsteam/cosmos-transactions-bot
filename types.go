package main

import (
	"fmt"
	"strings"
)

type Msg interface {
	Serialize(Serializer Serializer) string
	Empty() bool
}

type Serializer struct {
	LinksSerializer  func(string, string) string
	StrongSerializer func(string) string
	CodeSerializer   func(string) string
}

type Report struct {
	Tx   Tx
	Msgs []Msg
}

func (r Report) Empty() bool {
	return r.Tx.Hash == "" || len(r.Msgs) == 0
}

type Reporter interface {
	Serialize(Report) string
	Init()
	Enabled() bool
	SendReport(Report) error
	Name() string
	Serializer() Serializer
}

func (s Serializer) getWalletWithLabel(address string) string {
	label, labelFound := labelsConfigManager.getWalletLabel(address)

	var sb strings.Builder

	sb.WriteString(s.LinksSerializer(makeMintscanAccountLink(address), address))

	if labelFound {
		sb.WriteString(fmt.Sprintf(" (%s)", s.CodeSerializer(label)))
	}

	return sb.String()
}
