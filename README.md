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
3. Erstelle eine `.env` Datei basierend auf `.env.example`
4. Füge deinen Token hinzu: `DEV_DISCORD_BOT_TOKEN=dein_token_hier`
5. Starte den Bot mit `make run-bot` oder `docker-compose up bot`

## Bot Commands

Der Bot verwendet Discord Slash-Commands. Alle Commands sind nach dem Start des Bots automatisch verfügbar.

### 🛡️ Moderation Commands

#### `/moderation`
Verwaltung der Moderations-Einstellungen (Nur für Administratoren)

**Subcommands:**

##### `set-channel`
Legt den Kanal fest, in dem Moderations-Logs gepostet werden.

```
/moderation set-channel channel:#mod-logs
```

**Parameter:**
- `channel` (Erforderlich): Der Discord-Kanal für Moderations-Logs

**Beispiel:**
```
/moderation set-channel channel:#moderation-logs
```

---

##### `toggle-edits`
Aktiviert oder deaktiviert das Logging von bearbeiteten Nachrichten.

```
/moderation toggle-edits enabled:true
```

**Parameter:**
- `enabled` (Erforderlich): `true` = aktiviert, `false` = deaktiviert

**Was wird geloggt:**
- Autor der Nachricht
- Original-Inhalt (vorher)
- Neuer Inhalt (nachher)
- Timestamp der Bearbeitung
- Link zur Nachricht

**Beispiel:**
```
/moderation toggle-edits enabled:true
```

---

##### `toggle-deletes`
Aktiviert oder deaktiviert das Logging von gelöschten Nachrichten.

```
/moderation toggle-deletes enabled:true
```

**Parameter:**
- `enabled` (Erforderlich): `true` = aktiviert, `false` = deaktiviert

**Was wird geloggt:**
- Autor der Nachricht
- Nachrichteninhalt
- Anhänge (falls vorhanden)
- Timestamp der Löschung
- Kanal, in dem die Nachricht war

**Beispiel:**
```
/moderation toggle-deletes enabled:false
```

---

##### `status`
Zeigt die aktuellen Moderations-Einstellungen an.

```
/moderation status
```

**Zeigt:**
- Aktueller Log-Kanal
- Status von Edit-Logging (an/aus)
- Status von Delete-Logging (an/aus)

**Beispiel-Output:**
```
📊 Moderations-Einstellungen

Log-Kanal: #mod-logs
✅ Message-Edit-Logging: Aktiviert
✅ Message-Delete-Logging: Aktiviert
```

---

### ⚙️ Features

#### Dynamische Konfiguration
Alle Bot-Einstellungen können **ohne Neustart** im laufenden Betrieb geändert werden. Änderungen werden in der Datenbank gespeichert und sind persistent.

#### Automatisches User-Tracking
Der Bot tracked automatisch alle User in der Datenbank:
- Bei Server-Beitritt (Member Join)
- Bei Profil-Änderungen (Member Update)
- Bei Nachrichten
- Bei Server-Verlassen (Member Remove)

#### Message-Logging
- **Bearbeitete Nachrichten**: Zeigt Original und bearbeiteten Text
- **Gelöschte Nachrichten**: Speichert Inhalt bevor die Nachricht verschwindet
- **Bot-Nachrichten werden ignoriert**: Nur User-Nachrichten werden geloggt

---

### 🔒 Berechtigungen

#### Administrator Commands
Folgende Commands erfordern **Administrator-Rechte**:
- `/moderation set-channel`
- `/moderation toggle-edits`
- `/moderation toggle-deletes`

Jeder Nutzer kann den Status einsehen:
- `/moderation status`

---

## Environment Variables

Siehe `.env.example` für alle verfügbaren Umgebungsvariablen:

```env
# Application Environment
APP_ENV=development

# Discord Bot Tokens
DEV_DISCORD_BOT_TOKEN=your_dev_bot_token_here
PROD_DISCORD_BOT_TOKEN=your_prod_bot_token_here

# API Port
PORT=8080

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_DATABASE=discord_bot
DB_USERNAME=postgres
DB_PASSWORD=your_secure_password_here
DB_SCHEMA=public
```

## Docker Deployment

```bash
# Bot + Database starten
docker-compose up bot

# Alle Services starten (Bot, API, Database)
docker-compose up

# Im Hintergrund starten
docker-compose up -d

# Logs anzeigen
docker-compose logs -f bot

# Stoppen
docker-compose down
```