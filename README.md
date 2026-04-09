# notify

A simple CLI tool to send Slack DMs from the terminal. Perfect for chaining after long-running commands.

```sh
make build && notify "Build done" || notify "Build FAILED"
```

## Install

### Using `go install` (requires Go)

```sh
go install github.com/sveinungkb/notify-cli@latest
```

### Build from source

```sh
git clone https://github.com/sveinungkb/notify-cli.git
cd notify-cli
go build -o notify .
mv notify /usr/local/bin/   # or ~/bin/, anywhere on your PATH
```

## Setup

### 1. Create the Slack app

The easiest way is to use the included app manifest:

1. Go to [api.slack.com/apps](https://api.slack.com/apps) → **Create New App** → **From an app manifest**
2. Select your workspace
3. Paste the contents of [`slack-app-manifest.yml`](./slack-app-manifest.yml) and click **Next** → **Create**
4. Click **Install to Workspace** → **Allow**
5. Copy the **Bot User OAuth Token** (`xoxb-...`) from the **OAuth & Permissions** page

### 2. Configure notify

```sh
notify config
```

You'll be prompted for:
1. Your Slack token (hidden input)
2. A default recipient username (optional — press Enter to skip)

Config is saved to `~/.notify`.

## Usage

```sh
# Send to your default recipient
notify "Build complete"

# Send to a specific user
notify sveinungkb "Deployment done"

# Chain with commands (message on success or failure)
make build && notify "Build succeeded" || notify "Build FAILED"
```

Usernames can optionally include a leading `@`.

---

*100% vibed into existence with [Claude](https://claude.ai). Free for anyone to fork and use.*
