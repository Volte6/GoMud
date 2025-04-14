package usercommands

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Jobs(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

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

	jobsTxt, _ := templates.Process("character/jobs", jobProgress, user.UserId)
	user.SendText(jobsTxt)

	return true, nil
}
