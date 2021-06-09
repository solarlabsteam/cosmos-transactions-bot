# cosmos-transactions-bot

![Latest release](https://img.shields.io/github/v/release/solarlabsteam/cosmos-transactions-bot)
[![Actions Status](https://github.com/solarlabsteam/cosmos-transactions-bot/workflows/test/badge.svg)](https://github.com/solarlabsteam/cosmos-transactions-bot/actions)

cosmos-transactions-bot is a tool that sends a message to configured channels on new transactions with a specific filter.

## How can I set it up?

Download the latest release from [the releases page](https://github.com/solarlabsteam/cosmos-transactions-bot/releases/). After that, you should unzip it and you are ready to go:

```sh
wget <the link from the releases page>
tar xvfz cosmos-transactions-bot_*
./cosmos-transactions-bot <params>
```

That's not really interesting, what you probably want to do is to have it running in the background. For that, first of all, we have to copy the file to the system apps folder:

```sh
sudo cp ./cosmos-transactions-bot /usr/bin
```

Then we need to create a systemd service for our app:

```sh
sudo nano /etc/systemd/system/cosmos-transactions-bot.service
```

You can use this template (change the user to whatever user you want this to be executed from. It's advised to create a separate user for that instead of running it from root):

```
[Unit]
Description=Cosmos Transactions Bot
After=network-online.target

[Service]
User=<username>
TimeoutStartSec=0
CPUWeight=95
IOWeight=95
ExecStart=cosmos-transactions-bot <params>
Restart=always
RestartSec=2
LimitNOFILE=800000
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

Then we'll add this service to the autostart and run it:

```sh
sudo systemctl enable cosmos-transactions-bot
sudo systemctl start cosmos-transactions-bot
sudo systemctl status cosmos-transactions-bot # validate it's running
```

If you need to, you can also see the logs of the process:

```sh
sudo journalctl -u cosmos-transactions-bot -f --output cat
```

## How does it work?

It subscribes to Tendermint JSON-RPC endpoint through Websockets (see [this](https://docs.tendermint.com/master/rpc/#/Websocket/subscribe) for more details). After that, once the new transaction with the specified filter is detected, the full node sends a Websocket message, and this program catches it and sends a message to a specified channel (or channels).

## How can I configure it?

You can pass the artuments to the executable file to configure it. Here is the parameters list:

- `--node` - the gRPC node URL. Defaults to `localhost:9090`
- `--log-devel` - logger level. Defaults to `info`. You can set it to `debug` to make it more verbose.
- `--telegram-token` - Telegram bot token
- `--telegram-chat` - Telegram user or chat ID
- `--slack-token` - Slack bot token
- `--slack-chat` - Slack user or chat ID
- `--mintscan-prefix` - This bot generates links to Mintscan for validators, using this prefix. Links have the following format: `https://mintscan.io/<mintscan-prefix>/validator/<validator ID>`. Defaults to `persistence`.
- `--query` - See below.


Additionally, you can pass a `--config` flag with a path to your config file (we use `.toml`, but anything supported by [viper](https://github.com/spf13/viper) should work).

### Query

You can specify a `--query` that serves as a filter. If the transaction does not this filter, this program won't send a notification on that. The default filter is `tx.height > 1`, which matches all transactions. You would probably want to use your own filter.

For example, we're using this tool to monitor new delegations for our validator and this is what we have in our `.toml` configuration file:

```
# sentvaloper1sazxkmhym0zcg9tmzvc4qxesqegs3q4u66tpmf is SOLAR Validator on Sentinel
query = "delegate.validator = 'sentvaloper1sazxkmhym0zcg9tmzvc4qxesqegs3q4u66tpmf'"
```

Unfortunately there is no OR operator support, so you cannot monitor different events. See [this](https://stackoverflow.com/questions/65709248/how-to-use-an-or-condition-with-the-tendermint-websocket-subscribe-method) and [this](https://github.com/tendermint/tendermint/issues/5206) for context. You can spawn a few instances of the app with different filters though.

See [the documentation](https://docs.tendermint.com/master/rpc/#/Websocket/subscribe) for more information.

## Notifications channels

Currently this program supports the following notifications channels:
1) Telegram

Go to @BotFather in Telegram and create a bot. After that, there are two options:
- you want to send messages to a user. This user should write a message to @getmyid_bot, then copy the `Your user ID` number. Also keep in mind that the bot won't be able to send messages unless you contact it first, so write a message to a bot before proceeding.
- you want to send messages to a channel. Write something to a channel, then forward it to @getmyid_bot and copy the `Forwarded from chat` number. Then add the bot as an admin.


Then run a program with `--telegram-token <token> --telegram-chat <chat ID>`.

2) Slack

Go to the Slack web interface -> Manage apps and create a new app.
Give the app the `chat:write` scope and add the integration to a channel by typing `/invite <bot username>` there.
After that, run the program with `--slack-token <token> --slack-chat <channel name>`.


## Which networks this is guaranteed to work?

In theory, it should work on a Cosmos-based blockchains that expose a gRPC endpoint.

## How can I contribute?

Bug reports and feature requests are always welcome! If you want to contribute, feel free to open issues or PRs.
