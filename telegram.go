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

	TelegramBot    *telegramBot.Bot
	HtmlSerializer Serializer
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
	}
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
	)
	return err
}

func (r TelegramReporter) Name() string {
	return "TelegramReporter"
}
