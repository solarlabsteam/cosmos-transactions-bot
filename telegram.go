package main

import (
	"fmt"
	"strings"
	"time"

	telegramBot "gopkg.in/tucnak/telebot.v2"
)

type TelegramReporter struct {
	TelegramToken string
	TelegramChat  int

	TelegramSetAliasCommand    string
	TelegramClearAliasCommand  string
	TelegramListAliasesCommand string

	TelegramBot    *telegramBot.Bot
	HtmlSerializer Serializer
	CacheManager   *CacheManager
}

func (r TelegramReporter) Serialize(report Report) string {
	var sb strings.Builder

	sb.WriteString(report.Tx.Serialize(r.Serializer()) + "\n\n")

	for _, msg := range report.Msgs {
		sb.WriteString(msg.Serialize(r.Serializer()) + "\n\n")
	}

	return sb.String()
}

func (r *TelegramReporter) Init() {
	if r.TelegramToken == "" || r.TelegramChat == 0 {
		log.Debug().Msg("Telegram credentials not set, not creating Telegram reporter.")
		return
	}

	bot, err := telegramBot.NewBot(telegramBot.Settings{
		Token:  TelegramToken,
		Poller: &telegramBot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Warn().Err(err).Msg("Could not create Telegram bot")
		return
	}

	r.TelegramBot = bot
	r.HtmlSerializer = Serializer{
		LinksSerializer: func(address string, text string) string {
			return fmt.Sprintf(`<a href="%s">%s</a>`, address, text)
		},
		StrongSerializer: func(text string) string {
			return fmt.Sprintf(`<strong>%s</strong>`, text)
		},
		CodeSerializer: func(text string) string {
			return fmt.Sprintf(`<code>%s</code>`, text)
		},
		MultilineCodeSerializer: func(text string) string {
			return fmt.Sprintf(`<pre>%s</pre>`, text)
		},
		CacheManager: r.CacheManager,
	}

	r.TelegramBot.Handle(r.TelegramSetAliasCommand, r.processSetAliasCommand)
	r.TelegramBot.Handle(r.TelegramClearAliasCommand, r.processClearAliasCommand)
	r.TelegramBot.Handle(r.TelegramListAliasesCommand, r.processListAliasesCommand)
	go r.TelegramBot.Start()
}

func (reporter *TelegramReporter) processSetAliasCommand(message *telegramBot.Message) {
	reporter.logQuery(message, reporter.TelegramSetAliasCommand)

	text := fmt.Sprintf("Usage: `%s` &lt;wallet-address&gt; &lt;alias&gt;", reporter.TelegramSetAliasCommand)

	args := strings.SplitN(message.Text, " ", 3)

	if len(args) > 2 {
		labelsConfigManager.setWalletLabel(args[1], args[2])
		text = fmt.Sprintf(
			"Successfully set alias for %s: %s",
			reporter.HtmlSerializer.LinksSerializer(makeMintscanAccountLink(args[1]), args[1]),
			reporter.HtmlSerializer.CodeSerializer(args[2]),
		)
	} else {
		log.Info().Msg("/set-alias: args length <= 2")
	}

	if err := reporter.sendMessage(message, text); err != nil {
		log.Error().Err(err).Msg("Could not send response to /set-alias command")
	}
}

func (reporter *TelegramReporter) processClearAliasCommand(message *telegramBot.Message) {
	reporter.logQuery(message, reporter.TelegramClearAliasCommand)

	text := fmt.Sprintf("Usage: `%s` &lt;wallet-address&gt;", reporter.TelegramClearAliasCommand)

	args := strings.SplitN(message.Text, " ", 2)

	if len(args) >= 2 {
		labelsConfigManager.clearWalletLabel(args[1])
		text = fmt.Sprintf(
			"Successfully cleared alias for %s",
			reporter.HtmlSerializer.LinksSerializer(makeMintscanAccountLink(args[1]), args[1]),
		)
	} else {
		log.Info().Msg("/clear-alias: args length < 2")
	}

	if err := reporter.sendMessage(message, text); err != nil {
		log.Error().Err(err).Msg("Could not send response to /clear-alias command")
	}
}

func (reporter *TelegramReporter) processListAliasesCommand(message *telegramBot.Message) {
	reporter.logQuery(message, reporter.TelegramListAliasesCommand)

	var sb strings.Builder
	sb.WriteString(reporter.HtmlSerializer.StrongSerializer("Wallet aliases:") + "\n")

	if len(labelsConfigManager.config.WalletLabels) == 0 {
		sb.WriteString(fmt.Sprintf(
			"No label aliases are set. You can set one using `%s` &lt;wallet-address&gt;",
			reporter.TelegramSetAliasCommand,
		))
	}

	for key, value := range labelsConfigManager.config.WalletLabels {
		sb.WriteString(fmt.Sprintf(
			"• %s: %s\n",
			reporter.HtmlSerializer.LinksSerializer(makeMintscanAccountLink(key), key),
			reporter.HtmlSerializer.CodeSerializer(value),
		))
	}

	if err := reporter.sendMessage(message, sb.String()); err != nil {
		log.Error().Err(err).Msg("Could not send response to /clear-alias command")
	}
}

func (reporter *TelegramReporter) logQuery(message *telegramBot.Message, command string) {
	log.Info().
		Str("command", command).
		Str("text", message.Text).
		Str("user", message.Sender.Username).
		Msg("Received command")
}

func (reporter *TelegramReporter) sendMessage(message *telegramBot.Message, text string) error {
	_, err := reporter.TelegramBot.Send(
		message.Chat,
		text,
		&telegramBot.SendOptions{
			ParseMode: telegramBot.ModeHTML,
			ReplyTo:   message,
		},
	)

	return err
}

func (r TelegramReporter) Enabled() bool {
	return r.TelegramBot != nil
}

func (r TelegramReporter) Serializer() Serializer {
	return r.HtmlSerializer
}

func (r TelegramReporter) SendReport(report Report) error {
	serializedReport := r.Serialize(report)
	_, err := r.TelegramBot.Send(
		&telegramBot.User{
			ID: r.TelegramChat,
		},
		serializedReport,
		telegramBot.ModeHTML,
		telegramBot.NoPreview,
	)
	return err
}

func (r TelegramReporter) Name() string {
	return "TelegramReporter"
}
