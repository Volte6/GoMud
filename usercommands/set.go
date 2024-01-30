package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Set(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		response.SendUserMessage(userId, "Set what?", true)
		response.Handled = true
		return response, nil
	}

	if args[0] == "description" {

		rest = strings.TrimSpace(rest[len(args[0]):])
		if len(rest) > 1024 {
			rest = rest[:1024]
		}
		user.Character.Description = rest

		response.SendUserMessage(userId, "Description set. Look at yourself to confirm.", true)
		response.Handled = true
		return response, nil
	}

	// Are they setting a macro?
	if len(args[0]) == 2 && args[0][0] == '=' {
		macroNum, _ := strconv.Atoi(string(args[0][1]))
		if macroNum == 0 {
			response.SendUserMessage(userId, "Invalid macro number supplied.", true)
			response.Handled = true
			return response, nil
		}
		if user.Macros == nil {
			user.Macros = make(map[string]string)
		}
		rest = strings.TrimSpace(rest[2:])

		if len(rest) > 128 {
			rest = rest[:128]
		}

		if len(rest) == 0 {
			delete(user.Macros, args[0])

			response.SendUserMessage(userId,
				fmt.Sprintf(`Macro <ansi fg="command">=%d</ansi> deleted.`, macroNum),
				true)
		} else {

			for _, cmd := range strings.Split(rest, ";") {
				if len(cmd) > 0 {
					if cmd[0] == '=' {
						response.SendUserMessage(userId,
							`You cannot reference macros inside of a macro`,
							true)
						response.Handled = true
						return response, nil
					}
				}
			}

			user.Macros[args[0]] = rest

			response.SendUserMessage(userId,
				fmt.Sprintf(`Macro set. Type <ansi fg="command">=%d</ansi> or press <ansi fg="command">F%d</ansi> to use it.`, macroNum, macroNum),
				true)
		}

		response.Handled = true
		return response, nil
	}
	// Setting macros:
	// set =1 "say hello"

	response.Handled = true
	return response, nil
}
