package bot

import (
	"fmt"
	"time"

	"discord-bot-template/bot/commands"
	"discord-bot-template/bot/handlers"
	"discord-bot-template/bot/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// onInteractionCreate wird aufgerufen wenn ein Slash Command ausgeführt wird
func (bot *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Nur Application Commands verarbeiten
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		logger := logging.NewLogger(bot.db, s, i.GuildID, "bot.interaction")
		cmdName := i.ApplicationCommandData().Name
		userTag := "Unknown"
		if i.Member != nil && i.Member.User != nil {
			userTag = i.Member.User.Username
		}
		logger.LogInfo("Command Executed", fmt.Sprintf("/%s by %s", cmdName, userTag), false)

		// Prüfe welcher Command ausgeführt wurde
		switch cmdName {
		// Commands now managed through Web UI
		/* case "moderation":
			commands.HandleModerationCommand(s, i, bot.settings)
		case "setmodrole":
			commands.HandleSetModRoleCommand(s, i, bot.db)
		case "setupcreatevoice":
			commands.HandleSetupCreateVoiceCommand(s, i, bot.db)
		case "purge-schedule":
			commands.HandlePurgeScheduleCommand(s, i, bot.db) */
		case "warn":
			commands.HandleWarnCommand(s, i, bot.db)
		case "note":
			commands.HandleNoteCommand(s, i, bot.db)
		case "userinfo":
			commands.HandleUserInfoCommand(s, i, bot.db)
		case "coinflip":
			commands.HandleCoinflipCommand(s, i, bot.db)
		case "purge":
			commands.HandlePurgeCommand(s, i, bot.db)
		}
	case discordgo.InteractionMessageComponent:
		// Handle Button/Select Menu Interactions
		customID := i.MessageComponentData().CustomID

		// Handle Event RSVP Buttons
		if len(customID) >= 12 {
			var eventID int64
			var action string

			if _, err := fmt.Sscanf(customID, "event_accept_%d", &eventID); err == nil {
				action = "accepted"
			} else if _, err := fmt.Sscanf(customID, "event_decline_%d", &eventID); err == nil {
				action = "declined"
			}

			if action != "" {
				userID := i.User.ID
				if i.Member != nil {
					userID = i.Member.User.ID
				}

				// Update RSVP status in database
				_, err := bot.db.Exec(`
					UPDATE event_guests 
					SET rsvp_status = $1, rsvp_at = $2 
					WHERE event_id = $3 AND user_id = $4
				`, action, time.Now(), eventID, userID)

				if err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "❌ Fehler beim Aktualisieren deines RSVP-Status.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}

				// Get event details for confirmation
				var eventTitle string
				bot.db.QueryRow("SELECT title FROM events WHERE id = $1", eventID).Scan(&eventTitle)
				if eventTitle == "" {
					eventTitle = "Event"
				}

				responseText := ""
				if action == "accepted" {
					responseText = fmt.Sprintf("✅ Du hast dem Event **%s** zugesagt!", eventTitle)
				} else {
					responseText = fmt.Sprintf("❌ Du hast dem Event **%s** abgesagt.", eventTitle)
				}

				// Update the original message to show the response
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Content:    responseText,
						Components: []discordgo.MessageComponent{}, // Remove buttons
					},
				})
				return
			}
		}

		// Handle Create Voice Buttons und Select Menus
		if len(customID) >= 3 && customID[:3] == "cv_" {
			// Create Voice Buttons und Select Menus
			if customID == "cv_kick_select" || customID == "cv_block_select" || customID == "cv_unblock_select" {
				handlers.HandleCreateVoiceSelects(s, i, bot.db)
			} else {
				handlers.HandleCreateVoiceButtons(s, i, bot.db)
			}
		}
	case discordgo.InteractionModalSubmit:
		// Handle Modal Submissions
		customID := i.ModalSubmitData().CustomID
		if customID == "cv_rename_modal" || customID == "cv_limit_modal" {
			handlers.HandleCreateVoiceModals(s, i, bot.db)
		}
	}
}
