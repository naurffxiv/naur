# Clearingway
Clearingway is a Discord bot designed to analyse players' FFLogs data and provide relevant roles based on completion.

## Make Commands

`make test` - Runs all tests

`make build` - Builds the bot binary into `./bin/clearingway`

`make format` - Formats all Go files using `goimports`

`make lint` - Runs linter checks using `golangci-lint`

`make lint-fix` - Runs linter checks and attempts to fix issues using `golangci-lint`

## Environment Variables

| ENV                  | Description                                                          | Required | Default   |
|----------------------|----------------------------------------------------------------------|----------|-----------|
| CONFIG_PATH          | Relative path from project root to location of config files          | Yes      | ./configs |
| ENV                  | The environment the bot is running in: "production" or "development" | Yes      |           |
| DISCORD_TOKEN        | [Discord](https://discord.com/developers/docs/intro) Bot Token       | Yes      |           |
| FFLOGS_CLIENT_ID     | Client ID for [FFLogs](https://www.fflogs.com/profile)               | Yes      |           |
| FFLOGS_CLIENT_SECRET | Client ID for [FFLogs](https://www.fflogs.com/profile)               | Yes      |           |