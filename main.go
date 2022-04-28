package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	json "github.com/tendermint/tendermint/libs/json"

	"github.com/tendermint/tendermint/crypto/tmhash"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	jsonRpcTypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	events "github.com/tendermint/tendermint/types"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	ConfigPath       string
	LabelsConfigPath string

	LogLevel          string
	Queries           []string
	MintscanProject   string
	CoingeckoCurrency string

	TelegramToken              string
	TelegramChat               int
	TelegramSetAliasCommand    string
	TelegramClearAliasCommand  string
	TelegramListAliasesCommand string

	SlackToken              string
	SlackChat               string
	SlackSigningSecret      string
	SlackListenAddress      string
	SlackSetAliasCommand    string
	SlackClearAliasCommand  string
	SlackListAliasesCommand string

	NodeAddress          string
	TendermintRpcAddress string

	BaseDenom        string
	Denom            string
	DenomCoefficient float64

	Printer = message.NewPrinter(language.English)

	reporters []Reporter

	client *tmclient.WSClient

	log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	SentTransactions map[string]bool = make(map[string]bool)

	labelsConfigManager *LabelsConfigManager
	cacheManager        *CacheManager
	grpcWrapper         *GrpcWrapper
	coingeckoWrapper    *CoingeckoWrapper
)

var rootCmd = &cobra.Command{
	Use:  "cosmos-transactions-bot",
	Long: "Tool to notify about the new transactions",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if ConfigPath == "" {
			return nil
		}

		viper.SetConfigFile(ConfigPath)
		if err := viper.ReadInConfig(); err != nil {
			log.Info().Err(err).Msg("Error reading config file")
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}

		// Credits to https://carolynvanslyck.com/blog/2020/08/sting-of-the-viper/
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if !f.Changed && viper.IsSet(f.Name) {
				val := viper.Get(f.Name)

				// array values
				if sliceVal, ok := f.Value.(pflag.SliceValue); ok {
					log.Trace().Str("name", f.Name).Msg("Treating flag as slice value")

					aInterface, ok := val.([]interface{})
					if !ok {
						log.Fatal().
							Str("name", f.Name).
							Msg("Could not parse Viper value as array. Probably you've declared the value as not array?")
					}

					aString := make([]string, len(aInterface))
					for i, v := range aInterface {
						aString[i] = v.(string)
					}

					if err := sliceVal.Replace(aString); err != nil {
						log.Fatal().
							Err(err).
							Msg("Could not replace value")
					}
					return
				}

				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					log.Fatal().Err(err).Msg("Could not set flag")
				}
			}
		})

		return nil
	},
	Run: Execute,
}

func Execute(cmd *cobra.Command, args []string) {
	logLevel, err := zerolog.ParseLevel(LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not parse log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	grpcWrapper = InitGrpcWrapper(NodeAddress)
	grpcWrapper.setDenom()
	defer grpcWrapper.CloseConnection()

	labelsConfigManager = initLabelsConfig(LabelsConfigPath)
	coingeckoWrapper = NewCoingeckoWrapper(CoingeckoCurrency)
	cacheManager = NewCacheManager(grpcWrapper, coingeckoWrapper)

	reporters = []Reporter{
		&TelegramReporter{
			TelegramToken:              TelegramToken,
			TelegramChat:               TelegramChat,
			TelegramSetAliasCommand:    TelegramSetAliasCommand,
			TelegramClearAliasCommand:  TelegramClearAliasCommand,
			TelegramListAliasesCommand: TelegramListAliasesCommand,
			CacheManager:               cacheManager,
		},
		&SlackReporter{
			SlackToken:              SlackToken,
			SlackChat:               SlackChat,
			SlackSigningSecret:      SlackSigningSecret,
			SlackListenAddress:      SlackListenAddress,
			SlackSetAliasCommand:    SlackSetAliasCommand,
			SlackClearAliasCommand:  SlackClearAliasCommand,
			SlackListAliasesCommand: SlackListAliasesCommand,
			CacheManager:            cacheManager,
		},
	}

	for _, reporter := range reporters {
		reporter.Init()

		if reporter.Enabled() {
			log.Info().Str("name", reporter.Name()).Msg("Init reporter")
		}
	}

	client, err = tmclient.NewWS(
		TendermintRpcAddress,
		"/websocket",
		tmclient.PingPeriod(5*time.Second),
		tmclient.OnReconnect(func() {
			log.Info().Msg("Reconnected to websocket...")
			subscribeToUpdates()
		}),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create a client")
		os.Exit(1)
	}

	err = client.Start()
	if err != nil {
		log.Error().Err(err).Msg("Failed to start a client")
		os.Exit(1)
	}
	defer client.Stop() // nolint

	subscribeToUpdates()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case result := <-client.ResponsesCh:
			processResponse(result)
		case <-quit:
			os.Exit(0)
		}
	}
}

func subscribeToUpdates() {
	for _, query := range Queries {
		if err := client.Subscribe(context.Background(), query); err != nil {
			log.Fatal().Err(err).Str("query", query).Msg("Failed to subscribe to query")
		}

		log.Info().Str("query", query).Msg("Listening for incoming transactions")
	}
}

func processResponse(result jsonRpcTypes.RPCResponse) {
	report := generateReport(result)

	if report.Empty() {
		log.Info().Msg("Report is empty, not sending.")
		return
	}

	for _, reporter := range reporters {
		if !reporter.Enabled() {
			log.Debug().Str("name", reporter.Name()).Msg("Reporter is disabled.")
			continue
		}

		log.Info().Str("name", reporter.Name()).Msg("Sending a report to reporter...")
		if err := reporter.SendReport(report); err != nil {
			log.Error().Err(err).Str("name", reporter.Name()).Msg("Could not send message")
		}
	}

	cacheManager.clearCache()
}

func generateReport(result jsonRpcTypes.RPCResponse) Report {
	report := Report{
		Msgs: []Msg{},
	}

	if result.Error != nil && result.Error.Message != "" {
		log.Error().Str("msg", result.Error.Error()).Msg("Got error in RPC response")
		return Report{}
	}

	var resultEvent ctypes.ResultEvent
	if err := json.Unmarshal(result.Result, &resultEvent); err != nil {
		log.Error().Err(err).Msg("Failed to parse event")
		return Report{}
	}

	if resultEvent.Data == nil {
		log.Debug().Msg("Event does not have data, skipping.")
		return Report{}
	}

	txResult := resultEvent.Data.(events.EventDataTx).TxResult
	txHash := fmt.Sprintf("%X", tmhash.Sum(txResult.Tx))
	var tx tx.Tx

	if err := proto.Unmarshal(txResult.Tx, &tx); err != nil {
		log.Error().Err(err).Msg("Could not parse tx")
	}

	txMessages := tx.GetBody().GetMessages()
	report.Tx = parseTx(txResult)

	if _, ok := SentTransactions[txHash]; ok {
		log.Debug().Str("hash", txHash).Msg("Transaction already sent, skipping.")
		return Report{}
	}

	log.Info().
		Int64("height", txResult.Height).
		Str("memo", tx.GetBody().GetMemo()).
		Str("hash", txHash).
		Int("len", len(txMessages)).
		Msg("Got transaction")

	for _, message := range txMessages {
		var msg Msg

		switch message.TypeUrl {
		case "/cosmos.bank.v1beta1.MsgSend":
			msg = ParseMsgSend(message)
		case "/cosmos.gov.v1beta1.MsgVote":
			msg = ParseMsgVote(message)
		case "/cosmos.gov.v1beta1.MsgSubmitProposal":
			msg = ParseMsgSubmitProposal(message, txResult.Height)
		case "/cosmos.staking.v1beta1.MsgDelegate":
			msg = ParseMsgDelegate(message)
		case "/cosmos.staking.v1beta1.MsgUndelegate":
			msg = ParseMsgUndelegate(message)
		case "/cosmos.staking.v1beta1.MsgBeginRedelegate":
			msg = ParseMsgBeginRedelegate(message)
		case "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress":
			msg = ParseMsgSetWithdrawAddress(message)
		case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
			msg = ParseMsgWithdrawDelegatorReward(message, txResult.Height)
		case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission":
			msg = ParseMsgWithdrawValidatorCommission(message, txResult.Height)
		case "/ibc.applications.transfer.v1.MsgTransfer":
			msg = ParseMsgIbcTransfer(message)
		case "/ibc.core.channel.v1.MsgRecvPacket":
			msg = ParseMsgIbcRecvPacket(message)
		default:
			log.Warn().Str("type", message.TypeUrl).Msg("Got a message which is not supported")
		}

		if msg != nil && !msg.Empty() {
			report.Msgs = append(report.Msgs, msg)
		}
	}

	SentTransactions[txHash] = true

	return report
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	rootCmd.PersistentFlags().StringVar(&LabelsConfigPath, "labels-config", "", "Labels config file path")
	rootCmd.PersistentFlags().StringVar(&BaseDenom, "base-denom", "", "Cosmos coin base denom (like uatom)")
	rootCmd.PersistentFlags().StringVar(&Denom, "denom", "", "Cosmos coin display denom (like atom)")
	rootCmd.PersistentFlags().Float64Var(&DenomCoefficient, "denom-coefficient", 1_000_000, "Denom coefficient from base denom to display denom")
	rootCmd.PersistentFlags().StringVar(&LogLevel, "log-level", "info", "Logging level")
	rootCmd.PersistentFlags().StringSliceVar(&Queries, "query", []string{"tx.height > 1"}, "Tx filter to subscribe to")

	rootCmd.PersistentFlags().StringVar(&TelegramToken, "telegram-token", "", "Telegram bot token")
	rootCmd.PersistentFlags().IntVar(&TelegramChat, "telegram-chat", 0, "Telegram chat or user ID")
	rootCmd.PersistentFlags().StringVar(&TelegramSetAliasCommand, "telegram-set-alias-command", "/set_alias", "Telegram slash command to set alias")
	rootCmd.PersistentFlags().StringVar(&TelegramClearAliasCommand, "telegram-clear-alias-command", "/clear_alias", "Telegram slash command to clear alias")
	rootCmd.PersistentFlags().StringVar(&TelegramListAliasesCommand, "telegram-list-aliases-command", "/list_aliases", "Telegram slash command to list aliases")

	rootCmd.PersistentFlags().StringVar(&SlackToken, "slack-token", "", "Slack bot token")
	rootCmd.PersistentFlags().StringVar(&SlackChat, "slack-chat", "", "Slack chat or user ID")
	rootCmd.PersistentFlags().StringVar(&SlackSigningSecret, "slack-signing-secret", "", "Slack signing secret for slash commands handling")
	rootCmd.PersistentFlags().StringVar(&SlackListenAddress, "slack-listen-address", ":9500", "An address where Slack slash command handler would be exposed at")
	rootCmd.PersistentFlags().StringVar(&SlackSetAliasCommand, "slack-set-alias-command", "/set-alias", "Slack slash command to set alias")
	rootCmd.PersistentFlags().StringVar(&SlackClearAliasCommand, "slack-clear-alias-command", "/clear-alias", "Slack slash command to clear alias")
	rootCmd.PersistentFlags().StringVar(&SlackListAliasesCommand, "slack-list-aliases-command", "/list-aliases", "Slack slash command to list aliases")

	rootCmd.PersistentFlags().StringVar(&MintscanProject, "mintscan-project", "cosmos", "mintscan.io/* project to generate links to")
	rootCmd.PersistentFlags().StringVar(&CoingeckoCurrency, "coingecko-currency", "", "Coingecko currency name")
	rootCmd.PersistentFlags().StringVar(&NodeAddress, "node", "localhost:9090", "RPC node address")
	rootCmd.PersistentFlags().StringVar(&TendermintRpcAddress, "tendermint-rpc", "tcp://localhost:26657", "Tendermint RPC node address")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
