package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Uncurse(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load mob details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)

	for _, pid := range room.GetPlayers() {

		if user := users.GetByUserId(pid); user != nil {

			if user.Character.Equipment.Weapon.IsCursed() {
				user.Character.Equipment.Weapon.Uncursed = true
				response.SendUserMessage(user.UserId, "You feel a curse lifted from your weapon.", true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Weapon.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Offhand.IsCursed() {
				user.Character.Equipment.Offhand.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Offhand.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Offhand.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Head.IsCursed() {
				user.Character.Equipment.Head.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Head.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Head.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Neck.IsCursed() {
				user.Character.Equipment.Neck.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Neck.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Neck.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Body.IsCursed() {
				user.Character.Equipment.Body.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Body.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Body.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Belt.IsCursed() {
				user.Character.Equipment.Belt.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Belt.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Belt.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Gloves.IsCursed() {
				user.Character.Equipment.Gloves.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Gloves.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Gloves.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Ring.IsCursed() {
				user.Character.Equipment.Ring.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Ring.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Ring.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Legs.IsCursed() {
				user.Character.Equipment.Legs.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Legs.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Legs.NameSimple()), true, user.UserId)
			}

			if user.Character.Equipment.Feet.IsCursed() {
				user.Character.Equipment.Feet.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Feet.NameSimple()), true)
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Feet.NameSimple()), true, user.UserId)
			}

		}
	}

	response.Handled = true
	return response, nil
}
