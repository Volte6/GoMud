package usercommands

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Online(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	headers := []string{`Name`, `Level`, `Alignment`, `Profession`, `Online`, `Role`}

	allFormatting := [][]string{}

	rows := [][]string{}

	rowsAdmin := [][]string{}
	rowsMod := [][]string{}
	rowsUser := [][]string{}

	userCt := 0
	for _, uid := range users.GetOnlineUserIds() {

		user := users.GetByUserId(uid)

		if user != nil {

			connTime := user.GetConnectTime()

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

			userCt++
			row := []string{
				user.Character.Name,
				strconv.Itoa(user.Character.Level),
				user.Character.AlignmentName(),
				skills.GetProfession(user.Character.GetAllSkillRanks()),
				timeStr,
				user.Permission,
			}

			formatting := []string{
				`<ansi fg="username">%s</ansi>`,
				`<ansi fg="red">%s</ansi>`,
				`<ansi fg="` + user.Character.AlignmentName() + `">%s</ansi>`,
				`<ansi fg="white" bold="true">%s</ansi>`,
				`<ansi fg="magenta">%s</ansi>`,
				`<ansi fg="role-` + user.Permission + `" bold="true">%s</ansi>`,
			}

			allFormatting = append(allFormatting, formatting)

			if user.Permission == users.PermissionAdmin {
				rowsAdmin = append(rowsAdmin, row)
			} else if user.Permission == users.PermissionMod {
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
	response.SendUserMessage(userId, tplTxt, false)

	response.Handled = true
	return response, nil
}
