package commands

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	cmdutils "discord-bot-template/bot/utils/commands"

	"github.com/bwmarrin/discordgo"
)

// SetupCoinflipCommand registriert den /coinflip Command
func SetupCoinflipCommand(s *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "coinflip",
		Description: "Wirft eine Münze mit einer benutzerdefinierten Anzahl von Seiten",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "sides",
				Description: "Anzahl der Seiten (Standard: 2)",
				Required:    false,
				MinValue:    func() *float64 { v := 2.0; return &v }(),
				MaxValue:    100,
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
	return err
}

// HandleCoinflipCommand behandelt den /coinflip Command
func HandleCoinflipCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	// Extrahiere Parameter
	options := i.ApplicationCommandData().Options
	sides := 2 // Standard: klassischer Coinflip

	if len(options) > 0 && options[0].Name == "sides" {
		sides = int(options[0].IntValue())
	}

	// Validierung
	if sides < 2 {
		cmdutils.RespondError(s, i, "Die Münze muss mindestens 2 Seiten haben.")
		return
	}

	if sides > 100 {
		cmdutils.RespondError(s, i, "Die Münze darf maximal 100 Seiten haben.")
		return
	}

	// Random number generator initialisieren
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := rng.Intn(sides) + 1

	// Spezielle Behandlung für 2-seitige Münze
	var resultText string
	var emoji string

	if sides == 2 {
		if result == 1 {
			resultText = "Kopf"
			emoji = "🪙"
		} else {
			resultText = "Zahl"
			emoji = "🔢"
		}
	} else {
		resultText = fmt.Sprintf("Seite %d", result)
		emoji = "🎲"
	}

	// Erstelle Embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Coinflip (%d-seitig)", emoji, sides),
		Description: fmt.Sprintf("**%s**", resultText),
		Color:       0xFFD700,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Gewürfelt von %s", i.Member.User.Username),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Füge Wahrscheinlichkeit hinzu wenn mehr als 2 Seiten
	if sides > 2 {
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:   "Ergebnis",
				Value:  fmt.Sprintf("%d von %d", result, sides),
				Inline: true,
			},
			{
				Name:   "Wahrscheinlichkeit",
				Value:  fmt.Sprintf("1/%d (%.2f%%)", sides, 100.0/float64(sides)),
				Inline: true,
			},
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
