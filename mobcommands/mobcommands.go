package mobcommands

import (
	"strings"
	"time"

	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/util"
)

// Signature of user command
type MobCommand func(rest string, mobId int) (bool, string, error)

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
		"portal":         {Portal, false},
		"remove":         {Remove, false},
		"restock":        {Restock, false},
		"restockservant": {RestockServant, false},
		"say":            {Say, true},
		"sayto":          {SayTo, true},
		"shout":          {Shout, true},
		"shoot":          {Shoot, false},
		"show":           {Show, false},
		"sneak":          {Sneak, false},
		"suicide":        {Suicide, true},
		//		"stash":  {Stash, false},
		"throw":   {Throw, false},
		"trash":   {Trash, false},
		"uncurse": {Uncurse, false},
		"wander":  {Wander, false},
	}
)

func GetAllMobCommands() []string {
	result := []string{}

	for cmd, _ := range mobCommands {
		result = append(result, cmd)
	}

	return result
}

func TryCommand(cmd string, rest string, mobId int) (bool, string, error) {

	cmd = strings.ToLower(cmd)
	rest = strings.TrimSpace(rest)

	cmd = keywords.TryCommandAlias(cmd)

	mobDisabled := false

	if mob := mobs.GetInstance(mobId); mob != nil {
		mobDisabled = mob.Character.IsDisabled()
	}
	// Try any room props, only return if the response indicates it was handled
	/*
		if !mobDisabled {
			if handled, nextCommand, err := RoomProps(cmd, rest, userId); err != nil {
				return response, err
			} else if response.Handled {
				return response, err
			}
		}
	*/
	if cmdInfo, ok := mobCommands[cmd]; ok {
		if mobDisabled && !cmdInfo.AllowedWhenDowned {

			return true, ``, nil
		}

		start := time.Now()
		defer func() {
			util.TrackTime(`mob-cmd[`+cmd+`]`, time.Since(start).Seconds())
		}()

		handled, nextCommand, err := cmdInfo.Func(rest, mobId)
		return handled, nextCommand, err

	}
	// Try moving if they aren't disabled
	if !mobDisabled {
		start := time.Now()
		defer func() {
			util.TrackTime(`mob-cmd[go]`, time.Since(start).Seconds())
		}()

		if handled, nextCommand, err := Go(cmd, mobId); err != nil {
			return handled, nextCommand, err
		} else if handled {
			return true, ``, nil
		}

	}
	if emoteText, ok := emoteAliases[cmd]; ok {
		handled, nextCommand, err := Emote(emoteText, mobId)
		return handled, nextCommand, err
	}

	return false, ``, nil
}
