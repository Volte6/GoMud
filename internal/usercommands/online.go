package usercommands

import (
	"fmt"
	"strconv"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Online(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	headers := []string{`Name`, `Level`, `Alignment`, `Profession`, `Online`, `Role`}

	allFormatting := [][]string{}

	rows := [][]string{}

	rowsAdmin := [][]string{}
	rowsMod := [][]string{}
	rowsUser := [][]string{}

	userCt := 0
	for _, uid := range users.GetOnlineUserIds() {

		u := users.GetByUserId(uid)

		if u != nil {

			onlineInfo := u.GetOnlineInfo()

			userCt++

			onlineTime := onlineInfo.OnlineTimeStr
			if onlineInfo.IsAFK {
				onlineTime += ` <ansi fg="8">(afk)</ansi>`
			}

			row := []string{
				onlineInfo.CharacterName,
				strconv.Itoa(onlineInfo.Level),
				onlineInfo.Alignment,
				onlineInfo.Profession,
				onlineTime,
				onlineInfo.Permission,
			}

			formatting := []string{
				`<ansi fg="username">%s</ansi>`,
				`<ansi fg="red">%s</ansi>`,
				`<ansi fg="` + onlineInfo.Alignment + `">%s</ansi>`,
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
