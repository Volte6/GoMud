package usercommands

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Jobs(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf(`user %d not found`, userId)
	}

	type JobDisplay struct {
		Name       string
		Experience string
		Completion string
		BarFull    string
		BarEmpty   string
	}

	jobProgress := []JobDisplay{}
	allRanks := user.Character.GetAllSkillRanks()

	for _, rank := range skills.GetProfessionRanks(allRanks) {

		barFull, barEmpty := util.ProgressBar(rank.Completion, 39)

		jobProgress = append(jobProgress, JobDisplay{
			Name:       rank.Profession,
			Experience: rank.ExperienceTitle,
			Completion: fmt.Sprintf(`%d%%`, int(math.Floor(rank.Completion*100))),
			BarFull:    barFull,
			BarEmpty:   barEmpty,
		})

	}

	// Sort lexigraphically
	sort.Slice(jobProgress, func(i, j int) bool {
		return strings.Compare(jobProgress[i].Name, jobProgress[j].Name) == -1
	})

	jobsTxt, _ := templates.Process("character/jobs", jobProgress)
	response.SendUserMessage(userId, jobsTxt, false)

	response.Handled = true
	return response, nil
}
