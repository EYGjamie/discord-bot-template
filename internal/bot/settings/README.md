# Bot Settings System

Dynamisches Einstellungs-System für den Discord Bot mit Datenbank-Backend und Caching.

## Features

- ✅ Persistente Speicherung in PostgreSQL
- ✅ Automatisches Caching mit 30s Refresh-Intervall
- ✅ Thread-Safe mit RWMutex
- ✅ Einstellungen können im laufenden Betrieb geändert werden
- ✅ Aktivieren/Deaktivieren ohne Werte zu löschen
- ✅ Typisierte Getter (String, Bool, Int)

## Verwendung

### Settings Manager initialisieren

Der Settings Manager wird automatisch im Bot initialisiert:

```go
settingsManager := settings.NewManager(db)
```

### Einstellungen abrufen

```go
// String-Wert abrufen
channelID := settingsManager.GetString("moderation_channel_id", "default")

// Boolean-Wert abrufen
enabled := settingsManager.GetBool("log_message_edits", false)

// Integer-Wert abrufen
maxWarnings := settingsManager.GetInt("max_warnings", 3)

// Prüfen ob eine Einstellung aktiviert ist
if settingsManager.IsEnabled("feature_xyz") {
    // Feature ist aktiviert
}
```

### Einstellungen setzen

```go
// String speichern
err := settingsManager.SetString("moderation_channel_id", "123456789", true)

// Boolean speichern
err := settingsManager.SetBool("log_message_edits", true, true)

// Einstellung aktivieren/deaktivieren (Wert bleibt erhalten)
err := settingsManager.SetEnabled("log_message_edits", false)

// Einstellung löschen
err := settingsManager.Delete("old_setting")
```

## Beispiel: Moderations-Logging

### 1. Moderations-Kanal setzen

Mit dem `/moderation` Command:

```
/moderation set-channel channel:#mod-logs
```

Dies speichert die Channel-ID in der Datenbank.

### 2. Message-Edit-Logging aktivieren

```
/moderation toggle-edits enabled:true
```

### 3. Message-Delete-Logging aktivieren

```
/moderation toggle-deletes enabled:true
```

### 4. Status anzeigen

```
/moderation status
```

Zeigt alle aktuellen Einstellungen an.

## Verfügbare Einstellungen

| Key | Typ | Beschreibung |
|-----|-----|--------------|
| `moderation_channel_id` | string | Channel-ID für Moderations-Logs |
| `log_message_edits` | bool | Logging von bearbeiteten Nachrichten |
| `log_message_deletes` | bool | Logging von gelöschten Nachrichten |

## Neue Einstellungen hinzufügen

### In Code verwenden

```go
// Neue Feature-Flag prüfen
if settingsManager.IsEnabled("new_feature") {
    // Feature-Code hier
}

// Neuen Config-Wert abrufen
timeout := settingsManager.GetInt("command_timeout_seconds", 30)
```

### Per Command setzen

Erstelle einen neuen Slash-Command oder erweitere den `/moderation` Command:

```go
err := settingsManager.SetBool("new_feature", true, true)
```

## Best Practices

1. **Default-Werte**: Immer Default-Werte in Getter-Aufrufen angeben
2. **Enabled-Flag**: Nutze `IsEnabled()` für Feature-Flags
3. **Cache**: Der Cache wird automatisch alle 30s aktualisiert
4. **Thread-Safety**: Alle Methoden sind thread-safe
5. **Performance**: Getter lesen aus dem Cache, keine DB-Abfragen

## Datenbank-Schema

```sql
CREATE TABLE bot_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'string',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Troubleshooting

### Einstellung wird nicht aktualisiert

- Warte bis zu 30 Sekunden für automatischen Cache-Refresh
- Oder rufe `settingsManager.RefreshCache()` manuell auf

### Einstellung existiert nicht

```go
// Prüfe ob Einstellung existiert
allSettings := settingsManager.GetAll()
if setting, exists := allSettings["my_key"]; exists {
    // Einstellung existiert
}
```

### Performance-Probleme

- Getter lesen aus dem Cache (sehr schnell)
- Setter schreiben in die DB (etwas langsamer)
- Cache-Refresh läuft im Hintergrund
