# Discord Bot Events

This directory contains event handlers for Discord gateway events.

## Available Events

- Message events (create, update, delete)
- Member events (join, leave, update)
- Guild events (create, update, delete)
- Voice state events
- Reaction events

## Example

```go
package events

import "github.com/bwmarrin/discordgo"

func OnGuildMemberAdd(bot_session *discordgo.Session, m *discordgo.GuildMemberAdd) {
    // Welcome new members
}
```
