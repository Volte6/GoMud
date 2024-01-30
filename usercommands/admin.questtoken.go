package usercommands

import (
	"fmt"

	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func QuestToken(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// args should look like one of the following:
	// questtoken <tokenname>
	// questtoken list
	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.questtoken", nil)
		response.SendUserMessage(userId, infoOutput, false)
	} else if args[0] == "list" {

		allTokens := user.Character.GetQuestProgress()
		headers := []string{"Token Name"}
		rows := [][]string{}

		if len(allTokens) == 0 {
			rows = append(rows, []string{"None"})
		} else {
			for _, qt := range allTokens {
				rows = append(rows, []string{qt})
			}
		}

		searchResultsTable := templates.GetTable("Quest Tokens", headers, rows)
		tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
		response.SendUserMessage(userId, tplTxt, false)

	} else {

		cmdQueue.QueueQuest(userId, args[0])

	}

	response.Handled = true
	return response, nil
}
