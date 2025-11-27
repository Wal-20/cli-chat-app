# CLI Chat

A terminal chat client with a Go REST/WebSocket backend.  

## Features

- Keyboard-driven TUI
- Messaging via REST and WebSockets
- Login, create/join rooms, manage members
- Persistent sessions (`~/.cli-chat-config.json`)

## Install

```bash
# linux, macos-amd64, macos-arm64
curl -L -o chat-cli https://cli-chat.duckdns.org/download/{target}
chmod +x chat-cli
./chat-cli

# windows
curl.exe -L -o chat-cli.exe https://cli-chat.duckdns.org/download/windows
.\chat-cli.exe

# Point the client at a different server
export SERVER_URL="http://your-server:8080"
./chat-cli
```

## Usage

1. Launch the client.
2. Log in or register with a username/password.
3. Navigate rooms with the keyboard:
   - `Tab`: switch lists
   - `Enter`: open selected room
   - `c`: create room
   - `Ctrl+J`: join by ID
   - `n`: view notifications
   - `Ctrl+D`: delete owned room
4. Type messages and press `Enter` to send.

## Support

Open an issue or submit a PR in this repository if you run into problems or have feature requests.
