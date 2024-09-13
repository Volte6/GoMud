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
				response.SendUserMessage(user.UserId, "You feel a curse lifted from your weapon.")
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Weapon.NameSimple()), user.UserId)
			}

			if user.Character.Equipment.Offhand.IsCursed() {
				user.Character.Equipment.Offhand.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Offhand.NameSimple()))
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Offhand.NameSimple()), user.UserId)
			}

			if user.Character.Equipment.Head.IsCursed() {
				user.Character.Equipment.Head.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Head.NameSimple()))
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Head.NameSimple()), user.UserId)
			}

			if user.Character.Equipment.Neck.IsCursed() {
				user.Character.Equipment.Neck.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Neck.NameSimple()))
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Neck.NameSimple()), user.UserId)
			}

			if user.Character.Equipment.Body.IsCursed() {
				user.Character.Equipment.Body.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Body.NameSimple()))
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Body.NameSimple()), user.UserId)
			}

			if user.Character.Equipment.Belt.IsCursed() {
				user.Character.Equipment.Belt.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Belt.NameSimple()))
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Belt.NameSimple()), user.UserId)
			}

			if user.Character.Equipment.Gloves.IsCursed() {
				user.Character.Equipment.Gloves.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Gloves.NameSimple()))
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Gloves.NameSimple()), user.UserId)
			}

			if user.Character.Equipment.Ring.IsCursed() {
				user.Character.Equipment.Ring.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Ring.NameSimple()))
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Ring.NameSimple()), user.UserId)
			}

			if user.Character.Equipment.Legs.IsCursed() {
				user.Character.Equipment.Legs.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Legs.NameSimple()))
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Legs.NameSimple()), user.UserId)
			}

			if user.Character.Equipment.Feet.IsCursed() {
				user.Character.Equipment.Feet.Uncursed = true
				response.SendUserMessage(user.UserId, fmt.Sprintf(`You feel a curse lifted from your <ansi fg="itemname">%s</ansi>.`, user.Character.Equipment.Feet.NameSimple()))
				response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi>'s <ansi fg="itemname">%s</ansi> glows briefly.`, user.Character.Name, user.Character.Equipment.Feet.NameSimple()), user.UserId)
			}

		}
	}

	response.Handled = true
	return response, nil
}
