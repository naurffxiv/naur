# clearingway

Clearingway listens to a Discord channel for the message `/clears <world> <first-name> <last-name>`.

When it hears this message, it tries to find the relevant character in the Lodestone. If it finds them, it then parses their
fflogs and tries to assign them a few roles:

1. A role for the highest current parse they have in any relevant encounter (Gold, Orange, Purple, Blue, Green, Grey)
2. A role for every relevant encounter they've cleared ("P1S-Cleared," "P2S-Cleared," "P3S-Cleared," etc.)
3. A combo legend role purely for flexing purposes for every ultimate they've cleared (The Legend, The Double Legend, The Triple Legend, The Tetra Legend)
4. The role "NA's Comfiest" if they have any relevant encounter clears with a parse between 0 and 0.9
5. The role "The Nice Legend" if they have any ultimate clears with a parse between 69.0 and 69.9
6. The role "Nice" if they have any relevant clears with a parse between 69.0 and 69.9
7. The role "The Comfy Legend" if they have any ultimate clears with a parse between 0 and 0.9

It can be configured with the `config.yaml` file found in this repository.

## Development

Clearingway runs in Docker for local development. You only need [Docker](https://docs.docker.com/get-docker/) and Docker Compose installed.

### Setup

1. Copy `.env.example` to `.env` and fill in your credentials:

   * **DISCORD_TOKEN**: Create a [Discord bot for Clearingway](https://discord.com/developers/applications) and add the bot token here.
   * **FFLOGS_CLIENT_ID**: The client ID from [fflogs](https://www.fflogs.com/api/clients/).
   * **FFLOGS_CLIENT_SECRET**: The client secret from [fflogs](https://www.fflogs.com/api/clients/).

2. Start the bot:

   ```bash
   make dev
   ```

   This builds the image and runs the bot in the foreground so logs appear in your terminal.

### Common commands

| Command | Description |
| --- | --- |
| `make dev` | Build and run in the foreground |
| `make up` | Build and run in the background |
| `make down` | Stop the bot |
| `make logs` | Follow container logs |
| `make watch` | Rebuild and restart automatically when code changes |
| `make build` | Build the image without starting |
| `make test` | Run the test suite |

`config.yaml` is mounted into the container, so you can edit guild configuration without rebuilding the image. After Go code changes, run `make dev` again or use `make watch` to rebuild automatically.
