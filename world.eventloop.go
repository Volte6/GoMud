package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

// Should only handle sending messages out to users
func (w *World) EventLoop() {

	var c configs.Config = configs.GetConfig()

	turnNow := util.GetTurnCount()

	// Player joined the world
	//
	eq := events.GetQueue(events.Log{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Looking
	//
	eq = events.GetQueue(events.Looking{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Auctions
	//
	eq = events.GetQueue(events.Auction{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Day/Night handling
	//
	eq = events.GetQueue(events.DayNightCycle{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Player joined the world
	//
	eq = events.GetQueue(events.PlayerSpawn{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Player left the world
	//
	eq = events.GetQueue(events.PlayerDespawn{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Player Levelup Notifications
	//
	eq = events.GetQueue(events.LevelUp{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Death Notifications
	//
	eq = events.GetQueue(events.PlayerDeath{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// ScriptedEvents
	//
	eq = events.GetQueue(events.ScriptedEvent{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// ItemOwnership
	//
	eq = events.GetQueue(events.ItemOwnership{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// System commands such as /reload
	//
	eq = events.GetQueue(events.System{})
	for eq.Len() > 0 {

		e := eq.Poll()

		sys, typeOk := e.(events.System)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "System", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		if sys.Command == `reload` {

			events.AddToQueue(events.Broadcast{
				Text: `Reloading flat files...`,
			})

			loadAllDataFiles(true)

			events.AddToQueue(events.Broadcast{
				Text:            `Done.` + term.CRLFStr,
				SkipLineRefresh: true,
			})

		} else if sys.Command == `kick` {
			w.Kick(sys.Data.(int))
		} else if sys.Command == `leaveworld` {

			if userInfo := users.GetByUserId(sys.Data.(int)); userInfo != nil {
				events.AddToQueue(events.PlayerDespawn{
					UserId:        userInfo.UserId,
					RoomId:        userInfo.Character.RoomId,
					Username:      userInfo.Username,
					CharacterName: userInfo.Character.Name,
				})
			}

		} else if sys.Command == `logoff` {
			w.logOff(sys.Data.(int))
		}

	}

	//
	// Handle Turn Queue
	//
	eq = events.GetQueue(events.NewTurn{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Handle RoomAction Queue
	// Needs a major overhaul/change to how it works.
	//
	eq = events.GetQueue(events.RoomAction{})
	for eq.Len() > 0 {

		e := eq.Poll()

		action, typeOk := e.(events.RoomAction)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "RoomAction", "Actual Type", e.Type())
			continue
		}

		//mudlog.Debug(`Event`, `type`, action.Type(), `RoomId`, action.RoomId, `SourceUserId`, action.SourceUserId, `SourceMobId`, action.SourceMobId, `WaitTurns`, action.WaitTurns, `Action`, action.Action)

		if action.ReadyTurn > turnNow {

			if int(action.ReadyTurn-turnNow)%c.TurnsPerRound() == 0 {
				// Get the parts of the command
				parts := strings.SplitN(action.Action, ` `, 3)
				if parts[0] == `detonate` {
					// Make sure the room exists
					room := rooms.LoadRoom(action.RoomId)
					if room == nil {
						continue
					}

					var itemName string

					if len(parts) > 2 {
						itemName = parts[2]
					} else {
						itemName = parts[1]
					}

					itm, found := room.FindOnFloor(itemName, false)
					if !found {
						continue
					}

					room.SendText(fmt.Sprintf(`The <ansi fg="itemname">%s</ansi> looks like it's about to explode...`, itm.DisplayName()))
				}

			}

			events.Requeue(action)
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		// Make sure the room exists
		room := rooms.LoadRoom(action.RoomId)
		if room == nil {
			continue
		}

		if action.Action == `mutator` {

			mutName := action.Details.(string)

			if mutName == `wildfire` {
				if room.GetBiome().Burns() && room.Mutators.Add(mutName) {
					room.SendText(colorpatterns.ApplyColorPattern(`A wildfire burns through the area!`, `flame`, colorpatterns.Stretch))
				}
				continue
			}

			room.Mutators.Add(mutName)

			continue
		}

		// Get the parts of the command
		parts := strings.SplitN(action.Action, ` `, 3)

		// Is it a detonation?
		// Possible formats:
		// donate [#mobId|@userId] !itemId:uid
		// TODO: Refactor this into a scripted event/function
		if parts[0] == `detonate` {

			// Detonate can't be the only information
			if len(parts) < 2 {
				continue
			}

			var itemName string
			var targetName string

			if len(parts) > 2 {
				targetName = parts[1]
				itemName = parts[2]
			} else {
				itemName = parts[1]
			}

			itm, found := room.FindOnFloor(itemName, false)
			if !found {
				continue
			}

			iSpec := itm.GetSpec()
			if iSpec.Type != items.Grenade {
				continue
			}

			room.RemoveItem(itm, false)

			room.SendText(`<ansi fg="red">--- --- --- --- --- --- --- --- --- --- --- ---</ansi>`)
			room.SendText(fmt.Sprintf(`The <ansi fg="itemname">%s</ansi> <ansi fg="red">EXPLODES</ansi>!`, itm.DisplayName()))
			room.SendText(`<ansi fg="red">--- --- --- --- --- --- --- --- --- --- --- ---</ansi>`)

			room.SendTextToExits(`You hear a large <ansi fg="red">!!!EXPLOSION!!!</ansi>`, false)

			if len(iSpec.BuffIds) == 0 {
				continue
			}

			hitMobs := true
			hitPlayers := true

			targetPlayerId, targetMobId := room.FindByName(targetName)

			if targetPlayerId > 0 {
				hitMobs = false
			}

			if targetMobId > 0 {
				hitPlayers = false
			}

			events.Requeue(events.RoomAction{
				RoomId:  room.RoomId,
				Action:  `mutator`,
				Details: `Details`,
			})

			if hitPlayers {

				for _, uid := range room.GetPlayers() {

					// If not hitting self and pvp is disabled, skip
					if action.SourceUserId > 0 && action.SourceUserId != uid && configs.GetConfig().PVP != `enabled` {
						continue
					}

					for _, buffId := range iSpec.BuffIds {
						events.AddToQueue(events.Buff{
							UserId:        uid,
							MobInstanceId: 0,
							BuffId:        buffId,
						})
					}
				}

			}

			if !hitMobs {
				continue
			}

			for _, mid := range room.GetMobs() {

				for _, buffId := range iSpec.BuffIds {
					events.AddToQueue(events.Buff{
						UserId:        0,
						MobInstanceId: mid,
						BuffId:        buffId,
					})
				}

				if action.SourceUserId == 0 {
					continue
				}

				sourceUser := users.GetByUserId(action.SourceUserId)
				if sourceUser == nil {
					continue
				}

				mob := mobs.GetInstance(mid)
				if mob == nil {
					continue
				}

				mob.Character.TrackPlayerDamage(sourceUser.UserId, 0)

				if sourceUser.Character.RoomId == mob.Character.RoomId {
					// Mobs get aggro when attacked
					if mob.Character.Aggro == nil {
						mob.PreventIdle = true

						mob.Command(fmt.Sprintf("attack %s", sourceUser.ShorthandId()))

					}
				} else {

					var foundExitName string

					// Look for them nearby and go to them
					for exitName, exitInfo := range room.Exits {
						if exitInfo.RoomId == sourceUser.Character.RoomId {
							foundExitName = exitName
							break
						}
					}

					if foundExitName == `` {
						// Look for them nearby and go to them
						for exitName, exitInfo := range room.ExitsTemp {
							if exitInfo.RoomId == sourceUser.Character.RoomId {

								mob.Command(fmt.Sprintf("go %s", exitName))
								mob.Command(fmt.Sprintf("attack %s", sourceUser.ShorthandId()))

								break
							}
						}
					}

					if foundExitName != `` {

						mob.Command(fmt.Sprintf("go %s", foundExitName))
						mob.Command(fmt.Sprintf("attack %s", sourceUser.ShorthandId()))

					}
				}

			}

		}

	}

	//
	// Handle Buff Queue
	//
	eq = events.GetQueue(events.Buff{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Handle Quest Queue
	//
	eq = events.GetQueue(events.Quest{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Handle NewRound events
	//
	eq = events.GetQueue(events.NewRound{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Dispatch GMCP events
	//
	eq = events.GetQueue(events.GMCPOut{})
	for eq.Len() > 0 {

		e := eq.Poll()

		gmcp, typeOk := e.(events.GMCPOut)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "GMCPOut", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		if gmcp.UserId < 1 {
			continue
		}

		connId := users.GetConnectionId(gmcp.UserId)
		if connId == 0 {
			continue
		}

		switch v := gmcp.Payload.(type) {
		case []byte:
			connections.SendTo(term.GmcpPayload.BytesWithPayload(v), connId)
		case string:
			connections.SendTo(term.GmcpPayload.BytesWithPayload([]byte(v)), connId)
		default:
			payload, err := json.Marshal(gmcp.Payload)
			if err != nil {
				mudlog.Error("Event", "Type", "GMCPOut", "data", gmcp.Payload, "error", err)
				continue
			}
			connections.SendTo(term.GmcpPayload.BytesWithPayload(payload), connId)
		}

	}

	//
	// Handle RoomChange events
	//
	eq = events.GetQueue(events.RoomChange{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Dispatch MSP events
	//
	eq = events.GetQueue(events.MSP{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	//
	// What follows are communication based events
	// (Events expected to send output to users)
	//
	//
	redrawPrompts := make(map[uint64]string)

	//
	// System-wide broadcasts
	//
	eq = events.GetQueue(events.Broadcast{})
	for eq.Len() > 0 {

		e := eq.Poll()

		broadcast, typeOk := e.(events.Broadcast)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "Broadcast", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		messageColorized := templates.AnsiParse(broadcast.Text)

		var sentToConnectionIds []connections.ConnectionId

		//
		// If it's communication, respect deafeaning rules
		//
		skipConnectionIds := []connections.ConnectionId{}
		if broadcast.IsCommunication {
			for _, u := range users.GetAllActiveUsers() {
				if u.Deafened && !broadcast.SourceIsMod {
					skipConnectionIds = append(skipConnectionIds, u.ConnectionId())
				}
			}
		}

		if broadcast.SkipLineRefresh {

			sentToConnectionIds = connections.Broadcast(
				[]byte(messageColorized),
				skipConnectionIds...,
			)

		} else {

			sentToConnectionIds = connections.Broadcast(
				[]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+messageColorized),
				skipConnectionIds...,
			)

		}

		for _, connId := range sentToConnectionIds {
			if _, ok := redrawPrompts[connId]; !ok {
				user := users.GetByConnectionId(connId)
				redrawPrompts[connId] = templates.AnsiParse(user.GetCommandPrompt(true))
			}
		}
	}

	eq = events.GetQueue(events.WebClientCommand{})
	for eq.Len() > 0 {

		e := eq.Poll()

		cmd, typeOk := e.(events.WebClientCommand)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "Message", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		if !connections.IsWebsocket(cmd.ConnectionId) {
			continue
		}

		connections.SendTo([]byte(cmd.Text), cmd.ConnectionId)

	}

	//
	// Outbound text strings
	//
	eq = events.GetQueue(events.Message{})
	for eq.Len() > 0 {

		e := eq.Poll()

		message, typeOk := e.(events.Message)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "Message", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		//mudlog.Debug("Message{}", "userId", message.UserId, "roomId", message.RoomId, "length", len(messageColorized), "IsCommunication", message.IsCommunication)

		if message.UserId > 0 {

			if user := users.GetByUserId(message.UserId); user != nil {

				// If they are deafened, they cannot hear user communications
				if message.IsCommunication && user.Deafened {
					continue
				}

				connections.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+templates.AnsiParse(message.Text)), user.ConnectionId())
				if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
					redrawPrompts[user.ConnectionId()] = templates.AnsiParse(user.GetCommandPrompt(true))
				}

			}
		}

		if message.RoomId > 0 {

			room := rooms.LoadRoom(message.RoomId)
			if room == nil {
				continue
			}

			for _, userId := range room.GetPlayers() {
				skip := false

				if message.UserId == userId {
					continue
				}

				exLen := len(message.ExcludeUserIds)
				if exLen > 0 {
					for _, excludeId := range message.ExcludeUserIds {
						if excludeId == userId {
							skip = true
							break
						}
					}
				}

				if skip {
					continue
				}

				if user := users.GetByUserId(userId); user != nil {

					// If they are deafened, they cannot hear user communications
					if message.IsCommunication && user.Deafened {
						continue
					}

					// If this is a quiet message, make sure the player can hear it
					if message.IsQuiet {
						if !user.Character.HasBuffFlag(buffs.SuperHearing) {
							continue
						}
					}

					connections.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+templates.AnsiParse(message.Text)), user.ConnectionId())
					if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
						redrawPrompts[user.ConnectionId()] = templates.AnsiParse(user.GetCommandPrompt(true))
					}

				}
			}

		}

	}

	for connectionId, prompt := range redrawPrompts {
		connections.SendTo([]byte(prompt), connectionId)
	}
}
