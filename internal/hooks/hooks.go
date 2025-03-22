package hooks

import (
	"github.com/volte6/gomud/internal/events"
)

// Register hooks here...
func RegisterListeners() {

	// Buffs
	events.RegisterListener(events.Buff{}, ApplyBuffs)

	// RoomChange Listeners
	events.RegisterListener(events.RoomChange{}, LocationGMCPUpdates)
	events.RegisterListener(events.RoomChange{}, LocationMusicChange)

	// NewRound Listeners
	events.RegisterListener(events.NewRound{}, PruneVMs)
	events.RegisterListener(events.NewRound{}, InactivePlayers)
	events.RegisterListener(events.NewRound{}, UpdateZoneMutators)
	events.RegisterListener(events.NewRound{}, CheckNewDay)
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

	// ItemOwnership
	events.RegisterListener(events.ItemOwnership{}, CheckItemQuests)

	// MSP Sound
	events.RegisterListener(events.MSP{}, PlaySound)
	// Quest Events
	events.RegisterListener(events.Quest{}, HandleQuestUpdate)
	// Spawn events
	events.RegisterListener(events.PlayerSpawn{}, HandleJoin)
	events.RegisterListener(events.PlayerDespawn{}, HandleLeave, events.Last) // This is a final listener, has to happen last

	// Levelup Notifications
	events.RegisterListener(events.LevelUp{}, SendLevelNotifications)
	events.RegisterListener(events.LevelUp{}, CheckGuide)

	// Day/Night cycle
	events.RegisterListener(events.DayNightCycle{}, NotifySunriseSunset)

	// Looking
	events.RegisterListener(events.Looking{}, HandleLookHints)

	// GMCPOut
	events.RegisterListener(events.GMCPOut{}, GMCPOut_SendGMCP)
	// Messages
	events.RegisterListener(events.Message{}, Message_SendMessage)
	// Prompt
	events.RegisterListener(events.RedrawPrompt{}, RedrawPrompt_SendRedraw)

	events.RegisterListener(events.WebClientCommand{}, WebClientCommand_SendWebClientCommand)

	events.RegisterListener(events.Broadcast{}, Broadcast_SendToAll)

	// Log tee to users
	events.RegisterListener(events.Log{}, FollowLogs)

	// Listener for debugging some stuff (catches all events)
	/*
		events.RegisterListener(nil, func(e events.Event) events.ListenerReturn {
			t := e.Type()
			if t != `NewTurn` && t != `Message` && t != `NewRound` && t != `Broadcast` {
				mudlog.Info("Event", "e.Type", e.Type(), "e", e)
			}
			return events.Continue
		})
	*/

}
