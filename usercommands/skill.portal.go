package usercommands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Portal Skill
Level 1 - Teleport back to town square
Level 2 - Teleport back to the root of the area you are in
Level 3 - Set a new destination for your portal teleportation
Level 4 - Create a physical portal that you can share with players, or return through.
*/
func Portal(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// This is a hack because using "portal" to enter an existing portal is very common
	if rest == `` {
		if response, err := Go(`portal`, userId, cmdQueue); response.Handled {
			return response, err
		}
	}

	skillLevel := user.Character.GetSkillLevel(skills.Portal)

	if skillLevel == 0 {
		response.SendUserMessage(userId, "You don't know how to portal.", true)
		response.Handled = true
		return response, errors.New(`you don't know how to portal`)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	// Establish the default portal location
	portalTargetRoomId := 1 // Defaults to Town Square of Frostfang

	if skillLevel >= 2 { // Defaults to root of current zone
		portalTargetRoomId, _ = rooms.GetZoneRoot(user.Character.Zone)
		if portalTargetRoomId == 75 {
			portalTargetRoomId = 1 // If they are in the holding zone, send htem back to TS
		}
	}

	if skillLevel >= 3 { // Defaults to wherever the player last set it
		portalSetting := user.Character.GetSetting("portal")
		if portalSetting != "" {
			if settingRoomId, err := strconv.Atoi(portalSetting); err == nil {
				portalTargetRoomId = settingRoomId
			}
		}
	}

	portalLifeInSeconds := 30 + (user.Character.Stats.Mysticism.ValueAdj * 10) // 0 mysticism = 30 seconds, 100 mysticism = 1030 seconds

	// Make sure we haven't borked anything
	if portalTargetRoomId < 1 {
		portalTargetRoomId = 1
	}

	// If no argument supplied, is a direct teleport.
	if rest == "" {

		if user.Character.Aggro != nil {
			response.SendUserMessage(userId, "You can't do that! You are in combat!", true)
			response.Handled = true
			return response, nil
		}

		if !user.Character.TryCooldown(skills.Portal.String(), 10) {
			response.SendUserMessage(userId,
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Portal.String())),
				true)
			response.Handled = true
			return response, errors.New(`you're doing that too often`)
		}

		// move to portalTargetRoomId
		response.Handled = true
		oldRoomId := user.Character.RoomId

		if err := rooms.MoveToRoom(user.UserId, portalTargetRoomId); err == nil {

			response.SendUserMessage(userId, "You draw a quick symbol in the air with your finger, and the world warps around you. You seem to have moved.", true)
			response.SendRoomMessage(oldRoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> draws a quick symbol in the air, and is sucked into a portal!`, user.Character.Name),
				true)
			response.SendRoomMessage(portalTargetRoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> suddenly pops into existence!`, user.Character.Name),
				true)
		} else {
			response.SendUserMessage(userId, "Oops, portal sad!", true)
		}
		response.Handled = true
		return response, nil
	}

	if skillLevel >= 3 {

		if rest == "set" {
			user.Character.SetSetting("portal", strconv.Itoa(user.Character.RoomId))

			response.SendUserMessage(userId, "You enscribe a glowing pentagram on the ground with your finger, which then fades away. Your portals now lead to this area.", true)
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> enscribes a glowing pentagram on the ground with their finger. It quickly fades away.`, user.Character.Name),
				true)
		}

		if rest == "unset" || rest == "clear" {
			user.Character.SetSetting("portal", "")

			response.SendUserMessage(userId, "You draw an arcane symbol in the air with your finger. Your portals now lead to their default location.", true)
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> draws a shape in the air with their finger. The floor trembles mildly, but then returns to normal.`, user.Character.Name),
				true)
		}
	}

	if skillLevel == 4 {

		if rest == "open" {

			// Load current room details
			targetRoom := rooms.LoadRoom(portalTargetRoomId)
			if targetRoom == nil {
				return response, fmt.Errorf(`room %d not found`, portalTargetRoomId)
			}

			if portalTargetRoomId == user.Character.RoomId {
				response.SendUserMessage(userId, "You can't open a portal to the room you're already in!", true)
				response.Handled = true
				return response, nil
			}

			if !user.Character.TryCooldown(skills.Portal.String(), 10) {
				response.SendUserMessage(userId,
					fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Portal.String())),
					true)
				response.Handled = true
				return response, errors.New(`you're doing that too often`)
			}

			// Check whether they already have a portal open, and if so, shut it down.
			if currentPortal := user.Character.GetSetting("portal:open"); len(currentPortal) > 0 {
				// Data looks like {roomId1}:{roomId2}
				// Extract the two room id's
				portalRooms := strings.Split(currentPortal, ":")
				portalRoomId1, _ := strconv.Atoi(portalRooms[0])
				portalRoomId2, _ := strconv.Atoi(portalRooms[1])

				var oldPortalsRemoved bool = false

				if r1 := rooms.LoadRoom(portalRoomId1); r1 != nil {

					if tmpExit, found := r1.FindTemporaryExitByUserId(userId); found {
						if r1.RemoveTemporaryExit(tmpExit) {
							response.SendRoomMessage(r1.RoomId,
								fmt.Sprintf("Suddenly, the %s before you snaps closed, leaving no evidence it ever existed.", templates.GlowingPortal),
								true)
							oldPortalsRemoved = true
						}
					}
				}

				if r2 := rooms.LoadRoom(portalRoomId2); r2 != nil {

					if tmpExit, found := r2.FindTemporaryExitByUserId(userId); found {
						if r2.RemoveTemporaryExit(tmpExit) {
							response.SendRoomMessage(r2.RoomId,
								fmt.Sprintf("Suddenly, the %s before you snaps closed, leaving no evidence it ever existed.", templates.GlowingPortal),
								true)
							oldPortalsRemoved = true
						}
					}
				}

				if oldPortalsRemoved {
					response.SendUserMessage(userId, "Your old portals snap shut.", true)
				}

				user.Character.SetSetting("portal:open", "")
			}

			// Target = portalTargetRoomId
			// Current = user.Character.RoomId
			// At this point we have no open portals, we can create a new one.
			newPortalExitName := fmt.Sprintf("glowing portal from %s", user.Character.Name)
			newPortal := rooms.TemporaryRoomExit{
				RoomId:  portalTargetRoomId,
				Title:   fmt.Sprintf(`%s from <ansi fg="username">%s</ansi>`, templates.GlowingPortal, user.Character.Name),
				UserId:  userId,
				Expires: time.Now().Add(time.Duration(portalLifeInSeconds) * time.Second),
			}

			// Spawn a portal in the room that leads to the portal location
			if !room.AddTemporaryExit(newPortalExitName, newPortal) {
				response.SendUserMessage(userId, "Something went wrong. That's the problem with portal!", true)
				response.Handled = true
				return response, fmt.Errorf("failed to add temporary exit to room")
			}
			response.SendUserMessage(userId,
				fmt.Sprintf("You trace the shape of a doorway in front of you with your finger, which becomes a %s to another area.", templates.GlowingPortal),
				true)
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> traces the shape of a doorway with their finger, and a %s appears!`, user.Character.Name, templates.GlowingPortal),
				true)

			// Modify it for this room
			newPortal.RoomId = user.Character.RoomId

			if !targetRoom.AddTemporaryExit(newPortalExitName, newPortal) {
				response.SendUserMessage(userId, "Something went wrong. That's the problem with portal!", true)
				response.Handled = true
				return response, fmt.Errorf("failed to add temporary exit to room")
			}

			response.SendRoomMessage(portalTargetRoomId,
				fmt.Sprintf("A %s suddenly appears!", templates.GlowingPortal),
				true)

			user.Character.SetSetting("portal:open", fmt.Sprintf("%d:%d", user.Character.RoomId, portalTargetRoomId))
		}
	}

	response.Handled = true
	return response, nil
}
