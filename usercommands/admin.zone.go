package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/users"
	"github.com/volte6/gomud/util"
)

func Zone(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	handled := true

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.zone", nil)
		user.SendText(infoOutput)

		return handled, nil
	}

	roomCmd := strings.ToLower(args[0])
	args = args[1:]

	zoneConfig := rooms.GetZoneConfig(room.Zone)
	if zoneConfig == nil {
		user.SendText(fmt.Sprintf(`Couldn't find zone info for <ansi fg="red">%s</ansi>`, room.Zone))
		return true, nil
	}

	if roomCmd == `info` {

		user.SendText(``)
		user.SendText(fmt.Sprintf(`<ansi fg="yellow-bold">Zone Config for <ansi fg="red">%s</ansi></ansi>`, room.Zone))
		user.SendText(fmt.Sprintf(`  <ansi fg="yellow-bold">Root Room Id:    </ansi> <ansi fg="red">%d</ansi>`, zoneConfig.RoomId))

		if zoneConfig.MobAutoScale.Maximum == 0 {
			user.SendText(`  <ansi fg="yellow-bold">Mob AutoScale:</ansi>    <ansi fg="red">[disabled]</ansi>`)
		} else {
			user.SendText(fmt.Sprintf(`  <ansi fg="yellow-bold">Mob AutoScale:</ansi>    <ansi fg="red">%d</ansi> - <ansi fg="red">%d</ansi>`, zoneConfig.MobAutoScale.Minimum, zoneConfig.MobAutoScale.Maximum))
		}

		user.SendText(``)

		return true, nil
	}

	// Everthing after this point requires additional args
	if len(args) < 1 {
		user.SendText(`Not enough arguments provided.`)
		return true, nil
	}

	if roomCmd == `set` {

		setWhat := args[0]

		args = args[1:]

		if setWhat == `autoscale` {
			if len(args) < 2 {
				user.SendText(`Use <ansi fg="command">zone set autoscale 0 0</ansi> to clear autoscaling.`)
				return true, nil
			}

			min, _ := strconv.Atoi(args[0])
			max, _ := strconv.Atoi(args[1])

			if min < 0 || max < 0 {
				user.SendText(`Min/Max can't be less than zero.`)
				return true, nil
			}

			zoneConfig.MobAutoScale.Minimum = min
			zoneConfig.MobAutoScale.Maximum = max
			zoneConfig.Validate()

			user.SendText(`Done!`)
			return true, nil
		}

	}

	return true, nil
}
