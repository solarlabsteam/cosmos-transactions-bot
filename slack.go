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

	SlackSetAliasCommand    string
	SlackClearAliasCommand  string
	SlackListAliasesCommand string

	SlackClient        slack.Client
	MarkdownSerializer Serializer
	CacheManager       *CacheManager
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
		MultilineCodeSerializer: func(text string) string {
			return fmt.Sprintf("```\n%s\n```", text)
		},
		CacheManager: r.CacheManager,
	}

	go r.InitSlashHandler()
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
		case reporter.SlackSetAliasCommand:
			reporter.processSetAliasCommand(s, w)
		case reporter.SlackClearAliasCommand:
			reporter.processClearAliasCommand(s, w)
		case reporter.SlackListAliasesCommand:
			reporter.processListAliasesCommand(s, w)
		default:
			log.Debug().Msg("Unsupported command, skipping.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Debug().Msg("Slash command processed.")
	})

	log.Info().Str("address", reporter.SlackListenAddress).Msg("Slack slash commands handler is listening")
	if err := http.ListenAndServe(reporter.SlackListenAddress, nil); err != nil {
		log.Fatal().Err(err).Msg("Could not start Slack slash commands handler")
	}
}

func (reporter *SlackReporter) processSetAliasCommand(s slack.SlashCommand, w http.ResponseWriter) {
	text := fmt.Sprintf("Usage: `%s` &lt;wallet-address&gt; &lt;alias&gt;", reporter.SlackSetAliasCommand)

	args := strings.SplitN(s.Text, " ", 2)

	if len(args) >= 2 {
		labelsConfigManager.setWalletLabel(args[0], args[1])
		text = fmt.Sprintf(
			"Successfully set alias for %s: %s",
			reporter.MarkdownSerializer.LinksSerializer(makeMintscanAccountLink(args[0]), args[0]),
			reporter.MarkdownSerializer.CodeSerializer(args[1]),
		)
	} else {
		log.Info().Msg("/set-alias: args length < 2")
	}

	if err := writeMessage(text, w); err != nil {
		log.Error().Err(err).Msg("Could not send response to /set-alias command")
	}
}

func (reporter *SlackReporter) processClearAliasCommand(s slack.SlashCommand, w http.ResponseWriter) {
	text := fmt.Sprintf("Usage: `%s` &lt;wallet-address&gt;", reporter.SlackClearAliasCommand)

	if strings.TrimSpace(s.Text) != "" {
		labelsConfigManager.clearWalletLabel(s.Text)
		text = fmt.Sprintf(
			"Successfully cleared alias for %s",
			reporter.MarkdownSerializer.LinksSerializer(makeMintscanAccountLink(s.Text), s.Text),
		)
	} else {
		log.Info().Msg("/clear-alias: args length == ''")
	}

	if err := writeMessage(text, w); err != nil {
		log.Error().Err(err).Msg("Could not send response to /clear-alias command")
	}
}

func (reporter *SlackReporter) processListAliasesCommand(s slack.SlashCommand, w http.ResponseWriter) {
	var sb strings.Builder
	sb.WriteString(reporter.MarkdownSerializer.StrongSerializer("Wallet aliases:") + "\n")

	if len(labelsConfigManager.config.WalletLabels) == 0 {
		sb.WriteString(fmt.Sprintf(
			"No label aliases are set. You can set one using `%s` &lt;wallet-address&gt;",
			reporter.SlackSetAliasCommand,
		))
	}

	for key, value := range labelsConfigManager.config.WalletLabels {
		sb.WriteString(fmt.Sprintf(
			"• %s: %s\n",
			reporter.MarkdownSerializer.LinksSerializer(makeMintscanAccountLink(key), key),
			reporter.MarkdownSerializer.CodeSerializer(value),
		))
	}

	if err := writeMessage(sb.String(), w); err != nil {
		log.Error().Err(err).Msg("Could not send response to /clear-alias command")
	}
}

func writeMessage(text string, w http.ResponseWriter) error {
	params := &slack.Msg{
		Text:         text,
		ResponseType: "in_channel",
	}
	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	return err

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
		slack.MsgOptionDisableLinkUnfurl(),
	)
	return err
}

func (r SlackReporter) Name() string {
	return "SlackReporter"
}
