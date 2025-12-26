# Voice Logging Service

Dieses Service-Package ist verantwortlich für das Tracking und Logging von Voice-Channel-Aktivitäten in Discord.

## Funktionsweise

### Getrackte Daten

Für jeden User wird folgendes geloggt:
- **Zeiträume**: Genauer Join- und Leave-Zeitstempel für jeden Voice-Channel-Besuch
- **Muted-Zeit**: Gesamtdauer, in der der User gemuted war (Mute oder Self-Mute)
- **Deafened-Zeit**: Gesamtdauer, in der der User deafened war (Deaf oder Self-Deaf)
- **Streaming-Zeit**: Gesamtdauer, in der der User Screenshare aktiv hatte
- **Channel-ID**: In welchem Voice-Channel sich der User befand
- **Totale Zeit**: Automatisch berechnete Gesamtzeit im Channel

### State Management

Das Service nutzt eine In-Memory-Map (`activeSessions`), um den aktuellen Status jedes Users zu tracken:
- Aktueller Mute/Deafen/Streaming-Status
- Zeitpunkt der letzten Statusänderung
- Akkumulierte Durations für jede Aktivität

### Event-Handling

Das Service verarbeitet vier Szenarien:

1. **Voice Join**: User betritt einen Voice-Channel
   - Neuer Eintrag in `user_voice_logs` wird erstellt
   - Session-State wird initialisiert

2. **Voice Leave**: User verlässt einen Voice-Channel
   - Finale Durations werden berechnet
   - Datenbank-Eintrag wird mit `left_at` und `total_duration` aktualisiert
   - Session wird beendet

3. **Channel Switch**: User wechselt zwischen Voice-Channels
   - Alte Session wird beendet (wie Voice Leave)
   - Neue Session wird gestartet (wie Voice Join)

4. **State Change**: User ändert Mute/Deafen/Streaming im selben Channel
   - Durations werden inkrementell aktualisiert
   - Datenbank wird mit aktuellen Werten synchronisiert

## Datenbank-Schema

```sql
CREATE TABLE user_voice_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL,
    channel_id VARCHAR(32) NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    muted_duration BIGINT NOT NULL DEFAULT 0,
    deafen_duration BIGINT NOT NULL DEFAULT 0,
    stream_duration BIGINT NOT NULL DEFAULT 0,
    total_duration BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## Verwendung

```go
import "discord-bot-template/bot/services/events/voice"

// In Voice State Update Event
func OnVoiceStateUpdate(session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate, db *sql.DB) {
    voice.LogVoiceStateUpdate(db, voiceState)
}
```

## Abfragen

### Aktive Voice-Session eines Users
```go
activeLog, err := tables.GetActiveVoiceLogByUser(db, userID)
```

### Alle Voice-Logs eines Users
```go
logs, err := tables.GetUserVoiceLogsByUser(db, userID)
```

### Voice-Statistiken eines Users
```go
totalTime, mutedTime, deafenTime, streamTime, err := tables.GetUserVoiceStatistics(db, userID)
```

### Voice-Logs für einen Channel
```go
logs, err := tables.GetUserVoiceLogsByChannel(db, channelID)
```

## Concurrency Safety

Das Service verwendet einen `sync.RWMutex` für thread-safe Zugriffe auf die `activeSessions`-Map, da Discord-Events asynchron verarbeitet werden können.
