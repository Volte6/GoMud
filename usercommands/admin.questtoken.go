package usercommands

import (
	"fmt"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/quests"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func QuestToken(rest string, userId int) (util.MessageQueue, error) {

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
		user.SendText(infoOutput)
	} else if args[0] == "list" {

		allTokens := user.Character.GetQuestProgress()
		headers := []string{"Quest Name", "Token/Steps"}
		rows := [][]string{}

		if len(allTokens) == 0 {
			rows = append(rows, []string{"None", "None"})
		} else {
			for qid, qt := range allTokens {
				qTokenStr := ``
				qToken := fmt.Sprintf(`%d-%s`, qid, qt)
				qInfo := quests.GetQuest(qToken)
				for _, step := range qInfo.Steps {
					if step.Id == qt {
						qTokenStr += fmt.Sprintf(`[%d-%s] `, qid, step.Id)
					} else {
						qTokenStr += fmt.Sprintf(`%d-%s `, qid, step.Id)
					}
				}
				rows = append(rows, []string{qInfo.Name, qTokenStr})
			}
		}

		searchResultsTable := templates.GetTable("Quest Tokens", headers, rows)
		tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
		user.SendText(tplTxt)

	} else if args[0] == "all" {

		allQuests := quests.GetAllQuests()
		headers := []string{"Quest Name", "Token/Steps"}
		rows := [][]string{}

		if len(allQuests) == 0 {
			rows = append(rows, []string{"None", "None"})
		} else {
			for _, qInfo := range allQuests {
				qTokenStr := ``
				for _, step := range qInfo.Steps {
					qTokenStr += fmt.Sprintf(`%d-%s `, qInfo.QuestId, step.Id)
				}
				rows = append(rows, []string{qInfo.Name, qTokenStr})
			}
		}

		searchResultsTable := templates.GetTable("Quest Tokens", headers, rows)
		tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
		user.SendText(tplTxt)

	} else {

		events.AddToQueue(events.Quest{
			UserId:     userId,
			QuestToken: args[0],
		})

	}

	response.Handled = true
	return response, nil
}
