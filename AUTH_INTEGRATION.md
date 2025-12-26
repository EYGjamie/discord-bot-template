# Authentifizierung - User-ID Weitergabe

## Wie funktioniert es?

Die Moderator-ID wird jetzt automatisch aus dem Request ermittelt. Das Backend unterstützt mehrere Methoden:

### 1. Cookie (Empfohlen für Production)
```javascript
// Frontend sendet Cookie mit
document.cookie = "user_id=YOUR_DISCORD_ID";
```

### 2. Authorization Header
```javascript
// Frontend sendet Bearer Token mit Discord ID
fetch('/api/moderation/warns', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_DISCORD_ID',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    user_id: 'TARGET_USER_ID',
    reason: 'Breaking rules'
  })
});
```

### 3. X-User-ID Header
```javascript
fetch('/api/moderation/warns', {
  method: 'POST',
  headers: {
    'X-User-ID': 'YOUR_DISCORD_ID',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    user_id: 'TARGET_USER_ID',
    reason: 'Breaking rules'
  })
});
```

### 4. Query Parameter (Nur für Testing)
```
POST /api/moderation/warns?user_id=YOUR_DISCORD_ID
```

## Frontend Integration

Im Frontend musst du die User-ID des eingeloggten Users in einem der folgenden Orte speichern:

### Option 1: localStorage nach Discord Login
```typescript
// Nach erfolgreichem Discord Login
localStorage.setItem('discord_user_id', user.id);

// In api.ts
const getUserId = () => localStorage.getItem('discord_user_id') || '';

export const api = {
  moderation: {
    createWarn: (data: { user_id: string; reason: string }) => 
      fetch(`${API_BASE_URL}/api/moderation/warns`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId()
        },
        body: JSON.stringify(data),
      }).then(res => res.json()),
  }
}
```

### Option 2: Context/State Management
```typescript
// In einem Auth Context
const AuthContext = createContext<{ userId: string | null }>({ userId: null });

// In api calls
const { userId } = useContext(AuthContext);
```

## Testing

Für Testing kannst du temporär eine feste Discord ID verwenden:

```typescript
// In api.ts - NUR FÜR TESTING
const TESTING_USER_ID = '1234567890'; // Deine Discord ID aus der DB

headers: { 
  'Content-Type': 'application/json',
  'X-User-ID': TESTING_USER_ID  // <-- Temporär für Testing
}
```

## Middleware

Die Middleware unterstützt aktuell folgende Authentifizierungsmethoden (siehe `backend/middleware/auth.go`):

1. Cookie: `user_id`
2. Authorization Header: `Bearer {discord_id}`
3. Custom Header: `X-User-ID`
4. Query Parameter: `?user_id={discord_id}` (nur Testing)

Du kannst weitere Methoden hinzufügen oder JWT-Token Validierung implementieren.
