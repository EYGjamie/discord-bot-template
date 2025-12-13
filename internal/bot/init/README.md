# Bot Initialisierung

Dieses Package enthält Funktionen zur Initialisierung und Synchronisierung von Discord-Server-Daten in die Datenbank.

## Übersicht

Wenn der Bot einem Discord-Server beitritt oder beim Start des Bots, werden automatisch alle relevanten Daten synchronisiert:

- **Rollen** - Alle Server-Rollen
- **Channels** - Alle Text-, Voice- und andere Channels
- **Members** - Alle Server-Mitglieder
- **Rollen-Zuweisungen** (optional) - Member-zu-Rolle Beziehungen

## Hauptfunktionen

### `SyncGuildOnJoin(session, guildID, db)`

Führt eine vollständige Basis-Synchronisierung durch:
1. Synchronisiert alle Rollen
2. Synchronisiert alle Channels
3. Synchronisiert alle Members

```go
err := guildInit.SyncGuildOnJoin(session, guildID, db)
```

### `SyncAllRoles(session, guildID, db)`

Synchronisiert alle Rollen eines Servers einzeln:
- Ruft alle Rollen vom Discord-Server ab
- Speichert/aktualisiert jede Rolle in der Datenbank
- Gibt Statistiken über erfolgreiche/fehlgeschlagene Syncs aus

### `SyncAllChannels(session, guildID, db)`

Synchronisiert alle Channels eines Servers:
- Ruft alle Channels vom Discord-Server ab
- Speichert/aktualisiert jeden Channel in der Datenbank
- Unterstützt alle Channel-Typen (Text, Voice, Category, etc.)

### `SyncAllMembers(session, guildID, db)`

Synchronisiert alle Members eines Servers:
- Nutzt Pagination für große Server (bis zu unbegrenzt viele Members)
- Holt Members in Batches von 1000
- Speichert/aktualisiert jeden Member in der Datenbank

### `SyncGuildRoleAssignments(session, guildID, db)`

Synchronisiert alle Rollen-Zuweisungen (Member ↔ Rolle):
- Sollte nach `SyncAllMembers` und `SyncAllRoles` aufgerufen werden
- Erstellt die Many-to-Many Beziehungen in der Datenbank

### `FullGuildSync(session, guildID, db)`

Führt eine komplette Synchronisierung inkl. Rollen-Zuweisungen durch:
- Ruft `SyncGuildOnJoin` auf
- Ruft zusätzlich `SyncGuildRoleAssignments` auf
- Ideal für initiale Setups oder manuelle Re-Syncs

## Event-Integration

### GuildCreate Event

Das `OnGuildCreate` Event wird automatisch ausgelöst:
1. Wenn der Bot einem neuen Server beitritt
2. Beim Start des Bots für jeden Server, auf dem er sich befindet

```go
func OnGuildCreate(bot_session *discordgo.Session, guild *discordgo.GuildCreate, db *sql.DB) {
    err := guildInit.SyncGuildOnJoin(bot_session, guild.ID, db)
    // ...
}
```

## Performance-Hinweise

### Große Server

Für große Server (>1000 Members):
- Die Synchronisierung kann einige Sekunden dauern
- Members werden in Batches von 1000 abgerufen
- Discord API Rate Limits werden beachtet

### Bot-Start

Beim Bot-Start wird `GuildCreate` für **jeden** Server ausgelöst:
- Bei vielen Servern kann dies zu hoher Last führen
- Erwäge das Hinzufügen von Delays zwischen Server-Syncs
- Oder implementiere eine Queue für Server-Synchronisierungen

## Verwendung

### Automatische Synchronisierung

Standardmäßig wird bei jedem `GuildCreate` Event automatisch synchronisiert. Keine weitere Konfiguration nötig.

### Manuelle Synchronisierung

Für manuelle Re-Syncs (z.B. via Command):

```go
import guildInit "discord-bot-template/internal/bot/init"

// Basis-Sync
err := guildInit.SyncGuildOnJoin(session, guildID, db)

// Oder vollständiger Sync mit Rollen-Zuweisungen
err := guildInit.FullGuildSync(session, guildID, db)
```

### Einzelne Komponenten synchronisieren

```go
// Nur Rollen
err := guildInit.SyncAllRoles(session, guildID, db)

// Nur Channels
err := guildInit.SyncAllChannels(session, guildID, db)

// Nur Members
err := guildInit.SyncAllMembers(session, guildID, db)
```

## Logging

Alle Sync-Funktionen loggen detailliert:
- Start und Ende der Synchronisierung
- Anzahl synchronisierter Elemente
- Fehler bei einzelnen Elementen
- Gesamtstatistiken

Beispiel-Output:
```
Starte vollständige Synchronisierung für Guild 123456789
Synchronisiere Rollen für Guild 123456789
Rollen-Synchronisierung abgeschlossen: 15 erfolgreich, 0 Fehler
Synchronisiere Channels für Guild 123456789
Channel-Synchronisierung abgeschlossen: 42 erfolgreich, 0 Fehler
Synchronisiere Members für Guild 123456789
Member-Synchronisierung abgeschlossen: 1337 erfolgreich, 0 Fehler
Vollständige Synchronisierung für Guild 123456789 abgeschlossen
```

## Error Handling

- Fehler bei der Synchronisierung werden geloggt
- Einzelne fehlgeschlagene Items brechen nicht die gesamte Sync ab
- API-Fehler werden zurückgegeben
- Bei kritischen Fehlern wird die Sync abgebrochen

## Zukünftige Erweiterungen

Mögliche Verbesserungen:
- Parallel-Synchronisierung für schnellere Performance
- Diff-basierte Updates (nur Änderungen synchronisieren)
- Incremental Sync statt Full Sync
- Sync-Queue für viele Server
- Retry-Mechanismus für fehlgeschlagene Items
