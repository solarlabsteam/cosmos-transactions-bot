package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	json "github.com/tendermint/tendermint/libs/json"
	"google.golang.org/grpc"

	"github.com/tendermint/tendermint/crypto/tmhash"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	jsonRpcTypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	events "github.com/tendermint/tendermint/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	ConfigPath string

	LogLevel        string
	Query           string
	MintscanProject string

	TelegramToken string
	TelegramChat  int
	SlackToken    string
	SlackChat     string

	NodeAddress string

	Denom            string
	DenomCoefficient float64

	Printer = message.NewPrinter(language.English)

	reporters []Reporter

	log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
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

	reporters = []Reporter{
		&TelegramReporter{
			TelegramToken: TelegramToken,
			TelegramChat:  TelegramChat,
		},
		&SlackReporter{
			SlackToken: SlackToken,
			SlackChat:  SlackChat,
		},
	}

	for _, reporter := range reporters {
		log.Info().Str("name", reporter.Name()).Msg("Init reporter")
		reporter.Init()
	}

	client, err := tmclient.NewWS("tcp://localhost:26657", "/websocket")
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

	err = client.Subscribe(context.Background(), Query)
	if err != nil {
		log.Error().Err(err).Str("query", Query).Msg("Failed to subscribe to query")
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	log.Info().Str("query", Query).Msg("Listening for incoming transactions")

	for {
		select {
		case result := <-client.ResponsesCh:
			processResponse(result)
		case <-quit:
			os.Exit(0)
		}
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
}

func generateReport(result jsonRpcTypes.RPCResponse) Report {
	report := Report{
		Msgs: []Msg{},
	}

	var resultEvent ctypes.ResultEvent
	if err := json.Unmarshal(result.Result, &resultEvent); err != nil {
		log.Error().Err(err).Msg("Failed to parse event")
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
		case "/cosmos.staking.v1beta1.MsgDelegate":
			msg = ParseMsgDelegate(message)
		case "/cosmos.staking.v1beta1.MsgUndelegate":
			msg = ParseMsgUndelegate(message)
		case "/cosmos.staking.v1beta1.MsgBeginRedelegate":
			msg = ParseMsgBeginRedelegate(message)
		case "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress":
			msg = ParseMsgSetWithdrawAddress(message)
		case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
			msg = ParseMsgWithdrawDelegatorReward(message)
		case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission":
			msg = ParseMsgWithdrawValidatorCommission(message)
		default:
			log.Warn().Str("type", message.TypeUrl).Msg("Got a message which is not supported")
		}

		if msg != nil && !msg.Empty() {
			report.Msgs = append(report.Msgs, msg)
		}
	}

	return report
}

func setDenom() {
	grpcConn, err := grpc.Dial(
		NodeAddress,
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}

	defer grpcConn.Close()

	bankClient := banktypes.NewQueryClient(grpcConn)
	denoms, err := bankClient.DenomsMetadata(
		context.Background(),
		&banktypes.QueryDenomsMetadataRequest{},
	)

	if err != nil {
		log.Fatal().Err(err).Msg("Error querying denom")
	}

	metadata := denoms.Metadatas[0] // always using the first one
	if Denom == "" {                // using display currency
		Denom = metadata.Display
	}

	for _, unit := range metadata.DenomUnits {
		log.Debug().
			Str("denom", unit.Denom).
			Uint32("exponent", unit.Exponent).
			Msg("Denom info")
		if unit.Denom == Denom {
			DenomCoefficient = math.Pow10(int(unit.Exponent))
			log.Info().
				Str("denom", Denom).
				Float64("coefficient", DenomCoefficient).
				Msg("Got denom info")
			return
		}
	}

	log.Fatal().Msg("Could not find the denom info")
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	rootCmd.PersistentFlags().StringVar(&Denom, "denom", "", "Cosmos coin denom")
	rootCmd.PersistentFlags().StringVar(&LogLevel, "log-level", "info", "Logging level")
	rootCmd.PersistentFlags().StringVar(&Query, "query", "tx.height > 1", "Tx filter to subscribe to")
	rootCmd.PersistentFlags().StringVar(&TelegramToken, "telegram-token", "", "Telegram bot token")
	rootCmd.PersistentFlags().IntVar(&TelegramChat, "telegram-chat", 0, "Telegram chat or user ID")
	rootCmd.PersistentFlags().StringVar(&SlackToken, "slack-token", "", "Slack bot token")
	rootCmd.PersistentFlags().StringVar(&SlackChat, "slack-chat", "", "Slack chat or user ID")
	rootCmd.PersistentFlags().StringVar(&MintscanProject, "mintscan-project", "crypto-org", "mintscan.io/* project to generate links to")
	rootCmd.PersistentFlags().StringVar(&NodeAddress, "node", "localhost:9090", "RPC node address")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
