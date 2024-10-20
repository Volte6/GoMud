package usercommands

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/skills"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/users"
	"github.com/volte6/gomud/util"
)

func Jobs(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

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
	user.SendText(jobsTxt)

	return true, nil
}
