package usercommands

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/parties"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Share(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	party := parties.Get(user.UserId)
	if party == nil {
		user.SendText("You can only share in a party.")
		return true, nil
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
			user.SendText("You can't share a negative amount of gold.")
			return true, nil
		}

		if giveGoldAmount > user.Character.Gold {
			user.SendText("You don't have that much gold to share.")
			return true, nil
		}

		partyMembersInRoom := []int{user.UserId} // make sure party leader gets first share
		for _, uid := range room.GetPlayers(rooms.FindAll) {
			if uid == user.UserId {
				continue
			}
			if party.IsMember(uid) {
				partyMembersInRoom = append(partyMembersInRoom, uid)
			}
		}

		split := int(math.Floor(float64(giveGoldAmount) / float64(len(partyMembersInRoom))))
		leftOver := giveGoldAmount - split*len(partyMembersInRoom)

		for _, uid := range partyMembersInRoom {

			user.Command(fmt.Sprintf("give %d gold to @%d", split, uid))

		}

		if leftOver > 0 {

			randomMember := partyMembersInRoom[util.Rand(len(partyMembersInRoom))]

			user.Command(fmt.Sprintf("give %d gold to @%d", leftOver, randomMember))

		}

	} else {

		user.SendText(`You can share gold by typing <ansi fg="command">share [amt] gold</ansi>?`)
	}

	return true, nil
}
