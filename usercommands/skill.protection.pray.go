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
func Pray(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Protection)

	if skillLevel < 4 {
		response.SendUserMessage(userId, "You don't know how to pray.", true)
		response.Handled = true
		return response, fmt.Errorf("you don't know how to pray")
	}

	if !user.Character.TryCooldown(skills.Protection.String(), configs.GetConfig().MinutesToRounds(5)) {
		response.SendUserMessage(userId,
			`You can only pray once every 5 minutes.`,
			true)
		response.Handled = true
		return response, errors.New(`you can only pray once every 5 minutes`)
	}

	prayPlayerId, prayMobId := 0, 0

	if rest == `` {
		prayPlayerId = userId
	} else {
		prayPlayerId, prayMobId = room.FindByName(rest)
	}

	if prayPlayerId == 0 && prayMobId == 0 {
		response.SendUserMessage(userId, "Aid whom?", true)
		response.Handled = true
		return response, nil
	}

	possibleBuffIds := []int{4, 11, 14, 16, 17, 18}
	totalBuffCount := 1 + int(float64(user.Character.Stats.Mysticism.ValueAdj)/15) + util.Rand(2)

	if totalBuffCount > len(possibleBuffIds) {
		totalBuffCount = len(possibleBuffIds)
	}

	if prayPlayerId > 0 {

		if prayPlayerId == userId {
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> begins to pray.`, user.Character.Name), true, user.UserId)
		} else {
			targetUser := users.GetByUserId(prayPlayerId)
			if targetUser != nil {
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> puts his hand over <ansi fg="username">%s</ansi> and begins to pray.`, user.Character.Name, targetUser.Character.Name), true, user.UserId, targetUser.UserId)
				response.SendUserMessage(targetUser.UserId, fmt.Sprintf(`<ansi fg="username">%s</ansi> puts his hand over you and begins to pray.`, user.Character.Name), true)
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
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> glows for a moment.`, user.Character.Name), true)
		}

	} else if prayMobId > 0 {

		if mob := mobs.GetInstance(prayMobId); mob != nil {
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> puts his hand over <ansi fg="mobname">%s</ansi> and begins to pray.`, user.Character.Name, mob.Character.Name), true, user.UserId)

			for i := 0; i < totalBuffCount; i++ {
				randBuffIndex := util.Rand(len(possibleBuffIds))

				events.AddToQueue(events.Buff{
					UserId:        0,
					MobInstanceId: prayMobId,
					BuffId:        possibleBuffIds[randBuffIndex],
				})

				possibleBuffIds = append(possibleBuffIds[:randBuffIndex], possibleBuffIds[randBuffIndex+1:]...)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> glows for a moment.`, mob.Character.Name), true)
			}
		}

	} else {
		response.SendUserMessage(userId, "Pray for whom?", true)
	}

	response.Handled = true
	return response, nil
}
