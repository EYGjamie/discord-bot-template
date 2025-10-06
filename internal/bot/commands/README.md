# Discord Bot Commands

This directory contains all slash command implementations for the OVGU Discord bot.

## Structure

Each command should be implemented as a separate file with the following structure:

```go
package commands

import "github.com/bwmarrin/discordgo"

var ExampleCommand = &discordgo.ApplicationCommand{
    Name:        "example",
    Description: "An example command",
}

func HandleExample(bot_session *discordgo.Session, bot_interaction *discordgo.InteractionCreate) {
    // Command implementation
}
```

## Registering Commands

Commands are automatically registered in `internal/bot/bot.go`
