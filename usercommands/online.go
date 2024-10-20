package usercommands

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/volte6/gomud/configs"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/skills"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/users"
	"github.com/volte6/gomud/util"
)

func Online(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	headers := []string{`Name`, `Level`, `Alignment`, `Profession`, `Online`, `Role`}

	allFormatting := [][]string{}

	rows := [][]string{}

	rowsAdmin := [][]string{}
	rowsMod := [][]string{}
	rowsUser := [][]string{}

	c := configs.GetConfig()
	afkRounds := uint64(c.SecondsToRounds(int(c.AfkSeconds)))
	roundNow := util.GetRoundCount()

	userCt := 0
	for _, uid := range users.GetOnlineUserIds() {

		u := users.GetByUserId(uid)

		if u != nil {

			connTime := u.GetConnectTime()

			// subtract 3 hours
			//connTime = connTime.Add(-2 * time.Hour)
			//connTime = connTime.Add(-2 * time.Minute)

			oTime := time.Since(connTime)

			h := int(math.Floor(oTime.Hours()))
			m := int(math.Floor(oTime.Minutes())) - (h * 60)
			s := int(math.Floor(oTime.Seconds())) - (h * 60 * 60) - (m * 60)

			timeStr := ``
			if h > 0 {
				timeStr = fmt.Sprintf(`%dh%dm`, h, m)
			} else if m > 0 {
				timeStr = fmt.Sprintf(`%dm`, m)
			} else {
				timeStr = fmt.Sprintf(`%ds`, s)
			}

			if afkRounds > 0 && roundNow-u.GetLastInputRound() >= afkRounds {
				timeStr += ` <ansi fg="8">(afk)</ansi>`
			}

			userCt++
			row := []string{
				u.Character.Name,
				strconv.Itoa(u.Character.Level),
				u.Character.AlignmentName(),
				skills.GetProfession(u.Character.GetAllSkillRanks()),
				timeStr,
				u.Permission,
			}

			formatting := []string{
				`<ansi fg="username">%s</ansi>`,
				`<ansi fg="red">%s</ansi>`,
				`<ansi fg="` + u.Character.AlignmentName() + `">%s</ansi>`,
				`<ansi fg="white-bold">%s</ansi>`,
				`<ansi fg="magenta">%s</ansi>`,
				`<ansi fg="role-` + u.Permission + `-bold">%s</ansi>`,
			}

			allFormatting = append(allFormatting, formatting)

			if u.Permission == users.PermissionAdmin {
				rowsAdmin = append(rowsAdmin, row)
			} else if u.Permission == users.PermissionMod {
				rowsMod = append(rowsMod, row)
			} else {
				rowsUser = append(rowsUser, row)
			}

		}
	}

	rows = append(rows, rowsAdmin...)
	rows = append(rows, rowsMod...)
	rows = append(rows, rowsUser...)

	tableTitle := fmt.Sprintf(`%d users online`, userCt)
	if userCt == 1 {
		tableTitle = fmt.Sprintf(`%d user online`, userCt)
	}

	onlineResultsTable := templates.GetTable(tableTitle, headers, rows, allFormatting...)
	tplTxt, _ := templates.Process("tables/generic", onlineResultsTable)
	user.SendText(tplTxt)

	return true, nil
}
