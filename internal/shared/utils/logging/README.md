# Logging System

Dieses Logging-System bietet eine umfassende Lösung für das Logging im Discord Bot mit folgenden Features:

## Features

### 1. **Multi-Level Logging**
- `ERROR`: Für Fehler mit optionalem Stack Trace
- `WARN`: Für Warnungen
- `INFO`: Für Informationen
- `DEBUG`: Für Debug-Informationen

### 2. **Datenbank-Logging**
- Logs werden in der `logs` Tabelle in der PostgreSQL-Datenbank gespeichert
- Indexiert nach Level, Guild ID, Zeitstempel und Source
- Unterstützt Guild-spezifische Logs
- Automatisches Cleanup alter Logs mit `DeleteOldLogs()`

### 3. **File-Fallback**
- Bei DB-Ausfällen werden Logs automatisch in Dateien geschrieben
- Logs werden nach Level und Datum getrennt: `error_2025-12-15.log`, `info_2025-12-15.log`, etc.
- Dateien werden im `logs/` Verzeichnis gespeichert

### 4. **Notifications**
- **Error-Logs**: Senden IMMER eine Notification an registrierte User
- **Info-Logs**: Optionale Notifications über Parameter `sendNotification`
- Stack Traces werden automatisch in Notifications eingebunden (max. 1000 Zeichen)

### 5. **Console-Logging**
- Alle Logs werden zusätzlich in die Console geschrieben
- Format: `[LEVEL] [SOURCE] TITLE: MESSAGE`

## Verwendung

### Logger erstellen

```go
import (
    "discord-bot-template/internal/shared/utils/logging"
)

// Logger mit allen Features (DB, Notifications)
logger := logging.NewLogger(db, session, guildID, "bot.commands")

// Logger ohne Guild-spezifische Features
logger := logging.NewLogger(db, nil, "", "bot.startup")
```

### Error-Logging (mit automatischer Notification)

```go
// Error mit Stack Trace
logger.LogError(
    "Database Connection Failed",
    "Could not connect to PostgreSQL",
    string(debug.Stack()),
)

// Error ohne Stack Trace
logger.LogError(
    "Command Failed",
    "User does not have permission",
    "",
)
```

### Info-Logging (mit optionaler Notification)

```go
// Info ohne Notification
logger.LogInfo(
    "User Joined",
    "User John#1234 joined the server",
    false,
)

// Info mit Notification
logger.LogInfo(
    "System Update",
    "Bot version 2.0.0 deployed successfully",
    true,
)
```

### Weitere Log-Levels

```go
// Warning
logger.LogWarn(
    "Rate Limit Approaching",
    "80% of API rate limit reached",
)

// Debug
logger.LogDebug(
    "Cache Hit",
    "Retrieved invite from cache",
)
```

## Datenbank-Tabelle

Die `logs` Tabelle wird automatisch beim Start erstellt:

```sql
CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    guild_id VARCHAR(255),
    level VARCHAR(20) NOT NULL,
    title VARCHAR(500) NOT NULL,
    message TEXT NOT NULL,
    stack_trace TEXT,
    source VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Log-Abfrage

```go
import "discord-bot-template/internal/database/tables"

// Alle Error-Logs der letzten 100 Einträge
logs, err := tables.GetLogs(db, "", tables.LogLevelError, 100)

// Guild-spezifische Info-Logs
logs, err := tables.GetLogs(db, guildID, tables.LogLevelInfo, 50)

// Alle Logs
logs, err := tables.GetLogs(db, "", "", 100)
```

## Log-Cleanup

```go
// Lösche alle Logs älter als 30 Tage
deleted, err := tables.DeleteOldLogs(db, 30*24*time.Hour)
log.Printf("Deleted %d old logs", deleted)
```

## Migration von alten Logging-Funktionen

### Vorher (error_log.go, info_log.go)

```go
package utils

// Alte Dateien nur mit Console-Logging
```

### Nachher (logger.go)

```go
import "discord-bot-template/internal/shared/utils/logging"

logger := logging.NewLogger(db, session, guildID, "source")
logger.LogError("Title", "Message", "")
logger.LogInfo("Title", "Message", false)
```

## Vorteile

1. **Persistenz**: Logs überleben Container-Neustarts
2. **Ausfallsicherheit**: Automatischer Fallback auf Dateien
3. **Benachrichtigungen**: Admin-Teams werden bei Errors informiert
4. **Durchsuchbar**: Logs können in der DB gefiltert werden
5. **Strukturiert**: Konsistentes Format über den gesamten Bot

## Docker-Integration

Das `logs/` Verzeichnis sollte als Volume gemounted werden:

```yaml
volumes:
  - ./logs:/app/logs
```

So bleiben Log-Dateien auch bei Container-Updates erhalten.