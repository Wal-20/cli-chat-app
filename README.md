CLI Chat — User Guide

What Is This?
- A simple terminal chat app. You install a small client that connects to your organization’s chat server.
- You’ll need the server address from your admin (example: https://chat.example.com).

What You Need
- The CLI Chat client for your device (Windows, macOS, or Linux)
- The server URL provided by your admin

Quick Start (Recommended)
- Windows
  - Download the client for Windows (you’ll receive a link from your admin).
  - Put the file anywhere (e.g., Downloads or Desktop).
  - Create a file named `.env` in the same folder with one line:
    SERVER_URL=https://chat.example.com
  - Double‑click the client to start. Or run it from a terminal.
  - Tip: You can set it permanently via PowerShell: `setx SERVER_URL "https://chat.example.com"` and then run the app.

- macOS
  - Download the macOS client (Intel or Apple Silicon, as provided by your admin).
  - Move it to a folder you control (e.g., `~/Applications`) and allow it to run:
    - In a terminal: `chmod +x ./cli-chat` (replace name as received)
    - If macOS blocks it, right‑click → Open once; or run: `xattr -d com.apple.quarantine ./cli-chat`
  - Set the server address either by placing a `.env` next to the app:
    SERVER_URL=https://chat.example.com
    or by exporting it in your shell: `export SERVER_URL=https://chat.example.com`
  - Start the app: `./cli-chat`

- Linux
  - Download the Linux client for your CPU (x86_64/arm64; link from your admin).
  - Make it executable and put it on your PATH:
    chmod +x ./cli-chat
    mkdir -p ~/.local/bin
    mv ./cli-chat ~/.local/bin/
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.profile
    source ~/.profile
  - Set the server address (either way works):
    - One‑time in the app folder: create `.env` with `SERVER_URL=https://chat.example.com`
    - Or permanent in your shell: `echo 'export SERVER_URL="https://chat.example.com"' >> ~/.profile && source ~/.profile`
  - Start the app: `cli-chat`

One‑Line Install (If Provided)
- Your admin may give you a one‑liner that installs or runs the client automatically. Examples:
  - Linux/macOS install: `curl -fsSL https://get.cli-chat.example/install.sh | bash`
  - Linux/macOS run (no install): `bash <(curl -fsSL https://get.cli-chat.example/run)`
  - Windows (PowerShell): `iwr -useb https://get.cli-chat.example/install.ps1 | iex`
- Only run these if you trust the source. These commands download and execute code on your machine.

First Run
- When the app starts, enter a username and password.
- If your username doesn’t exist yet, the app creates it for you.
- You’ll stay signed in; your login info is stored securely at:
  - Windows: `C:\Users\YOUR_NAME\.cli-chat-config.json`
  - macOS/Linux: `~/.cli-chat-config.json`

Switching Servers
- Change the server by editing the `.env` file next to the app or updating the `SERVER_URL` environment variable, then restart the app.

Uninstall
- Delete the app file you downloaded.
- Remove your saved login if you want to sign out everywhere:
  - Windows: delete `C:\Users\YOUR_NAME\.cli-chat-config.json`
  - macOS/Linux: `rm ~/.cli-chat-config.json`
- If you set `SERVER_URL` permanently, remove it:
  - Windows (PowerShell): `reg delete "HKCU\Environment" /v SERVER_URL /f`
  - macOS/Linux: remove the `export SERVER_URL=...` line from your shell profile (`~/.profile`, `~/.zshrc`, etc.)

Troubleshooting
- “SERVER_URL not set; create a .env…”: Create a `.env` file next to the app with `SERVER_URL=https://chat.example.com` or set the environment variable as shown above.
- “Can’t connect to server”: Check your internet and confirm the exact server URL with your admin. You can test in a browser: visit `https://chat.example.com/api/health`.

Privacy & Security
- Your login token is stored locally on your device at the paths listed above and is only used to connect to your chat server.
- Don’t share or upload your `.cli-chat-config.json` file.
