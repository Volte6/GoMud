package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Uncurse(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load mob details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)

	for _, pid := range room.GetPlayers() {

		if user := users.GetByUserId(pid); user != nil {

			if user.Character.Equipment.Weapon.GetSpec().Cursed && !user.Character.Equipment.Weapon.Uncursed {
				user.Character.Equipment.Weapon.Uncursed = true
				response.SendUserMessage(user.UserId, "You feel a curse lifted from your weapon.", true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Weapon.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Offhand.GetSpec().Cursed && !user.Character.Equipment.Offhand.Uncursed {
				user.Character.Equipment.Offhand.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Offhand.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Offhand.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Head.GetSpec().Cursed && !user.Character.Equipment.Head.Uncursed {
				user.Character.Equipment.Head.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Head.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Head.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Neck.GetSpec().Cursed && !user.Character.Equipment.Neck.Uncursed {
				user.Character.Equipment.Neck.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Neck.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Neck.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Body.GetSpec().Cursed && !user.Character.Equipment.Body.Uncursed {
				user.Character.Equipment.Body.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Body.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Body.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Belt.GetSpec().Cursed && !user.Character.Equipment.Belt.Uncursed {
				user.Character.Equipment.Belt.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Belt.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Belt.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Gloves.GetSpec().Cursed && !user.Character.Equipment.Gloves.Uncursed {
				user.Character.Equipment.Gloves.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Gloves.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Gloves.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Ring.GetSpec().Cursed && !user.Character.Equipment.Ring.Uncursed {
				user.Character.Equipment.Ring.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Ring.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Ring.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Legs.GetSpec().Cursed && !user.Character.Equipment.Legs.Uncursed {
				user.Character.Equipment.Legs.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Legs.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Legs.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Feet.GetSpec().Cursed && !user.Character.Equipment.Feet.Uncursed {
				user.Character.Equipment.Feet.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Feet.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Feet.NameSimple()), true, user.UserId)
			}

		}
	}

	response.Handled = true
	return response, nil
}
