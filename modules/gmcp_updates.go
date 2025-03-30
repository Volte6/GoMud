package modules

import (
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/plugins"
	"github.com/volte6/gomud/internal/users"
)

// ////////////////////////////////////////////////////////////////////
// NOTE: The init function in Go is a special function that is
// automatically executed before the main function within a package.
// It is used to initialize variables, set up configurations, or
// perform any other setup tasks that need to be done before the
// program starts running.
// ////////////////////////////////////////////////////////////////////
func init() {

	//
	// We can use all functions only, but this demonstrates
	// how to use a struct
	//
	g := GMCPModule{
		plug: plugins.New(`GMCP`, `1.0`),
	}

	events.RegisterListener(events.EquipmentChange{}, g.equipmentChangeHandler)
	events.RegisterListener(events.PlayerSpawn{}, g.playSpawnHandler)

}

type GMCPModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug *plugins.Plugin
}

func (g *GMCPModule) equipmentChangeHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.EquipmentChange)
	if !typeOk {
		return events.Continue // Return false to stop halt the event chain for this event
	}

	if evt.UserId == 0 {
		return events.Continue
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil || !connections.GetClientSettings(user.ConnectionId()).GmcpEnabled(`Char`) {
		return events.Continue
	}

	// Changing equipment might affect stats, inventory, maxhp/maxmp etc
	events.AddToQueue(events.GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: `char.inventory, char.stats, char.vitals`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth *Can comma seaparate for multiple*
	})

	return events.Continue
}

func (g *GMCPModule) playSpawnHandler(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PlayerSpawn)
	if !typeOk {
		return events.Continue // Return false to stop halt the event chain for this event
	}

	if evt.UserId == 0 {
		return events.Continue
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil || !connections.GetClientSettings(user.ConnectionId()).GmcpEnabled(`Char`) {
		return events.Continue
	}

	// Send full update
	events.AddToQueue(events.GMCPUpdate{
		UserId:     evt.UserId,
		Identifier: `char`, // char, char.info, char.inventory, char.stats, char.vitals, char.worth *Can comma seaparate for multiple*
	})

	return events.Continue
}
