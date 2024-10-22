package usercommands

import (
	"strings"

	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/util"

	"github.com/volte6/gomud/users"
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

		for _, u := range users.GetAllActiveUsers() {
			if strings.EqualFold(searchUser, u.Username) {
				if u.Permission == users.PermissionAdmin {
					user.SendText(`<ansi fg="alert-4">Admin permissions cannot be removed this way.</ansi>`)
					return true, nil
				}

				u.Permission = newPerms

				users.SaveUser(*u)

				u.SendText(`<ansi fg="alert-3">Your permission has been set to: ` + newPerms + `</ansi>`)
				user.SendText("Permissions changed.")

				return true, nil
			}
		}

		found := false
		users.SearchOfflineUsers(func(u *users.UserRecord) bool {

			if strings.EqualFold(searchUser, u.Username) {

				if u.Permission == users.PermissionAdmin {
					user.SendText(`<ansi fg="alert-4">Admin permissions cannot be removed this way.</ansi>`)
					return false
				}

				found = true

				u.Permission = newPerms

				users.SaveUser(*u)

				return false
			}

			return true
		})

		if found {
			user.SendText("Permissions changed.")
			return true, nil
		}

		user.SendText("Could not find user.")
	}

	return true, nil
}
