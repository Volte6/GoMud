package mobcommands

import (
	"strings"
	"time"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/util"
)

// Signature of user command
type MobCommand func(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error)

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
		"callforhelp":    {CallForHelp, false},
		"despawn":        {Despawn, false},
		"drop":           {Drop, false},
		"emote":          {Emote, true},
		"equip":          {Equip, false},
		"get":            {Get, false},
		"give":           {Give, false},
		"givequest":      {GiveQuest, false},
		"gearup":         {Gearup, false},
		"go":             {Go, false},
		"ifitem":         {IfItem, false},     // Special prefix to mob commands
		"ifnotitem":      {IfNotItem, false},  // Special prefix to mob commands
		"ifquest":        {IfQuest, false},    // Special prefix to mob commands
		"ifnotquest":     {IfNotQuest, false}, // Special prefix to mob commands
		"look":           {Look, false},
		"lookforaid":     {LookForAid, false},
		"lookfortrouble": {LookForTrouble, false},
		"portal":         {Portal, false},
		"remove":         {Remove, false},
		"restock":        {Restock, false},
		"restockservant": {RestockServant, false},
		"say":            {Say, true},
		"shout":          {Shout, true},
		"shoot":          {Shoot, false},
		"sneak":          {Sneak, false},
		"suicide":        {Suicide, true},
		//		"stash":  {Stash, false},
		"throw":   {Throw, false},
		"trash":   {Trash, false},
		"uncurse": {Uncurse, false},
		"wander":  {Wander, false},
	}

	aliases map[string]string = map[string]string{
		".":       "say",
		`l`:       `look`,
		`examine`: `look`,
		"enter":   "go",
		"wield":   "equip",
		"wear":    "equip",
		"rem":     "remove",
		"unequip": "remove",
		"unwear":  "remove",
		"unwield": "remove",
		"toss":    "throw",
		"a":       "attack",
		"fight":   "attack",
		"kill":    "attack",
		"k":       "attack",
		"g":       "get",
		"lft":     "lookfortrouble",
		`yell`:    `shout`,
		`scream`:  `shout`,
		`holler`:  `shout`,
	}

	directionalAliases map[string]string = map[string]string{
		`n`:  `north`,
		`s`:  `south`,
		`e`:  `east`,
		`w`:  `west`,
		`u`:  `up`,
		`d`:  `down`,
		`nw`: `northwest`,
		`ne`: `northeast`,
		`sw`: `southwest`,
		`se`: `southeast`,
	}
)

func GetAllMobCommands() []string {
	result := []string{}

	for cmd, _ := range mobCommands {
		result = append(result, cmd)
	}

	return result
}

func TryCommand(cmd string, rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	cmd = strings.ToLower(cmd)
	rest = strings.TrimSpace(rest)

	if alias, ok := aliases[cmd]; ok {
		cmd = alias
	}
	mobDisabled := false

	if mob := mobs.GetInstance(mobId); mob != nil {
		mobDisabled = mob.Character.IsDisabled()
	}
	// Try any room props, only return if the response indicates it was handled
	/*
		if !mobDisabled {
			if response, err := RoomProps(cmd, rest, userId); err != nil {
				return response, err
			} else if response.Handled {
				return response, err
			}
		}
	*/
	if cmdInfo, ok := mobCommands[cmd]; ok {
		if mobDisabled && !cmdInfo.AllowedWhenDowned {
			response := NewMobCommandResponse(mobId)
			response.Handled = true
			return response, nil
		}

		start := time.Now()
		defer func() {
			util.TrackTime(`mob-cmd[`+cmd+`]`, time.Since(start).Seconds())
		}()

		response, err := cmdInfo.Func(rest, mobId, cmdQueue)
		return response, err

	}
	// Try moving if they aren't disabled
	if !mobDisabled {
		start := time.Now()
		defer func() {
			util.TrackTime(`mob-cmd[go]`, time.Since(start).Seconds())
		}()

		if response, err := Go(cmd, mobId, cmdQueue); err != nil {
			return response, err
		} else if response.Handled {
			return response, err
		}

	}
	if emoteText, ok := emoteAliases[cmd]; ok {
		response, err := Emote(emoteText, mobId, cmdQueue)
		return response, err
	}

	return NewMobCommandResponse(mobId), nil
}

func init() {

	// Put directional aliases into the alias map
	for cmd, alias := range directionalAliases {
		aliases[cmd] = alias
	}

}
