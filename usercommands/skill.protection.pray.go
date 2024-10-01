package usercommands

import (
	"errors"
	"fmt"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Protection Skill
Level 4 - Pray to gods for a blessing
*/
func Pray(rest string, user *users.UserRecord) (bool, error) {

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Protection)

	if skillLevel < 4 {
		user.SendText("You don't know how to pray.")
		return true, fmt.Errorf("you don't know how to pray")
	}

	if !user.Character.TryCooldown(skills.Protection.String(), configs.GetConfig().MinutesToRounds(5)) {
		user.SendText(
			`You can only pray once every 5 minutes.`,
		)
		return true, errors.New(`you can only pray once every 5 minutes`)
	}

	prayPlayerId, prayMobId := 0, 0

	if rest == `` {
		prayPlayerId = user.UserId
	} else {
		prayPlayerId, prayMobId = room.FindByName(rest)
	}

	if prayPlayerId == 0 && prayMobId == 0 {
		user.SendText("Aid whom?")
		return true, nil
	}

	possibleBuffIds := []int{4, 11, 14, 16, 17, 18}
	totalBuffCount := 1 + int(float64(user.Character.Stats.Mysticism.ValueAdj)/15) + util.Rand(2)

	if totalBuffCount > len(possibleBuffIds) {
		totalBuffCount = len(possibleBuffIds)
	}

	if prayPlayerId > 0 {

		if prayPlayerId == user.UserId {
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> begins to pray.`, user.Character.Name), user.UserId)
		} else {
			targetUser := users.GetByUserId(prayPlayerId)
			if targetUser != nil {
				room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> puts his hand over <ansi fg="username">%s</ansi> and begins to pray.`, user.Character.Name, targetUser.Character.Name), user.UserId, targetUser.UserId)
				targetUser.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> puts his hand over you and begins to pray.`, user.Character.Name))
			}
		}

		for i := 0; i < totalBuffCount; i++ {
			randBuffIndex := util.Rand(len(possibleBuffIds))

			events.AddToQueue(events.Buff{
				UserId:        prayPlayerId,
				MobInstanceId: 0,
				BuffId:        possibleBuffIds[randBuffIndex],
			})

			possibleBuffIds = append(possibleBuffIds[:randBuffIndex], possibleBuffIds[randBuffIndex+1:]...)
			room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> glows for a moment.`, user.Character.Name))
		}

	} else if prayMobId > 0 {

		if mob := mobs.GetInstance(prayMobId); mob != nil {
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> puts his hand over <ansi fg="mobname">%s</ansi> and begins to pray.`, user.Character.Name, mob.Character.Name), user.UserId)

			for i := 0; i < totalBuffCount; i++ {
				randBuffIndex := util.Rand(len(possibleBuffIds))

				events.AddToQueue(events.Buff{
					UserId:        0,
					MobInstanceId: prayMobId,
					BuffId:        possibleBuffIds[randBuffIndex],
				})

				possibleBuffIds = append(possibleBuffIds[:randBuffIndex], possibleBuffIds[randBuffIndex+1:]...)
				room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> glows for a moment.`, mob.Character.Name))
			}
		}

	} else {
		user.SendText("Pray for whom?")
	}

	return true, nil
}
