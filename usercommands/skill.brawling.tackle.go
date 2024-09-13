package usercommands

import (
	"fmt"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Brawling Skill
Level 3 - Attempt to tackle an opponent, making them miss a round.
*/
func Tackle(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Brawling)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 3 {
		return response, nil
	}

	if user.Character.Aggro == nil {
		response.SendUserMessage(userId, "Tackle is only used while in combat!", true)
		response.Handled = true
		return response, nil
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	if !user.Character.TryCooldown(skills.Brawling.String(`tackle`), 5) {
		response.SendUserMessage(userId, "You are too tired to tackle again so soon!", true)
		response.Handled = true
		return response, nil
	}

	attackMobInstanceId := user.Character.Aggro.MobInstanceId
	attackPlayerId := user.Character.Aggro.UserId

	if attackMobInstanceId > 0 {

		m := mobs.GetInstance(attackMobInstanceId)

		if m != nil {

			chanceIn100 := user.Character.Stats.Speed.ValueAdj - m.Character.Stats.Perception.ValueAdj
			if chanceIn100 < 0 {
				chanceIn100 = 0
			}
			chanceIn100 += 10
			roll := util.Rand(100)

			util.LogRoll(`Tackle`, roll, chanceIn100)

			if roll < chanceIn100 {

				response.SendUserMessage(userId,
					fmt.Sprintf(`You lunge and tackle <ansi fg="mobname">%s</ansi>!`, m.Character.Name),
					true)

				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges and tackles <ansi fg="mobname">%s</ansi>!`, user.Character.Name, m.Character.Name),
					true,
					userId,
				)

				events.AddToQueue(events.Buff{
					UserId:        0,
					MobInstanceId: attackMobInstanceId,
					BuffId:        12, // buff 12 is tackled
				})

			} else {
				response.SendUserMessage(userId,
					fmt.Sprintf(`You try to tackle <ansi fg="mobname">%s</ansi> and miss!`, m.Character.Name),
					true)

				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to tackle <ansi fg="mobname">%s</ansi> and misses!`, user.Character.Name, m.Character.Name),
					true,
					userId,
				)

			}
		}
	} else if attackPlayerId > 0 {

		u := users.GetByUserId(attackPlayerId)

		if u != nil {

			chanceIn100 := user.Character.Stats.Speed.ValueAdj - u.Character.Stats.Perception.ValueAdj
			if chanceIn100 < 0 {
				chanceIn100 = 0
			}
			chanceIn100 += 10
			roll := util.Rand(100)

			util.LogRoll(`Tackle`, roll, chanceIn100)

			if roll < chanceIn100 {

				response.SendUserMessage(userId,
					fmt.Sprintf(`You lunge and tackle <ansi fg="username">%s</ansi>!`, u.Character.Name),
					true)

				response.SendUserMessage(attackPlayerId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges and tackles you!`, user.Character.Name),
					true)

				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges and tackles <ansi fg="username">%s</ansi>!`, user.Character.Name, u.Character.Name),
					true,
					userId,
					attackPlayerId,
				)

				events.AddToQueue(events.Buff{
					UserId:        attackPlayerId,
					MobInstanceId: 0,
					BuffId:        12, // buff 12 is tackled
				})

			} else {
				response.SendUserMessage(userId,
					fmt.Sprintf(`You lunge to tackle <ansi fg="username">%s</ansi> and miss!`, u.Character.Name),
					true)

				response.SendUserMessage(attackPlayerId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges to tackles you and misses!`, user.Character.Name),
					true)

				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> lunges to tackle <ansi fg="username">%s</ansi> and misses!`, user.Character.Name, u.Character.Name),
					true,
					userId,
					attackPlayerId,
				)

			}
		}
	}

	response.Handled = true
	return response, nil
}
