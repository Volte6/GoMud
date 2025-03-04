package usercommands

import (
	"errors"
	"fmt"
	"math"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Suicide(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	config := configs.GetConfig()
	currentRound := util.GetRoundCount()

	if user.Character.Zone == `Shadow Realm` {
		user.SendText(`You're already dead!`)
		return true, errors.New(`already dead`)
	}

	if user.Character.HasBuffFlag(buffs.ReviveOnDeath) {

		user.Character.Health = user.Character.HealthMax.Value

		user.SendText(`You are revived in a shower of magical sparks!`)
		room.SendText(`<ansi fg="username">`+user.Character.Name+`</ansi> is suddenly revived in a shower of sparks!`, user.UserId)

		user.Character.CancelBuffsWithFlag(buffs.ReviveOnDeath)

		return true, nil
	}

	// Send a death msg to everyone in the room.
	room.SendText(
		fmt.Sprintf(`<ansi fg="username">%s</ansi> has died.`, user.Character.Name),
		user.UserId,
	)

	i := 0
	dmgCt := len(user.Character.PlayerDamage)

	if dmgCt > 0 {
		user.Character.KD.AddPvpDeath()
	} else {
		user.Character.KD.AddMobDeath()
	}

	killedByUserIds := []int{}
	killedBy := ``
	for uid, _ := range user.Character.PlayerDamage {

		if u := users.GetByUserId(uid); u != nil {

			// Update PK stats
			user.Character.KD.AddPlayerDeath(u.UserId, u.Character.Name)
			u.Character.KD.AddPlayerKill(user.UserId, user.Character.Name)

			if i > 0 {
				if i < dmgCt-1 {
					killedBy += ` and `
				} else {
					killedBy += `, `
				}
			}
			killedBy += `<ansi fg="username">` + u.Character.Name + `</ansi>`
			i++
		}

		killedByUserIds = append(killedByUserIds, uid)
	}

	msg := fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> <ansi fg="magenta-bold">***</ansi>%s`, user.Character.Name, term.CRLFStr)
	if killedBy != `` {
		msg = fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> (killed by %s) <ansi fg="magenta-bold">***</ansi>%s`, user.Character.Name, killedBy, term.CRLFStr)
	}

	events.AddToQueue(events.Broadcast{
		Text: msg,
	})

	allowPenalties := user.Character.Level > int(config.OnDeathProtectionLevels)

	events.AddToQueue(events.PlayerDeath{
		UserId:        user.UserId,
		RoomId:        user.Character.RoomId,
		Username:      user.Username,
		CharacterName: user.Character.Name,
		Permanent:     allowPenalties && bool(config.PermaDeath) && user.Character.ExtraLives == 0,
		KilledByUsers: killedByUserIds,
	})

	// If permadeath is enabled, do some extra bookkeeping
	if allowPenalties && bool(config.PermaDeath) {

		if user.Character.ExtraLives > 0 {

			user.Character.ExtraLives--

		} else {

			user.EventLog.Add(`death`, fmt.Sprintf(`<ansi fg="username">%s</ansi> has <ansi fg="red-bold">PERMA-DIED</ansi>`, user.Character.Name))

			// Perma-died!!!
			textOut, _ := templates.Process("character/permadeath", nil)
			user.SendText(colorpatterns.ApplyColorPattern(textOut, `red`))

			// Unequip everything
			for _, itm := range user.Character.GetAllWornItems() {
				Remove(itm.Name(), user, room, flags)
			}
			// drop all items / gold
			Drop("all", user, room, flags)

			rooms.MoveToRoom(user.UserId, -1)

			user.Character = characters.New()

			return true, nil
		}

	}

	user.EventLog.Add(`death`, fmt.Sprintf(`<ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED</ansi>`, user.Character.Name))

	// Only apply penalties if they were above the threshold
	if allowPenalties {

		if config.OnDeathEquipmentDropChance >= 0 {
			chanceInt := int(config.OnDeathEquipmentDropChance * 100)
			for _, itm := range user.Character.GetAllWornItems() {
				if util.Rand(100) < chanceInt {

					Remove(itm.Name(), user, room, flags)

					Drop(itm.Name(), user, room, flags)

				}
			}
		}

		if user.Character.Gold > 0 {
			user.EventLog.Add(`death`, fmt.Sprintf(`Dropped <ansi fg="gold">%d gold</ansi> on death`, user.Character.Gold))
			Drop(fmt.Sprintf(`%d gold`, user.Character.Gold), user, room, flags)
		}

		if config.OnDeathAlwaysDropBackpack {
			Drop("all", user, room, flags)

			user.EventLog.Add(`death`, `Dropped <ansi fg="alert-3">everthing in your backpack</ansi> on death`)

		} else if config.OnDeathEquipmentDropChance >= 0 {
			chanceInt := int(config.OnDeathEquipmentDropChance * 100)
			for _, itm := range user.Character.GetAllBackpackItems() {
				if util.Rand(100) < chanceInt {
					Drop(itm.Name(), user, room, flags)
					user.EventLog.Add(`death`, fmt.Sprintf(`Dropped your <ansi fg="itemname">%s</ansi> on death`, itm.Name()))
				}
			}
		}

		if user.Character.Level > 1 {

			setting, lossPct := config.GetDeathXPPenalty()
			if setting != `none` {

				if setting == `level` { // are they being brought down to the base of their current level?
					user.Character.Level--
					oldExperience := user.Character.Experience
					user.Character.Experience = user.Character.XPTNL()
					user.Character.Level++

					user.SendText(fmt.Sprintf(`You lost <ansi fg="yellow">%d experience points</ansi>.`, oldExperience-user.Character.Experience))

					user.EventLog.Add(`death`, fmt.Sprintf(`Lost <ansi fg="yellow">%d experience points</ansi> on death`, oldExperience-user.Character.Experience))

				} else if lossPct > 0 { // Are they losing a set %?

					loss := int(math.Floor(float64(user.Character.Experience) * lossPct))
					user.Character.Experience -= loss

					user.SendText(fmt.Sprintf(`You lost <ansi fg="yellow">%d experience points</ansi>.`, loss))

					user.EventLog.Add(`death`, fmt.Sprintf(`Lost <ansi fg="yellow">%d experience points</ansi> on death`, loss))
				}
			}

		}

	}

	user.Character.CancelBuffsWithFlag(buffs.All)

	user.Character.Health = -10
	user.Character.Mana = 0

	clear(user.Character.PlayerDamage)

	rooms.MoveToRoom(user.UserId, int(config.DeathRecoveryRoom))

	if config.CorpsesEnabled {
		room.AddCorpse(rooms.Corpse{
			UserId:       user.UserId,
			Character:    *user.Character,
			RoundCreated: currentRound,
		})
	}

	return true, nil
}
