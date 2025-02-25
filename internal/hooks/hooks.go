package hooks

import "github.com/volte6/gomud/internal/events"

// Register hooks here...
func RegisterListeners() {

	// RoomChange Listeners
	events.RegisterListener(events.RoomChange{}, LocationGMCPUpdates_Listener)

	// NewRound Listeners
	events.RegisterListener(events.NewRound{}, PruneVMs_Listener)
	events.RegisterListener(events.NewRound{}, InactivePlayers_Listener)
	events.RegisterListener(events.NewRound{}, UpdateZoneMutators_Listener)
	events.RegisterListener(events.NewRound{}, SunriseSunset_Listener)
	events.RegisterListener(events.NewRound{}, AuctionUpdate_Listener)
	events.RegisterListener(events.NewRound{}, SpawnLootGoblin_Listener)
	events.RegisterListener(events.NewRound{}, UserRoundTick_Listener)
	events.RegisterListener(events.NewRound{}, MobRoundTick_Listener)
	events.RegisterListener(events.NewRound{}, HandleRespawns_Listener)
	//
	// Combat goes here
	//
	events.RegisterListener(events.NewRound{}, DoCombat_Listener)
	//
	// Done with combat
	//
	events.RegisterListener(events.NewRound{}, AutoHeal_Listener)
	events.RegisterListener(events.NewRound{}, IdleMobs_Listener)

	// Turn Hooks
	events.RegisterListener(events.NewTurn{}, CleanupZombies_Listener)
	events.RegisterListener(events.NewTurn{}, AutoSave_Listener)
	events.RegisterListener(events.NewTurn{}, PruneBuffs_Listener)
	events.RegisterListener(events.NewTurn{}, ActionPoints_Listener)
	events.RegisterListener(events.NewTurn{}, LevelUp_Listener)

}
