package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/util"

	"github.com/volte6/gomud/internal/users"
)

func Modify(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if user.Permission != users.PermissionAdmin {
		user.SendText(`<ansi fg="alert-4">Only admins can use this command</ansi>`)
		return true, nil
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 3 {
		infoOutput, _ := templates.Process("admincommands/help/command.modify", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	// modify permission <username> <newperms>
	if args[0] == `permission` {

		searchUser := args[1]
		newPerms := args[2]

		if newPerms != users.PermissionUser && newPerms != users.PermissionMod && newPerms != users.PermissionAdmin {
			user.SendText(`<ansi fg="alert-4">Invalid permission type.</ansi>`)
			return true, nil
		}

		foundUsername := ``
		foundCharacterName := ``

		for _, u := range users.GetAllActiveUsers() {
			if strings.EqualFold(searchUser, u.Username) {
				if u.Permission == users.PermissionAdmin {
					user.SendText(`<ansi fg="alert-4">Admin permissions cannot be removed this way.</ansi>`)
					return true, nil
				}

				if u.Permission == newPerms {
					user.SendText(`<ansi fg="alert-4">That permission is already set for this user.</ansi>`)
					return true, nil
				}

				foundCharacterName = u.Character.Name
				foundUsername = u.Username

				u.Permission = newPerms

				users.SaveUser(*u)

				u.SendText(`<ansi fg="alert-3">Your permission has been set to: ` + newPerms + `</ansi>`)
				break
			}
		}

		if len(foundUsername) == 0 {
			users.SearchOfflineUsers(func(u *users.UserRecord) bool {

				if strings.EqualFold(searchUser, u.Username) {

					if u.Permission == users.PermissionAdmin {
						user.SendText(`<ansi fg="alert-4">Admin permissions cannot be removed this way.</ansi>`)
						return false
					}

					if u.Permission == newPerms {
						user.SendText(`<ansi fg="alert-4">That permission is already set for this user.</ansi>`)
						return false
					}

					foundCharacterName = u.Character.Name
					foundUsername = u.Username

					u.Permission = newPerms

					users.SaveUser(*u)

					return false
				}

				return true
			})
		}

		if len(foundUsername) > 0 {
			user.SendText(fmt.Sprintf(`Permissions changed for user <ansi fg="username">%s</ansi> (Character name: <ansi fg="username">%s</ansi>).`, foundUsername, foundCharacterName))
			return true, nil
		}

	}

	return true, nil
}
