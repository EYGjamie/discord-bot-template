# Discord Bot Template

A template project for building a Discord bot with a web application and API backend.

## Project Structure

```
discord-bot-template/
├── cmd/
│   ├── api/          # Web API Backend
│   └── bot/          # Discord Bot
├── internal/
│   ├── bot/          # Bot-Logik (Commands, Events, Handlers)
│   ├── database/     # Datenbank-Layer
│   ├── server/       # API-Server
│   └── shared/       # Gemeinsamer Code (Bot + API)
├── frontend/         # React Frontend
└── configs/          # Konfigurationsdateien
```

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## MakeFile

Run build make command with tests
```bash
make all
```

Build the API
```bash
make build
```

Build the Discord Bot
```bash
make build-bot
```

Run the API + Frontend
```bash
make run
```

Run the Discord Bot
```bash
make run-bot
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binaries from the last build:
```bash
make clean
```

## Discord Bot Setup

1. Erstelle eine Discord-Anwendung auf https://discord.com/developers/applications
2. Kopiere den Bot-Token
3. Erstelle eine `.env` Datei basierend auf `configs/bot.example.env`
4. Füge deinen Token hinzu: `DISCORD_TOKEN=dein_token_hier`
5. Starte den Bot mit `make run-bot`

## Environment Variables

Siehe `configs/bot.example.env` für alle verfügbaren Umgebungsvariablen.
