module main

go 1.16

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cosmos/cosmos-sdk v0.45.1
	github.com/cosmos/ibc-go v1.4.0
	github.com/gogo/protobuf v1.3.3
	github.com/rs/zerolog v1.23.0
	github.com/slack-go/slack v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/superoo7/go-gecko v1.0.0
	github.com/tendermint/tendermint v0.34.14
	golang.org/x/text v0.3.6
	google.golang.org/grpc v1.42.0
	gopkg.in/tucnak/telebot.v2 v2.3.5
)
