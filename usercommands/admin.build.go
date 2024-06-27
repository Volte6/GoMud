package usercommands

import (
	"fmt"
	"math"
	"strings"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func IBuild(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 1 {

		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.build", nil)
		response.SendUserMessage(userId, infoOutput, false)
	}

	cmdPrompt, _ := user.StartPrompt(`ibuild`, rest)

	// #build zone "The Arctic"
	if args[0] == "zone" {

		zoneQ := cmdPrompt.Ask(`New zone name?`, []string{``})
		if !zoneQ.Done {
			response.Handled = true
			return response, nil
		}

		if zoneQ.Response == `` {
			response.SendUserMessage(userId, `Aborting zone build`, true)
			user.ClearPrompt()
			response.Handled = true
			return response, nil
		}

		zoneName := zoneQ.Response

		if roomId, err := rooms.CreateZone(zoneName); err != nil {
			response.SendUserMessage(userId, err.Error(), true)
		} else {
			response.SendUserMessage(userId, fmt.Sprintf(`Zone %s created.`, zoneName), true)

			if err := rooms.MoveToRoom(user.UserId, roomId); err != nil {
				response.SendUserMessage(userId, err.Error(), true)
			} else {
				response.SendUserMessage(userId, fmt.Sprintf(`Moved to room %d.`, roomId), true)
				response.NextCommand = `look`
			}
		}

		user.ClearPrompt()
		response.Handled = true
		return response, nil

	}

	// #build room north <south>
	if args[0] == "room" {

		exitNameQ := cmdPrompt.Ask(`Room exit name?`, []string{})
		if !exitNameQ.Done {
			response.Handled = true
			return response, nil
		}

		if exitNameQ.Response == `` {
			response.SendUserMessage(userId, `Aborting room build`, true)
			user.ClearPrompt()
			response.Handled = true
			return response, nil
		}

		exitName := exitNameQ.Response

		dirNameQ := cmdPrompt.Ask(`Map direction?`, []string{})
		if !dirNameQ.Done {
			response.Handled = true
			return response, nil
		}

		if _, ok := rooms.DirectionDeltas[dirNameQ.Response]; !ok {
			dirNameQ.RejectResponse()
			response.SendUserMessage(userId, `Invalid map direction.`, true)
			response.Handled = true
			return response, nil

		}

		mapDirection := dirNameQ.Response

		retDirNameQ := cmdPrompt.Ask(`Return exit name (opt)?`, []string{}, ``)
		if !retDirNameQ.Done {
			response.Handled = true
			return response, nil
		}

		returnName := retDirNameQ.Response

		response.SendUserMessage(userId, fmt.Sprintf(`exitName: %s - mapDirection: %s - returnName: %s.`, exitName, mapDirection, returnName), true)

		user.ClearPrompt()
		response.Handled = true
		return response, nil

	}

	// TODO: WIP

	response.Handled = true
	return response, nil
}

func Build(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.build", nil)
		response.SendUserMessage(userId, infoOutput, false)
	} else {

		// #build zone "The Arctic"
		if args[0] == "zone" {

			zoneName := strings.Join(args[1:], ` `)

			if roomId, err := rooms.CreateZone(zoneName); err != nil {
				response.SendUserMessage(userId, err.Error(), true)
			} else {
				response.SendUserMessage(userId, fmt.Sprintf("Zone %s created.", zoneName), true)

				if err := rooms.MoveToRoom(user.UserId, roomId); err != nil {
					response.SendUserMessage(userId, err.Error(), true)
				} else {
					response.SendUserMessage(userId, fmt.Sprintf("Moved to room %d.", roomId), true)
					response.NextCommand = "look"
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
					response.SendUserMessage(userId, err.Error(), true)
					response.Handled = true
					return response, err
				}

				map2D, cX, cY := rGraph.Generate2DMap(11, 11, user.Character.RoomId)

				if len(map2D) < 1 {
					response.SendUserMessage(userId, "Error generating a 2d map", true)
					response.Handled = true
					return response, nil
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
							response.SendUserMessage(userId, fmt.Sprintf("Exiting room found at the %s direction. Connecting them.", exitName), true)
							rooms.ConnectRoom(user.Character.RoomId, destinationRoom.RoomId, exitName, mapDirection) // north/north-x2
						}
					}
				}

			}

			// Only build a new room if we don't already have a destination room from the above code tryin gto find/connect
			if destinationRoom == nil {
				if newRoom, err := rooms.BuildRoom(user.Character.RoomId, exitName, mapDirection); err != nil {
					response.SendUserMessage(userId, err.Error(), true)
				} else {
					destinationRoom = newRoom
				}

				if destinationRoom == nil {
					response.SendUserMessage(userId, fmt.Sprintf("Error building room %s.", exitName), true)
					return response, nil
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
				response.SendUserMessage(userId, err.Error(), true)
			} else {
				response.SendUserMessage(userId, fmt.Sprintf("Moved to room %d.", destinationRoom.RoomId), true)
				response.NextCommand = "look"
			}

		}

	}

	response.Handled = true
	return response, nil
}
