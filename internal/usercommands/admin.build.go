package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

/*
* Role Permissions:
* build 				(All)
 */
func Build(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.build", nil, user.UserId, user.UserId)
		user.SendText(infoOutput)
	} else {

		// #build zone "The Arctic"
		if args[0] == "zone" {

			zoneName := strings.Join(args[1:], ` `)

			if roomId, err := rooms.CreateZone(zoneName); err != nil {
				user.SendText(err.Error())
			} else {
				user.SendText(fmt.Sprintf("Zone %s created.", zoneName))

				if err := rooms.MoveToRoom(user.UserId, roomId); err != nil {
					user.SendText(err.Error())
				} else {
					user.SendText(fmt.Sprintf("Moved to room %d.", roomId))
					events.AddToQueue(events.Input{
						UserId:    user.UserId,
						InputText: `look`,
					}, -1)
				}
			}
		}

		// #build room north <south>
		if args[0] == "room" {

			exitName := args[1]
			mapDirection := exitName

			returnName := ""
			if len(args) > 2 {
				returnName = args[2]
			}

			// #build (room north) - room+north are two args
			var destinationRoom *rooms.Room = nil
			// If it's a compass direction, reject it if a room already exists in that direction

			zMapper := mapper.GetZoneMapper(room.Zone)
			if zMapper == nil {
				err := fmt.Errorf("Could not find mapper for zone: %s", room.Zone)
				mudlog.Error("Map", "error", err)
				user.SendText(`No map found (or an error occured)"`)
				return true, err
			}

			if gotoRoomId, _ := zMapper.FindAdjacentRoom(user.Character.RoomId, exitName, 1); gotoRoomId == 0 {

				if newRoom, err := rooms.BuildRoom(user.Character.RoomId, exitName, mapDirection); err != nil {
					user.SendText(err.Error())
				} else {
					destinationRoom = newRoom
				}

				if destinationRoom == nil {
					user.SendText(fmt.Sprintf("Error building room %s.", exitName))
					return false, nil
				}

			}

			// Connect the exit back
			if len(returnName) > 0 {
				returnMapDirection := returnName
				if strings.Contains(returnName, `-`) {
					returnMapDirection = returnName

					parts := strings.Split(returnName, `-`)
					returnName = parts[0]
				}

				rooms.ConnectRoom(destinationRoom.RoomId, user.Character.RoomId, returnName, returnMapDirection)
			}

			if err := rooms.MoveToRoom(user.UserId, destinationRoom.RoomId); err != nil {
				user.SendText(err.Error())
			} else {
				user.SendText(fmt.Sprintf("Moved to room %d.", destinationRoom.RoomId))

				events.AddToQueue(events.Input{
					UserId:    user.UserId,
					InputText: `look`,
				}, -1)
			}

		}

	}

	return true, nil
}
