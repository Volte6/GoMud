package usercommands

import (
	"fmt"
	"math"
	"sort"

	"github.com/volte6/mud/quests"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Quests(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf(`user %d not found`, userId)
	}

	type QuestRecord struct {
		Id          int
		Name        string
		Description string
		Completion  string
		BarFull     string
		BarEmpty    string
	}

	type QuestInfo struct {
		QuestsTotal int
		QuestsFound int
		Records     []QuestRecord
	}

	showHidden := rest == `all+`
	showComplete := (rest == `all`) || showHidden

	qInfo := QuestInfo{}
	allQuests := []QuestRecord{}
	var completion float64

	qInfo.QuestsTotal = 0

	allQuestProgress := user.Character.GetQuestProgress()

	if rest == `all+` {
		for _, quest := range quests.GetAllQuests() {
			if _, ok := allQuestProgress[quest.QuestId]; ok {
				continue
			}
			allQuestProgress[quest.QuestId] = `all+`
		}
	} else {

	}

	for questId, questStep := range allQuestProgress {
		questToken := quests.PartsToToken(questId, questStep)
		if questInfo := quests.GetQuest(questToken); questInfo != nil {

			// Secret quests are not rendered
			if !showHidden && questInfo.Secret {
				continue
			}

			qInfo.QuestsTotal++

			totalSteps := len(questInfo.Steps)
			completedSteps := 0
			description := questInfo.Description

			if questStep != `all+` {
				for _, step := range questInfo.Steps {
					completedSteps++
					if step.Id == questStep {
						description = step.Description
						break
					}
				}
			}

			completion = float64(completedSteps) / float64(totalSteps)

			if !showComplete && completion >= 1 {
				continue
			}

			barFull, barEmpty := util.ProgressBar(completion, 25)

			qDisplay := QuestRecord{
				Id:          questInfo.QuestId,
				Name:        questInfo.Name,
				Description: description,
				Completion:  fmt.Sprintf(`%d%%`, int(math.Floor(completion*100))),
				BarFull:     barFull,
				BarEmpty:    barEmpty,
			}

			allQuests = append(allQuests, qDisplay)
		}
	}

	qInfo.QuestsFound = len(allQuests)
	qInfo.Records = allQuests

	// Sort lexigraphically
	sort.Slice(allQuests, func(i, j int) bool {
		return allQuests[i].Id < allQuests[j].Id
	})

	questTxt, _ := templates.Process("character/quests", qInfo)
	response.SendUserMessage(userId, questTxt, false)

	response.Handled = true
	return response, nil
}
