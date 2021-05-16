package main

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

type SlackReporter struct {
	SlackToken string
	SlackChat  string

	SlackClient        slack.Client
	MarkdownSerializer Serializer
}

func (r SlackReporter) Serialize(report Report) string {
	var sb strings.Builder

	sb.WriteString(report.Tx.Serialize(r.Serializer()) + "\n\n")

	for _, msg := range report.Msgs {
		sb.WriteString(msg.Serialize(r.Serializer()) + "\n\n")
	}

	return sb.String()
}

func (r *SlackReporter) Init() {
	if r.SlackToken == "" || r.SlackChat == "" {
		log.Debug().Msg("Slack credentials not set, not creating Slack reporter.")
		return
	}

	client := slack.New(r.SlackToken)
	r.SlackClient = *client
	r.MarkdownSerializer = Serializer{
		LinksSerializer: func(address string, text string) string {
			return fmt.Sprintf(`<%s|%s>",`, address, text)
		},
		StrongSerializer: func(text string) string {
			return fmt.Sprintf(`*%s%`, text)
		},
		CodeSerializer: func(text string) string {
			return fmt.Sprintf("`%s`", text)
		},
	}
}

func (r SlackReporter) Enabled() bool {
	return r.SlackToken != "" && r.SlackChat != ""
}

func (r SlackReporter) Serializer() Serializer {
	return r.MarkdownSerializer
}

func (r SlackReporter) SendReport(report Report) error {
	serializedReport := r.Serialize(report)
	_, _, err := r.SlackClient.PostMessage(
		r.SlackChat,
		slack.MsgOptionText(serializedReport, false),
	)
	return err
}

func (r SlackReporter) Name() string {
	return "SlackReporter"
}
