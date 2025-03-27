package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/util"

	"github.com/volte6/gomud/internal/users"
)

/*
* Role Permissions:
* modify 				(All)
* modify.role			(Change user roles)
 */
func Modify(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 3 {
		infoOutput, _ := templates.Process("admincommands/help/command.modify", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	// modify permission <username> <newperms>
	if args[0] == `role` {

		if !user.HasRolePermission(`modify.role`) {
			user.SendText(`you do not have <ansi fg="command">modify.role</ansi> permission`)
			return true, nil
		}

		searchUser := args[1]
		newRole := args[2]

		allRoles := configs.GetRolesConfig()
		_, roleExists := allRoles[newRole]

		if newRole != users.RoleAdmin && newRole != users.RoleUser && !roleExists {
			user.SendText(`<ansi fg="alert-4">Invalid permission type.</ansi>`)
			return true, nil
		}

		foundUsername := ``
		foundCharacterName := ``

		for _, u := range users.GetAllActiveUsers() {
			if strings.EqualFold(searchUser, u.Username) {

				if u.Role == users.RoleAdmin {
					user.SendText(`<ansi fg="alert-4">Admin permissions cannot be removed this way.</ansi>`)
					return true, nil
				}

				if u.Role == newRole {
					user.SendText(`<ansi fg="alert-4">That permission is already set for this user.</ansi>`)
					return true, nil
				}

				foundCharacterName = u.Character.Name
				foundUsername = u.Username

				u.Role = newRole

				users.SaveUser(*u)

				u.SendText(`<ansi fg="alert-3">Your role has been set to: ` + newRole + `</ansi>`)
				break
			}
		}

		if len(foundUsername) == 0 {
			users.SearchOfflineUsers(func(u *users.UserRecord) bool {

				if strings.EqualFold(searchUser, u.Username) {

					if u.Role == users.RoleAdmin {
						user.SendText(`<ansi fg="alert-4">Admin permissions cannot be removed this way.</ansi>`)
						return false
					}

					if u.Role == newRole {
						user.SendText(`<ansi fg="alert-4">That permission is already set for this user.</ansi>`)
						return false
					}

					foundCharacterName = u.Character.Name
					foundUsername = u.Username

					u.Role = newRole

					users.SaveUser(*u)

					return false
				}

				return true
			})
		}

		if len(foundUsername) > 0 {
			user.SendText(fmt.Sprintf(`Role changed for user <ansi fg="username">%s</ansi> (Character name: <ansi fg="username">%s</ansi>).`, foundUsername, foundCharacterName))
			return true, nil
		}

	}

	return true, nil
}
