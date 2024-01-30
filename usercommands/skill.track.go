package usercommands

import (
	"errors"
	"fmt"
	"log/slog"
	"math"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Track(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Track)

	if skillLevel == 0 {
		response.SendUserMessage(userId, "You don't know how to track.", true)
		response.Handled = true
		return response, errors.New(`you don't know how to track`)
	}

	trackDistance := skillLevel*2 + int(math.Ceil(float64(user.Character.Stats.Perception.Value)/10)) // Keep the number even

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	if rest == "" {

		if skillLevel < 4 {
			response.SendUserMessage(userId, "Track who?", true)
			response.Handled = true
			return response, fmt.Errorf("track who?")
		}

		if !user.Character.TryCooldown(skills.Track.String(), 2) {
			response.SendUserMessage(userId,
				fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Track.String())),
				true)
			response.Handled = true
			return response, errors.New(`you're doing that too often`)
		}

		type TrackInfo struct {
			Name     string
			Strength string
		}

		visitorData := make([]TrackInfo, 0)

		visitors := room.Visitors()
		for visitorUserId, trailStrength := range visitors {
			if visitorUserId == userId {
				continue
			}
			strengthStr := ""
			strength := int(math.Round(trailStrength * 100))
			if strength < 10 {
				strengthStr = "Dead"
			} else if strength < 25 {
				strengthStr = "Weak"
			} else if strength < 50 {
				strengthStr = "Good"
			} else if strength < 75 {
				strengthStr = "Warm"
			} else {
				strengthStr = "Hot"
			}

			visitorData = append(visitorData, TrackInfo{
				Name:     users.GetByUserId(visitorUserId).Character.Name,
				Strength: strengthStr,
			})
		}

		trackTxt, _ := templates.Process("descriptions/track", visitorData)
		response.SendUserMessage(userId, trackTxt, false)

		response.Handled = true
		return response, nil
	}

	if !user.Character.TryCooldown(skills.Track.String(), 2) {
		response.SendUserMessage(userId,
			fmt.Sprintf("You need to wait %d more rounds to use that skill again.", user.Character.GetCooldown(skills.Track.String())),
			true)
		response.Handled = true
		return response, errors.New(`you're doing that too often`)
	}

	trackedUser := users.GetByCharacterName(rest)
	if trackedUser == nil {
		response.SendUserMessage(userId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> isn't around.`, rest),
			true)
		response.Handled = true
		return response, nil
	}

	if trackedUser.Character.RoomId == user.Character.RoomId {
		response.SendUserMessage(userId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> is in the room with you!`, trackedUser.Character.Name),
			true)
		response.Handled = true
		return response, nil
	}

	rGraph := rooms.GenerateZoneMap(user.Character.Zone, user.Character.RoomId, trackedUser.UserId, trackDistance, trackDistance, rooms.MapModeTracking)

	if rGraph.RoomCount() <= 1 {
		response.SendUserMessage(userId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> isn't around.`, rest),
			true)
		response.Handled = true
		return response, nil
	}

	mapData, err := rooms.DrawZoneMap(rGraph, "Tracking "+trackedUser.Character.Name, 65, 18)

	if err != nil {
		return response, err
	}

	mapTxt, err := templates.Process("maps/map", mapData)
	if err != nil {
		slog.Error("Map", "error", err.Error())
		response.SendUserMessage(userId, fmt.Sprintf(`No map found for "%s"`, rest), true)
		response.Handled = true
		return response, err
	}

	response.SendUserMessage(userId, mapTxt, false)

	response.Handled = true
	return response, nil
}
