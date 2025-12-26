package bot

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// InviteCache speichert Invite-Informationen für jede Guild
type InviteCache struct {
	mu      sync.RWMutex
	invites map[string]map[string]*discordgo.Invite // guildID -> inviteCode -> Invite
}

// NewInviteCache erstellt einen neuen Invite Cache
func NewInviteCache() *InviteCache {
	return &InviteCache{
		invites: make(map[string]map[string]*discordgo.Invite),
	}
}

// Update aktualisiert den Cache für eine Guild
func (ic *InviteCache) Update(s *discordgo.Session, guildID string) error {
	invites, err := s.GuildInvites(guildID)
	if err != nil {
		return err
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	if ic.invites[guildID] == nil {
		ic.invites[guildID] = make(map[string]*discordgo.Invite)
	}

	// Speichere alle Invites
	for _, invite := range invites {
		ic.invites[guildID][invite.Code] = invite
	}

	return nil
}

// GetUsedInvite vergleicht alte und neue Invites um herauszufinden welcher verwendet wurde
func (ic *InviteCache) GetUsedInvite(s *discordgo.Session, guildID string) (*discordgo.Invite, error) {
	// Hole aktuelle Invites
	currentInvites, err := s.GuildInvites(guildID)
	if err != nil {
		return nil, err
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	oldInvites := ic.invites[guildID]
	if oldInvites == nil {
		// Kein Cache vorhanden, initialisiere mit aktuellen Invites
		ic.invites[guildID] = make(map[string]*discordgo.Invite)
		for _, invite := range currentInvites {
			ic.invites[guildID][invite.Code] = invite
		}
		return nil, nil
	}

	// Vergleiche Uses-Count
	var usedInvite *discordgo.Invite
	for _, currentInvite := range currentInvites {
		oldInvite, exists := oldInvites[currentInvite.Code]
		if !exists {
			// Neuer Invite, ignorieren
			ic.invites[guildID][currentInvite.Code] = currentInvite
			continue
		}

		// Wenn Uses gestiegen ist, wurde dieser Invite verwendet
		if currentInvite.Uses > oldInvite.Uses {
			usedInvite = currentInvite
		}

		// Aktualisiere Cache
		ic.invites[guildID][currentInvite.Code] = currentInvite
	}

	return usedInvite, nil
}

// Remove entfernt eine Guild aus dem Cache
func (ic *InviteCache) Remove(guildID string) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	delete(ic.invites, guildID)
}

// InitializeForAllGuilds initialisiert den Cache für alle Guilds
func (ic *InviteCache) InitializeForAllGuilds(s *discordgo.Session) {
	guilds := s.State.Guilds
	for _, guild := range guilds {
		if err := ic.Update(s, guild.ID); err != nil {
			log.Printf("Fehler beim Initialisieren des Invite-Cache für Guild %s: %v", guild.ID, err)
		} else {
			log.Printf("Invite-Cache initialisiert für Guild: %s", guild.Name)
		}
	}
}
