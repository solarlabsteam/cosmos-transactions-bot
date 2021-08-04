package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
)

type SlackReporter struct {
	SlackToken         string
	SlackChat          string
	SlackSigningSecret string
	SlackListenAddress string

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
			return fmt.Sprintf(`<%s|%s>`, address, text)
		},
		StrongSerializer: func(text string) string {
			return fmt.Sprintf(`*%s*`, text)
		},
		CodeSerializer: func(text string) string {
			return fmt.Sprintf("`%s`", text)
		},
	}

	r.InitSlashHandler()
}

func (reporter *SlackReporter) InitSlashHandler() {
	if reporter.SlackSigningSecret == "" {
		log.Debug().Msg("Slack signing secret not set, not exposing slash commands handler.")
		return
	}

	http.HandleFunc("/slash", func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, reporter.SlackSigningSecret)
		if err != nil {
			log.Warn().Err(err).Msg("Could not create Slack secrets verifier.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			log.Warn().Err(err).Msg("Could not parse Slack slash command.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = verifier.Ensure(); err != nil {
			log.Warn().Err(err).Msg("Could not verify Slack slash command request.")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		log.Info().
			Str("command", s.Command).
			Str("text", s.Text).
			Str("channel", s.ChannelName).
			Str("user", s.UserName).
			Msg("Received command")

		switch s.Command {
		case "/set-alias":
			reporter.processAddAliasCommand(s, w, r)
		default:
			log.Debug().Msg("Unsupported command, skipping.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Debug().Msg("Slash command processed.")
	})

	if err := http.ListenAndServe(reporter.SlackListenAddress, nil); err != nil {
		log.Fatal().Err(err).Msg("Could not start Slack slash commands handler")
	}

	log.Info().Str("address", reporter.SlackListenAddress).Msg("Slack slash commands handler is listening")

}

func (reporter *SlackReporter) processAddAliasCommand(s slack.SlashCommand, w http.ResponseWriter, r *http.Request) {
	params := &slack.Msg{Text: s.Text}
	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		log.Error().Err(err).Msg("Could not send response to /set-alias command")
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
