package hooks

import (
	"github.com/volte6/gomud/internal/events"
)

// Register hooks here...
func RegisterListeners() {

	// RoomChange Listeners
	events.RegisterListener(events.RoomChange{}, LocationGMCPUpdates)
	events.RegisterListener(events.RoomChange{}, LocationMusicChange)

	// NewRound Listeners
	events.RegisterListener(events.NewRound{}, PruneVMs)
	events.RegisterListener(events.NewRound{}, InactivePlayers)
	events.RegisterListener(events.NewRound{}, UpdateZoneMutators)
	events.RegisterListener(events.NewRound{}, SunriseSunset)
	events.RegisterListener(events.NewRound{}, AuctionUpdate)
	events.RegisterListener(events.NewRound{}, SpawnLootGoblin)
	events.RegisterListener(events.NewRound{}, UserRoundTick)
	events.RegisterListener(events.NewRound{}, MobRoundTick)
	events.RegisterListener(events.NewRound{}, HandleRespawns)
	//
	// Combat goes here
	//
	events.RegisterListener(events.NewRound{}, DoCombat)
	//
	// Done with combat
	//
	events.RegisterListener(events.NewRound{}, AutoHeal)
	events.RegisterListener(events.NewRound{}, IdleMobs)

	// Turn Hooks
	events.RegisterListener(events.NewTurn{}, CleanupZombies)
	events.RegisterListener(events.NewTurn{}, AutoSave)
	events.RegisterListener(events.NewTurn{}, PruneBuffs)
	events.RegisterListener(events.NewTurn{}, ActionPoints)
	events.RegisterListener(events.NewTurn{}, LevelUp)

	// ItemOwnership
	events.RegisterListener(events.ItemOwnership{}, CheckItemQuests)

	// MSP Sound
	events.RegisterListener(events.MSP{}, PlaySound)
	// Quest Events
	events.RegisterListener(events.Quest{}, HandleQuestUpdate)
	// Spawn events
	events.RegisterListener(events.PlayerSpawn{}, HandleJoin)
	events.RegisterListener(events.PlayerDespawn{}, HandleLeave)

	// Listener for debugging some stuff (catches all events)
	/*
		events.RegisterListener(nil, func(e events.Event) bool {
			t := e.Type()
			if t != `NewTurn` && t != `Message` && t != `NewRound` && t != `Broadcast` {
				slog.Info("Event", "e.Type", e.Type(), "e", e)
			}
			return true
		})
	*/

	// Log tee to users
	events.RegisterListener(events.Log{}, FollowLogs)
}
