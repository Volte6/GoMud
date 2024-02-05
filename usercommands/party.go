package usercommands

import (
	"fmt"
	"math"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Party(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	args := util.SplitButRespectQuotes(rest)

	partyCommand := `list`
	if len(args) > 0 {
		partyCommand = strings.ToLower(args[0])
		rest, _ = strings.CutPrefix(rest, args[0])
		rest = strings.TrimSpace(rest)
	}

	currentParty := parties.Get(userId)

	if partyCommand == `create` || partyCommand == `new` || partyCommand == `start` {

		// check if they are already part of a party
		if currentParty != nil {
			if currentParty.Invited(userId) {
				response.SendUserMessage(userId, `You already have a pending party invite. Try <ansi fg="command">party accept/decline</ansi> first`, true)
			} else if currentParty.IsLeader(userId) {
				response.SendUserMessage(userId, `You already own a party Type <ansi fg="command">party list</ansi> for more info.`, true)
			} else {
				response.SendUserMessage(userId, `You are already party of a party.`, true)
			}
			response.Handled = true
			return response, nil
		}

		if currentParty = parties.New(userId); currentParty != nil {
			response.SendUserMessage(userId, `You started a new party!`, true)

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				mob := mobs.GetInstance(instId)
				if mob == nil {
					continue
				}
				if mob.Character.IsCharmed(userId) {
					currentParty.AddMob(instId)
				}
			}

		} else {
			response.SendUserMessage(userId, `Something went wrong.`, true)
		}

		response.Handled = true
		return response, nil
	}
	// Done with create

	//
	// Everything after this point requires a party or an invitation to a party
	//

	if partyCommand == `invite` {

		if rest == `` {
			response.SendUserMessage(userId, `Invite who?`, true)
			response.Handled = true
			return response, nil
		}

		// Not in a party? Create one.
		if currentParty == nil {
			currentParty = parties.New(userId)
		}

		if !currentParty.IsLeader(userId) {
			response.SendUserMessage(userId, `You are not the leader of your party.`, true)
			response.Handled = true
			return response, nil
		}

		invitePlayerId, mobInstId := room.FindByName(rest)

		if invitePlayerId == 0 && mobInstId == 0 {
			response.SendUserMessage(userId, fmt.Sprintf(`%s not found.`, rest), true)
			response.Handled = true
			return response, nil
		}

		if invitedParty := parties.Get(invitePlayerId); invitedParty != nil {
			response.SendUserMessage(userId, `That player is already in a party.`, true)
			response.Handled = true
			return response, nil
		}

		invitedUser := users.GetByUserId(invitePlayerId)

		if invitedUser != nil && currentParty.InvitePlayer(invitePlayerId) {
			response.SendUserMessage(userId, fmt.Sprintf(`You invited <ansi fg="username">%s</ansi> to your party.`, invitedUser.Character.Name), true)
			response.SendUserMessage(invitePlayerId, fmt.Sprintf(`<ansi fg="username">%s</ansi> invited you to their party. Type <ansi fg="command">party accept</ansi> or <ansi fg="command">party decline</ansi> to respond.`, user.Character.Name), true)
		} else {
			response.SendUserMessage(userId, `Something went wrong.`, true)
		}

		if mobInstId != 0 {
			if mob := mobs.GetInstance(mobInstId); mob != nil {
				if mob.Character.IsCharmed() {
					for _, partyUId := range currentParty.UserIds {
						if mob.Character.IsCharmed(partyUId) {
							response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> joined your party.`, mob.Character.Name), true)
							currentParty.AddMob(mobInstId)

							response.Handled = true
							return response, nil
						}
					}
				}
				response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> doesn't want to join you.`, mob.Character.Name), true)
				response.Handled = true
				return response, nil
			}
		}

		response.Handled = true
		return response, nil
	}

	//
	// what follows doesn't mamke sense unless they are in a party
	//

	if currentParty == nil {
		response.SendUserMessage(userId, `You are not attached to a party.`, true)
		response.Handled = true
		return response, nil
	}

	if partyCommand == `accept` || partyCommand == `join` {

		if currentParty.AcceptInvite(userId) {

			response.SendUserMessage(userId, `You joined the party!`, true)
			for _, uid := range currentParty.UserIds {
				if uid == userId {
					continue
				}
				response.SendUserMessage(uid, fmt.Sprintf(`<ansi fg="username">%s</ansi> joined the party!`, user.Character.Name), true)
			}

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				mob := mobs.GetInstance(instId)
				if mob == nil {
					continue
				}
				if mob.Character.IsCharmed(userId) {
					currentParty.AddMob(instId)
				}
			}

		} else {
			response.SendUserMessage(userId, `Something went wrong.`, true)
		}
		response.Handled = true
		return response, nil
	}

	if partyCommand == `decline` {

		if currentParty.DeclineInvite(userId) {
			response.SendUserMessage(currentParty.LeaderUserId, fmt.Sprintf(`<ansi fg="username">%s</ansi> declined the invitation.`, user.Character.Name), true)
			response.SendUserMessage(userId, `You decline the invitation.`, true)
		} else {
			response.SendUserMessage(userId, `Something went wrong.`, true)
		}
		response.Handled = true
		return response, nil
	}

	if partyCommand == `list` {

		headers := []string{"Name", "Status", "Lvl", "Health", "%", "Location", "Position"}
		rows := [][]string{}

		if currentParty != nil {
			isInvited := currentParty.Invited(userId)
			leaderId := currentParty.LeaderUserId

			for _, uid := range currentParty.UserIds {
				uStatus := "In Party"
				if leaderId == uid {
					uStatus = "Leader"
				}

				u := users.GetByUserId(uid)
				uLevel := fmt.Sprintf(`%d`, u.Character.Level)
				uRoom := rooms.LoadRoom(u.Character.RoomId)
				uHealth := fmt.Sprintf(`%d/%d`, u.Character.Health, u.Character.HealthMax.Value)
				uHealthPct := fmt.Sprintf(`%d%%`, int(math.Floor((float64(u.Character.Health)/float64(u.Character.HealthMax.Value))*100)))
				uLoc := uRoom.Title
				rank := currentParty.GetRank(u.UserId)

				if isInvited {
					uLevel = `-`
					uHealth = `-`
					uLoc = `-`
					uHealthPct = `-`
					rank = `-`
				}

				rows = append(rows, []string{
					u.Character.Name,
					uStatus,
					uLevel,
					uHealth,
					uHealthPct,
					uLoc,
					rank,
				})
			}

			for _, mobInstanceId := range currentParty.GetMobs() {
				m := mobs.GetInstance(mobInstanceId)
				mRoom := rooms.LoadRoom(m.Character.RoomId)
				rows = append(rows, []string{
					m.Character.Name,
					`Friend`,
					fmt.Sprintf(`%d`, m.Character.Level),
					fmt.Sprintf(`%d/%d`, m.Character.Health, m.Character.HealthMax.Value),
					fmt.Sprintf(`%d%%`, int(math.Floor((float64(m.Character.Health)/float64(m.Character.HealthMax.Value))*100))),
					mRoom.Title,
					`-`,
				})
			}

			for _, uid := range currentParty.InviteUserIds {
				u := users.GetByUserId(uid)
				rows = append(rows, []string{
					u.Character.Name,
					`Invited`,
					`-`,
					`-`,
					`-`,
					`-`,
					`-`,
				})
			}

			tableFormatting := []string{`<ansi fg="username">%s</ansi>`,
				`<ansi fg="white-bold">%s</ansi>`,
				`<ansi fg="yellow">%s</ansi>`,
				`<ansi fg="cyan-bold">%s</ansi>`,
				`<ansi fg="black-bold">%s</ansi>`,
				`<ansi fg="magenta-bold">%s</ansi>`,
				`<ansi fg="white-bold">%s</ansi>`}

			partyTableData := templates.GetTable(`Party Members`, headers, rows, tableFormatting)
			partyTxt, _ := templates.Process("tables/generic", partyTableData)
			response.SendUserMessage(userId, partyTxt, true)

			if isInvited {
				response.SendUserMessage(userId, `Type <ansi fg="command">party accept/decline</ansi> to finalize your party membership.`, true)
			}
		}
	}

	if currentParty.Invited(userId) {
		response.SendUserMessage(userId, `You haven't accepted an invitation to the party.`, true)
		response.Handled = true
		return response, nil
	}

	//
	// Everything after this point you must be in a party
	//
	if partyCommand == `autoattack` {
		autoAttackOn := false
		if rest == `on` {
			autoAttackOn = true
		} else if rest == `off` {
			autoAttackOn = false
		} else {
			response.SendUserMessage(userId, `Usage: <ansi fg="command">party autoattack [on/off]</ansi>`, true)
			response.Handled = true
			return response, nil
		}

		wasOnBefore := currentParty.SetAutoAttack(userId, autoAttackOn)

		if autoAttackOn {
			if wasOnBefore {
				response.SendUserMessage(userId, `You already have auto-attack enabled.`, true)
			} else {
				response.SendUserMessage(userId, `You are now auto-attacking with your party.`, true)
			}
		} else {
			if wasOnBefore {
				response.SendUserMessage(userId, `You are no longer auto-attacking with your party.`, true)
			} else {
				response.SendUserMessage(userId, `You already have auto-attacking disabled.`, true)
			}
		}
	}

	if partyCommand == `leave` || partyCommand == `quit` {

		if currentParty.IsLeader(userId) {

			if len(currentParty.UserIds) <= 1 {
				response.SendUserMessage(userId, `You disbanded the party.`, true)
				currentParty.Disband()
				response.Handled = true
				return response, nil
			}

			// promote someone else to leader
			for _, uid := range currentParty.UserIds {
				if uid == userId {
					continue
				}
				currentParty.LeaderUserId = uid
				newLeaderUser := users.GetByUserId(uid)
				response.SendUserMessage(uid, fmt.Sprintf(`<ansi fg="username">%s</ansi> is now the leader of the party.`, newLeaderUser.Character.Name), true)
				break
			}
		}

		currentParty.Leave(userId)

		response.SendUserMessage(userId, `You left the party.`, true)

		for _, uid := range currentParty.UserIds {
			if uid == userId {
				continue
			}
			response.SendUserMessage(uid, fmt.Sprintf(`<ansi fg="username">%s</ansi> left the party.`, user.Character.Name), true)
		}

	}

	if partyCommand == `disband` || partyCommand == `stop` {

		if !currentParty.IsLeader(userId) {
			response.SendUserMessage(userId, `You are not the leader of your party.`, true)
			response.Handled = true
			return response, nil
		}

		for _, uid := range currentParty.UserIds {
			if uid == userId {
				continue
			}
			response.SendUserMessage(uid, fmt.Sprintf(`<ansi fg="username">%s</ansi> disbanded the party.`, user.Character.Name), true)
		}
		for _, uid := range currentParty.InviteUserIds {
			response.SendUserMessage(uid, fmt.Sprintf(`<ansi fg="username">%s</ansi> disbanded the party.`, user.Character.Name), true)
		}

		currentParty.Disband()

		response.SendUserMessage(userId, `You disbanded the party.`, true)

		response.Handled = true
		return response, nil
	}

	if partyCommand == `kick` {

		if !currentParty.IsLeader(userId) {
			response.SendUserMessage(userId, `You are not the leader of your party.`, true)
			response.Handled = true
			return response, nil
		}

		allMembers := []string{}
		memberIds := map[string]int{}
		for _, uid := range currentParty.GetMembers() {
			u := users.GetByUserId(uid)
			if u == nil {
				continue
			}
			allMembers = append(allMembers, u.Character.Name)
			memberIds[u.Character.Name] = uid
		}

		matchUser, closeMatchUser := util.FindMatchIn(rest, allMembers...)
		if matchUser == `` {
			matchUser = closeMatchUser
		}

		if matchUser == `` {
			response.SendUserMessage(userId, fmt.Sprintf(`%s not found.`, rest), true)
			response.Handled = true
			return response, nil
		}

		kickUserId := memberIds[matchUser]

		currentParty.Leave(kickUserId)

		response.SendUserMessage(kickUserId, `You were kicked from the party.`, true)

		for _, uid := range currentParty.UserIds {
			response.SendUserMessage(uid, fmt.Sprintf(`<ansi fg="username">%s</ansi> was kicked from the party.`, matchUser), true)
		}
	}

	if partyCommand == `promote` {

		if !currentParty.IsLeader(userId) {
			response.SendUserMessage(userId, `You are not the leader of your party.`, true)
			response.Handled = true
			return response, nil
		}

		allMembers := []string{}
		memberIds := map[string]int{}
		for _, uid := range currentParty.GetMembers() {
			u := users.GetByUserId(uid)
			if u == nil {
				continue
			}
			allMembers = append(allMembers, u.Character.Name)
			memberIds[u.Character.Name] = uid
		}

		matchUser, closeMatchUser := util.FindMatchIn(rest, allMembers...)
		if matchUser == `` {
			matchUser = closeMatchUser
		}

		if matchUser == `` {
			response.SendUserMessage(userId, fmt.Sprintf(`%s not found.`, rest), true)
			response.Handled = true
			return response, nil
		}

		promoteUserId := memberIds[matchUser]

		currentParty.LeaderUserId = promoteUserId

		response.SendUserMessage(promoteUserId, `You have been promoted to party leader.`, true)

		for _, uid := range currentParty.UserIds {
			if uid != promoteUserId {
				response.SendUserMessage(uid, fmt.Sprintf(`<ansi fg="username">%s</ansi> is now the party leader.`, matchUser), true)
			}
		}

	}

	if partyCommand == `chat` || partyCommand == `say` {

		if len(rest) == 0 {
			response.SendUserMessage(userId, `What do you want to say?`, true)
			response.Handled = true
			return response, nil
		}

		for _, uId := range currentParty.GetMembers() {
			if uId == userId {
				continue
			}
			response.SendUserMessage(uId, fmt.Sprintf(`<ansi fg="magenta">(party)</ansi> <ansi fg="username">%s</ansi> says, "<ansi fg="yellow">%s</ansi>`, user.Character.Name, rest), true)
		}

		response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="magenta">(party)</ansi> You say, "<ansi fg="yellow">%s</ansi>"`, rest), true)
	}

	response.Handled = true
	return response, nil
}
