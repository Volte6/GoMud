package usercommands

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Share(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	party := parties.Get(userId)
	if party == nil {
		response.SendUserMessage(userId, "You can only share in a party.", true)
		response.Handled = true
		return response, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 2 && strings.ToLower(args[1]) == "gold" {

		giveGoldAmount := 0

		if args[0] == "all" {
			giveGoldAmount = user.Character.Gold
		} else {
			giveGoldAmount, _ = strconv.Atoi(args[0])
		}

		if giveGoldAmount < 0 {
			response.SendUserMessage(userId, "You can't share a negative amount of gold.", true)
			response.Handled = true
			return response, nil
		}

		if giveGoldAmount > user.Character.Gold {
			response.SendUserMessage(userId, "You don't have that much gold to share.", true)
			response.Handled = true
			return response, nil
		}

		partyMembersInRoom := []int{userId} // make sure party leader gets first share
		for _, uid := range room.GetPlayers(rooms.FindAll) {
			if uid == userId {
				continue
			}
			if party.IsMember(uid) {
				partyMembersInRoom = append(partyMembersInRoom, uid)
			}
		}

		split := int(math.Floor(float64(giveGoldAmount) / float64(len(partyMembersInRoom))))
		leftOver := giveGoldAmount - split*len(partyMembersInRoom)

		for _, uid := range partyMembersInRoom {
			cmdQueue.QueueCommand(userId, 0, fmt.Sprintf("give %d gold to @%d", split, uid))
		}

		if leftOver > 0 {
			randomMember := partyMembersInRoom[util.Rand(len(partyMembersInRoom))]
			cmdQueue.QueueCommand(userId, 0, fmt.Sprintf("give %d gold to @%d", leftOver, randomMember))
		}

	} else {

		response.SendUserMessage(userId, `You can share gold by typing <ansi fg="command">share [amt] gold</ansi>?`, true)
	}

	response.Handled = true
	return response, nil
}
