package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func (w *World) roundTick() {

	c := configs.GetConfig()

	roundNumber := util.IncrementRoundCount()

	if c.LogIntervalRoundCount > 0 && roundNumber%uint64(c.LogIntervalRoundCount) == 0 {
		slog.Info("World::RoundTick()", "roundNumber", roundNumber)
	}

	events.AddToQueue(events.NewRound{RoundNumber: roundNumber, TimeNow: time.Now(), Config: c})
}

func (w *World) logOff(userId int) {

	if user := users.GetByUserId(userId); user != nil {

		user.EventLog.Add(`conn`, `Logged off`)

		users.SaveUser(*user)

		worldManager.leaveWorld(userId)

		connId := user.ConnectionId()

		tplTxt, _ := templates.Process("goodbye", nil, templates.AnsiTagsPreParse)

		connections.SendTo([]byte(tplTxt), connId)

		if err := users.LogOutUserByConnectionId(connId); err != nil {
			slog.Error("Log Out Error", "connectionId", connId, "error", err)
		}

		connections.Remove(connId)

	}

}

func (w *World) PruneBuffs() {

	roomsWithPlayers := rooms.GetRoomsWithPlayers()
	for _, roomId := range roomsWithPlayers {
		// Get rooom
		if room := rooms.LoadRoom(roomId); room != nil {

			// Handle outstanding player buffs
			logOff := false
			for _, uId := range room.GetPlayers(rooms.FindBuffed) {

				user := users.GetByUserId(uId)

				logOff = false
				if buffsToPrune := user.Character.Buffs.Prune(); len(buffsToPrune) > 0 {
					for _, buffInfo := range buffsToPrune {
						scripting.TryBuffScriptEvent(`onEnd`, uId, 0, buffInfo.BuffId)

						if buffInfo.BuffId == 0 { // Log them out // logoff // logout
							if !user.Character.HasAdjective(`zombie`) { // if they are currently a zombie, we don't log them out from this buff being removed
								logOff = true
							}
						}
					}

					user.Character.Validate()

					if logOff {
						slog.Info("MEDITATION LOGOFF")
						w.logOff(uId)
					}
				}

			}
		}
	}

	// Handle outstanding mob buffs
	for _, mobInstanceId := range mobs.GetAllMobInstanceIds() {

		mob := mobs.GetInstance(mobInstanceId)

		if buffsToPrune := mob.Character.Buffs.Prune(); len(buffsToPrune) > 0 {
			for _, buffInfo := range buffsToPrune {
				scripting.TryBuffScriptEvent(`onEnd`, 0, mobInstanceId, buffInfo.BuffId)
			}

			mob.Character.Validate()
		}

	}

}

// Handle dropped players
func (w *World) HandleDroppedPlayers(droppedPlayers []int) {

	if len(droppedPlayers) == 0 {
		return
	}

	for _, userId := range droppedPlayers {
		if user := users.GetByUserId(userId); user != nil {

			user.SendText(`<ansi fg="red">you drop to the ground!</ansi>`)

			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> <ansi fg="red">drops to the ground!</ansi>`, user.Character.Name),
					user.UserId)
			}
		}
	}

	return
}

// Levelups
func (w *World) CheckForLevelUps() {

	onlineIds := users.GetOnlineUserIds()
	for _, userId := range onlineIds {
		user := users.GetByUserId(userId)

		if newLevel, statsDelta := user.Character.LevelUp(); newLevel {

			livesBefore := user.Character.ExtraLives

			c := configs.GetConfig()
			if c.PermaDeath && c.LivesOnLevelUp > 0 {
				user.Character.ExtraLives += int(c.LivesOnLevelUp)
				if user.Character.ExtraLives > int(c.LivesMax) {
					user.Character.ExtraLives = int(c.LivesMax)
				}
			}

			user.EventLog.Add(`xp`, fmt.Sprintf(`<ansi fg="username">%s</ansi> is now <ansi fg="magenta-bold">level %d</ansi>!`, user.Character.Name, user.Character.Level))

			levelUpData := map[string]interface{}{
				"level":          user.Character.Level,
				"statsDelta":     statsDelta,
				"trainingPoints": 1,
				"statPoints":     1,
				"livesUp":        user.Character.ExtraLives - livesBefore,
			}
			levelUpStr, _ := templates.Process("character/levelup", levelUpData)

			user.SendText(levelUpStr)

			events.AddToQueue(events.Broadcast{
				Text: fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> <ansi fg="yellow">has leveled up to level %d!</ansi> <ansi fg="magenta-bold">***</ansi>%s`, user.Character.Name, user.Character.Level, term.CRLFStr),
			})

			if user.Character.Level >= 5 {
				for _, mobInstanceId := range user.Character.CharmedMobs {
					if mob := mobs.GetInstance(mobInstanceId); mob != nil {

						if mob.MobId == 38 {
							mob.Command(`say I see you have grown much stronger and more experienced. My assistance is now needed elsewhere. I wish you good luck!`)
							mob.Command(`emote clicks their heels together and disappears in a cloud of smoke.`, 10)
							mob.Command(`suicide vanish`, 10)
						}
					}
				}
			}

			user.PlaySound(`levelup`, `other`)

			users.SaveUser(*user)

			continue
		}

	}

}
