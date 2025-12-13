package events

import (
	"database/sql"
	"log"

	"discord-bot-template/internal/database/tables"
	"discord-bot-template/internal/shared/services/events/user"

	"github.com/bwmarrin/discordgo"
)

// InviteCache Interface um Import-Cycle zu vermeiden
type InviteCache interface {
	GetUsedInvite(s *discordgo.Session, guildID string) (*discordgo.Invite, error)
}

func OnGuildMemberAdd(bot_session *discordgo.Session, memberAdd *discordgo.GuildMemberAdd, db *sql.DB, inviteCache InviteCache) {
	// Synchronisiere User in Datenbank
	user.UpsertUser(db, memberAdd.User, memberAdd.Member)

	// Ermittele welcher Invite verwendet wurde
	usedInvite, err := inviteCache.GetUsedInvite(bot_session, memberAdd.GuildID)
	if err != nil {
		log.Printf("Fehler beim Ermitteln des verwendeten Invites für User %s: %v", memberAdd.User.ID, err)
	}

	// Extrahiere Invite-Informationen
	var inviterID *string
	var inviteCode *string

	if usedInvite != nil {
		inviteCode = &usedInvite.Code
		if usedInvite.Inviter != nil {
			inviterID = &usedInvite.Inviter.ID
			log.Printf("User %s wurde durch Invite '%s' von %s eingeladen",
				memberAdd.User.Username,
				usedInvite.Code,
				usedInvite.Inviter.Username)
		} else {
			log.Printf("User %s wurde durch Invite '%s' eingeladen (Inviter unbekannt)",
				memberAdd.User.Username,
				usedInvite.Code)
		}
	} else {
		log.Printf("User %s ist beigetreten (verwendeter Invite konnte nicht ermittelt werden)",
			memberAdd.User.Username)
	}

	// Logge den Join in die Datenbank
	_, err = tables.LogUserJoin(db, memberAdd.User.ID, memberAdd.GuildID, inviterID, inviteCode)
	if err != nil {
		log.Printf("Fehler beim Loggen des Joins für User %s: %v", memberAdd.User.ID, err)
	}
}
