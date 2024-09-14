package usercommands

import (
	"fmt"
	"math"
	"strings"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func IBuild(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 1 {

		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.build", nil)
		user.SendText(infoOutput)
	}

	cmdPrompt, _ := user.StartPrompt(`ibuild`, rest)

	// #build zone "The Arctic"
	if args[0] == "zone" {

		zoneQ := cmdPrompt.Ask(`New zone name?`, []string{``})
		if !zoneQ.Done {
			return true, nil
		}

		if zoneQ.Response == `` {
			user.SendText(`Aborting zone build`)
			user.ClearPrompt()
			return true, nil
		}

		zoneName := zoneQ.Response
		if roomId, err := rooms.CreateZone(zoneName); err != nil {
			user.SendText(err.Error())
		} else {
			user.SendText(fmt.Sprintf(`Zone %s created.`, zoneName))

			if err := rooms.MoveToRoom(user.UserId, roomId); err != nil {
				user.SendText(err.Error())
			} else {
				user.SendText(fmt.Sprintf(`Moved to room %d.`, roomId))

				events.AddToQueue(events.Input{
					UserId:    userId,
					InputText: `look`,
				}, true)

			}
		}

		user.ClearPrompt()
		return true, nil

	}

	// #build room north <south>
	if args[0] == "room" {

		exitNameQ := cmdPrompt.Ask(`Room exit name?`, []string{})
		if !exitNameQ.Done {
			return true, nil
		}

		if exitNameQ.Response == `` {
			user.SendText(`Aborting room build`)
			user.ClearPrompt()
			return true, nil
		}

		exitName := exitNameQ.Response

		dirNameQ := cmdPrompt.Ask(`Map direction?`, []string{})
		if !dirNameQ.Done {
			return true, nil
		}

		if _, ok := rooms.DirectionDeltas[dirNameQ.Response]; !ok {
			dirNameQ.RejectResponse()
			user.SendText(`Invalid map direction.`)
			return true, nil

		}

		mapDirection := dirNameQ.Response

		retDirNameQ := cmdPrompt.Ask(`Return exit name (opt)?`, []string{}, ``)
		if !retDirNameQ.Done {
			return true, nil
		}

		returnName := retDirNameQ.Response

		user.SendText(fmt.Sprintf(`exitName: %s - mapDirection: %s - returnName: %s.`, exitName, mapDirection, returnName))

		user.ClearPrompt()
		return true, nil

	}

	// TODO: WIP

	return true, nil
}

func Build(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.build", nil)
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
						UserId:    userId,
						InputText: `look`,
					}, true)
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

			deltaD, ok := rooms.DirectionDeltas[exitName]
			if ok {
				rGraph := rooms.NewRoomGraph(100, 100, 0, rooms.MapModeAll)
				err := rGraph.Build(user.Character.RoomId, nil)
				if err != nil {
					user.SendText(err.Error())
					return true, err
				}

				map2D, cX, cY := rGraph.Generate2DMap(11, 11, user.Character.RoomId)

				if len(map2D) < 1 {
					user.SendText("Error generating a 2d map")
					return true, nil
				}

				// extra large exits get translated to their correct exit name, and the "mapdirection" updated to the specified one
				if math.Abs(float64(deltaD.Dy)) > 1 || math.Abs(float64(deltaD.Dx)) > 1 {
					if strings.Contains(exitName, `-`) {
						mapDirection = exitName // mapDirection will be "north-x2" for example
						parts := strings.Split(exitName, `-`)
						exitName = parts[0] // exitname will be "north" for example
					}
				}

				if cY+deltaD.Dy < len(map2D) && cX+deltaD.Dx < len(map2D[0]) {
					if cY+deltaD.Dy >= 0 && cX+deltaD.Dx >= 0 {
						if map2D[cY+deltaD.Dy][cX+deltaD.Dx] != nil {
							destinationRoom = rooms.LoadRoom(map2D[cY+deltaD.Dy][cX+deltaD.Dx].RoomId)
							user.SendText(fmt.Sprintf("Exiting room found at the %s direction. Connecting them.", exitName))
							rooms.ConnectRoom(user.Character.RoomId, destinationRoom.RoomId, exitName, mapDirection) // north/north-x2
						}
					}
				}

			}

			// Only build a new room if we don't already have a destination room from the above code tryin gto find/connect
			if destinationRoom == nil {
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
					UserId:    userId,
					InputText: `look`,
				}, true)
			}

		}

	}

	return true, nil
}
