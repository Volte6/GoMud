package mobcommands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/volte6/gomud/keywords"
	"github.com/volte6/gomud/mobs"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/util"
)

// Signature of user command
type MobCommand func(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error)

type CommandAccess struct {
	Func              MobCommand
	AllowedWhenDowned bool
}

var (
	mobCommands map[string]CommandAccess = map[string]CommandAccess{
		"aid":            {Aid, false},
		"alchemy":        {Alchemy, false},
		"attack":         {Attack, false},
		"backstab":       {Backstab, false},
		"befriend":       {Befriend, false},
		"break":          {Break, false},
		"broadcast":      {Broadcast, false},
		"cast":           {Cast, false},
		"converse":       {Converse, false},
		"callforhelp":    {CallForHelp, false},
		"despawn":        {Despawn, false},
		"drink":          {Drink, false},
		"drop":           {Drop, false},
		"eat":            {Eat, false},
		"emote":          {Emote, true},
		"equip":          {Equip, false},
		"get":            {Get, false},
		"give":           {Give, false},
		"givequest":      {GiveQuest, false},
		"gearup":         {Gearup, false},
		"go":             {Go, false},
		"look":           {Look, false},
		"lookforaid":     {LookForAid, false},
		"lookfortrouble": {LookForTrouble, false},
		"noop":           {Noop, true},
		"portal":         {Portal, false},
		"remove":         {Remove, false},
		"say":            {Say, true},
		"sayto":          {SayTo, true},
		"shout":          {Shout, true},
		"shoot":          {Shoot, false},
		"show":           {Show, false},
		"sneak":          {Sneak, false},
		"suicide":        {Suicide, true},
		//		"stash":  {Stash, false},
		"throw":  {Throw, false},
		"trash":  {Trash, false},
		"wander": {Wander, false},
	}
)

func GetAllMobCommands() []string {
	result := []string{}

	for cmd, _ := range mobCommands {
		result = append(result, cmd)
	}

	return result
}

func TryCommand(cmd string, rest string, mobId int) (bool, error) {

	cmd = strings.ToLower(cmd)
	rest = strings.TrimSpace(rest)

	cmd = keywords.TryCommandAlias(cmd)

	mobDisabled := false

	mob := mobs.GetInstance(mobId)
	if mob == nil {
		return false, errors.New(`mob instance doesn't exist`)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	mobDisabled = mob.Character.IsDisabled()

	// Try any room props, only return if the response indicates it was handled
	/*
		if !mobDisabled {
			if handled, err := RoomProps(cmd, rest, userId); err != nil {
				return response, err
			} else if response.Handled {
				return response, err
			}
		}
	*/

	if cmdInfo, ok := mobCommands[cmd]; ok {
		if mobDisabled && !cmdInfo.AllowedWhenDowned {

			return true, nil
		}

		start := time.Now()
		defer func() {
			util.TrackTime(`mob-cmd[`+cmd+`]`, time.Since(start).Seconds())
		}()

		handled, err := cmdInfo.Func(rest, mob, room)
		return handled, err

	}
	// Try moving if they aren't disabled
	if !mobDisabled {
		start := time.Now()
		defer func() {
			util.TrackTime(`mob-cmd[go]`, time.Since(start).Seconds())
		}()

		if handled, err := Go(cmd, mob, room); err != nil {
			return handled, err
		} else if handled {
			return true, nil
		}

	}
	if emoteText, ok := emoteAliases[cmd]; ok {
		handled, err := Emote(emoteText, mob, room)
		return handled, err
	}

	return false, nil
}
