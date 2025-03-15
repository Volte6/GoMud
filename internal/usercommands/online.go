package usercommands

import (
	"fmt"
	"strconv"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Online(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	headers := []string{
		templates.T(`User.Name`),
		templates.T(`Level`),
		templates.T(`Alignment`),
		templates.T(`Profession`),
		templates.T(`Online`),
		templates.T(`Role`),
	}

	if user.Permission != users.PermissionUser {
		headers = append([]string{templates.T(`UserId`)}, headers...)
		headers = append(headers, []string{templates.T(`Zone`), templates.T(`RoomId`)}...)
	}

	allFormatting := [][]string{}

	rows := [][]string{}

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

			if user.Permission != users.PermissionUser {
				row = append([]string{strconv.Itoa(u.UserId)}, row...)
				row = append(row, []string{u.Character.Zone, strconv.Itoa(u.Character.RoomId)}...)

				formatting = append([]string{`<ansi fg="userid">%s</ansi>`}, formatting...)
				formatting = append(formatting, []string{`<ansi fg="zone">%s</ansi>`, `<ansi fg="1">%s</ansi>`}...)
			}

			allFormatting = append(allFormatting, formatting)

			rows = append(rows, row)
		}
	}

	tableTitle := fmt.Sprintf(templates.T(`%d users online`), userCt)
	if userCt == 1 {
		tableTitle = fmt.Sprintf(templates.T(`%d user online`), userCt)
	}

	onlineResultsTable := templates.GetTable(tableTitle, headers, rows, allFormatting...)
	tplTxt, _ := templates.Process("tables/generic", onlineResultsTable)
	user.SendText(tplTxt)

	return true, nil
}
