package bot

import (
	"discord-bot-template/internal/bot/events"

	"github.com/bwmarrin/discordgo"
)

// registerHandlers registriert alle Event-Handler für den Bot
// Diese Funktion wird beim Start des Bots aufgerufen
func (bot *Bot) registerHandlers() {

	// OnReady Handler
	bot.session.AddHandler(bot.onReady)

	// Event Handler Registrierung
	bot.session.AddHandler(func(bot_session *discordgo.Session, MessageCreate *discordgo.MessageCreate) {
		events.OnMessageCreate(bot_session, MessageCreate, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, GuildMemberAdd *discordgo.GuildMemberAdd) {
		events.OnGuildMemberAdd(bot_session, GuildMemberAdd, bot.db, bot.inviteCache)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, GuildMemberRemove *discordgo.GuildMemberRemove) {
		events.OnGuildMemberRemove(bot_session, GuildMemberRemove, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, GuildMemberUpdate *discordgo.GuildMemberUpdate) {
		events.OnGuildMemberUpdate(bot_session, GuildMemberUpdate, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, MessageUpdate *discordgo.MessageUpdate) {
		events.OnMessageUpdate(bot_session, MessageUpdate, bot.settings)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, MessageDelete *discordgo.MessageDelete) {
		events.OnMessageDelete(bot_session, MessageDelete, bot.settings)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, UserUpdate *discordgo.UserUpdate) {
		events.OnUserUpdate(bot_session, UserUpdate, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, VoiceStateUpdate *discordgo.VoiceStateUpdate) {
		events.OnVoiceStateUpdate(bot_session, VoiceStateUpdate, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, GuildRoleCreate *discordgo.GuildRoleCreate) {
		events.OnGuildRoleCreate(bot_session, GuildRoleCreate, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, GuildRoleUpdate *discordgo.GuildRoleUpdate) {
		events.OnGuildRoleUpdate(bot_session, GuildRoleUpdate, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, GuildRoleDelete *discordgo.GuildRoleDelete) {
		events.OnGuildRoleDelete(bot_session, GuildRoleDelete, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, ChannelCreate *discordgo.ChannelCreate) {
		events.OnChannelCreate(bot_session, ChannelCreate, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, ChannelUpdate *discordgo.ChannelUpdate) {
		events.OnChannelUpdate(bot_session, ChannelUpdate, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, ChannelDelete *discordgo.ChannelDelete) {
		events.OnChannelDelete(bot_session, ChannelDelete, bot.db)
	})
	bot.session.AddHandler(func(bot_session *discordgo.Session, GuildCreate *discordgo.GuildCreate) {
		events.OnGuildCreate(bot_session, GuildCreate, bot.db)
	})

	// Interaction Handler
	bot.session.AddHandler(bot.onInteractionCreate)
}
