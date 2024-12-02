package usercommands

import (
	"fmt"
	"strings"
	"time"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/keywords"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

type CommandHelpItem struct {
	Command   string
	Type      string // command/skill
	Category  string
	AdminOnly bool
}

type CommandAccess struct {
	Func              UserCommand
	AllowedWhenDowned bool
	AdminOnly         bool
}

var (
	userCommands map[string]CommandAccess = map[string]CommandAccess{
		`aid`:         {Aid, false, false},
		`alias`:       {Alias, true, false},
		`appraise`:    {Appraise, false, false},
		`ask`:         {Ask, false, false},
		`attack`:      {Attack, false, false},
		`auction`:     {Auction, true, false},
		`backstab`:    {Backstab, false, false},
		`badcommands`: {BadCommands, true, true}, // Admin only
		`biome`:       {Biome, true, false},
		`broadcast`:   {Broadcast, true, false},
		`character`:   {Character, true, false},
		`tackle`:      {Tackle, false, false},
		`bank`:        {Bank, false, false},
		`break`:       {Break, false, false},
		`build`:       {Build, false, true},  // Admin only
		`ibuild`:      {IBuild, false, true}, // Admin only
		`buff`:        {Buff, false, true},   // Admin only
		`bump`:        {Bump, false, false},
		`buy`:         {Buy, false, false},
		`cast`:        {Cast, false, false},
		`cooldowns`:   {Cooldowns, true, false},
		`command`:     {Command, false, true}, // Admin only
		`conditions`:  {Conditions, true, false},
		`consider`:    {Consider, true, false},
		`deafen`:      {Deafen, true, true}, // Admin only
		`default`:     {Default, false, false},
		`disarm`:      {Disarm, false, false},
		`drop`:        {Drop, true, false},
		`drink`:       {Drink, false, false},
		`eat`:         {Eat, false, false},
		`emote`:       {Emote, true, false},
		`enchant`:     {Enchant, false, false},
		`exits`:       {Exits, true, false},
		`experience`:  {Experience, true, false},
		`equip`:       {Equip, false, false},
		`flee`:        {Flee, false, false},
		`follow`:      {Follow, false, false},
		`gearup`:      {Gearup, false, false},
		`get`:         {Get, false, false},
		`give`:        {Give, false, false},
		`go`:          {Go, false, false},
		`grant`:       {Grant, true, true}, // Admin only
		`help`:        {Help, true, false},
		`keyring`:     {KeyRing, true, false},
		`killstats`:   {Killstats, true, false},
		`history`:     {History, true, false},
		`inbox`:       {Inbox, true, false},
		`inspect`:     {Inspect, false, false},
		`inventory`:   {Inventory, true, false},
		`jobs`:        {Jobs, true, false},
		`leaderboard`: {Leaderboards, true, false},
		`list`:        {List, false, false},
		`locate`:      {Locate, true, true}, // Admin only
		`lock`:        {Lock, false, false},
		`look`:        {Look, true, false},
		`map`:         {Map, false, false},
		`mudmail`:     {Mudmail, true, true}, // Admin only
		`macros`:      {Macros, true, false},
		`modify`:      {Modify, true, true}, // Admin only
		`motd`:        {Motd, true, false},
		`mute`:        {Mute, true, true},
		`offer`:       {Offer, false, false},
		`online`:      {Online, true, false},
		`party`:       {Party, true, false},
		`password`:    {Password, true, false},
		`paz`:         {Paz, true, true}, // Admin only
		`peep`:        {Peep, false, false},
		`pet`:         {Pet, false, false},
		`picklock`:    {Picklock, false, false},
		`pickpocket`:  {Pickpocket, false, false},
		`prepare`:     {Prepare, true, true}, // Admin only
		`portal`:      {Portal, false, false},
		`pray`:        {Pray, false, false},
		`print`:       {Print, true, false},
		`printline`:   {PrintLine, true, false},
		`put`:         {Put, false, false},
		`pvp`:         {Pvp, true, false},
		`quests`:      {Quests, true, false},
		`quit`:        {Quit, true, false},
		`questtoken`:  {QuestToken, false, true}, // Admin only
		`rank`:        {Rank, false, false},
		`read`:        {Read, false, false},
		`recover`:     {Recover, false, false},
		`reload`:      {Reload, true, true}, // Admin only
		`remove`:      {Remove, false, false},
		`rename`:      {Rename, false, true},     // Admin only
		`redescribe`:  {Redescribe, false, true}, // Admin only
		`room`:        {Room, false, true},       // Admin only
		`save`:        {Save, true, false},
		`say`:         {Say, true, false},
		`scribe`:      {Scribe, false, false},
		`search`:      {Search, false, false},
		`sell`:        {Sell, false, false},
		`server`:      {Server, false, true}, // Admin only
		`set`:         {Set, true, false},
		`share`:       {Share, false, false},
		`shoot`:       {Shoot, false, false},
		`shout`:       {Shout, true, false},
		`show`:        {Show, true, false},
		`skills`:      {Skills, true, false},
		`skillset`:    {Skillset, false, true}, // Admin only
		`sneak`:       {Sneak, false, false},
		`spawn`:       {Spawn, false, true}, // Admin only
		`spells`:      {Spells, true, false},
		`stash`:       {Stash, false, false},
		`status`:      {Status, true, false},
		`storage`:     {Storage, false, false},
		`suicide`:     {Suicide, true, false},
		`tame`:        {Tame, false, false},
		`time`:        {Time, true, false},
		`throw`:       {Throw, false, false},
		`track`:       {Track, false, false},
		`trash`:       {Trash, false, false},
		`train`:       {Train, false, false},
		`unenchant`:   {Unenchant, false, false},
		`uncurse`:     {Uncurse, false, false},
		`unlock`:      {Unlock, false, false},
		`undeafen`:    {UnDeafen, true, true}, // Admin only
		`unmute`:      {UnMute, true, true},   // Admin only
		`use`:         {Use, false, false},
		`dual-wield`:  {DualWield, true, false},
		`whisper`:     {Whisper, true, false},
		`who`:         {Who, true, false},
		`zap`:         {Zap, true, true},   // Admin only
		`zone`:        {Zone, false, true}, // Admin only
		// Special command only used upon creating a new account
		`start`: {Start, false, false},
	}

	selfKeywords = []string{
		`me`,
		`self`,
		`myself`,
	}
)

// Returns a list of close match commands
func GetCmdSuggestions(text string, includeAdmin bool) []string {

	text = strings.ToLower(text)

	results := []string{}

	for _, info := range keywords.GetAllHelpTopicInfo() {
		if info.AdminOnly && !includeAdmin {
			continue
		}

		testCmd := strings.ToLower(info.Command)
		if testCmd != text && strings.HasPrefix(testCmd, text) {
			results = append(results, info.Command[len(text):])
		}
	}

	for alias, _ := range keywords.GetAllCommandAliases() {
		testCmd := strings.ToLower(alias)
		if testCmd != text && strings.HasPrefix(testCmd, text) {
			results = append(results, alias[len(text):])
		}
	}

	return results
}
func GetHelpSuggestions(text string, includeAdmin bool) []string {

	results := []string{}

	for _, cmd := range keywords.GetAllHelpTopics() {
		testCmd := strings.ToLower(cmd)
		if testCmd != text && strings.HasPrefix(testCmd, text) {
			results = append(results, cmd[len(text):])
		}
	}

	for alias, _ := range keywords.GetAllHelpAliases() {
		testCmd := strings.ToLower(alias)
		if testCmd != text && strings.HasPrefix(testCmd, text) {
			results = append(results, alias[len(text):])
		}
	}

	return results
}

// Signature of user command
type UserCommand func(rest string, user *users.UserRecord, room *rooms.Room) (bool, error)

func TryCommand(cmd string, rest string, userId int) (bool, error) {

	// Do not allow scripts to intercept server commands
	if cmd != `server` {

		alias := keywords.TryCommandAlias(cmd)
		skipScript := false
		if info, ok := userCommands[alias]; ok && info.AdminOnly {
			skipScript = true
		}

		if !skipScript {
			handled, err := scripting.TryRoomCommand(alias, rest, userId)
			if handled {
				return true, err
			}
		}

	}

	userDisabled := false
	isAdmin := false
	user := users.GetByUserId(userId)
	if user == nil {
		return false, fmt.Errorf(`user %d not found`, userId)
	}

	// Experimental, not sure if will have unexpected consequences.
	// Turn keywords for targetting self into actual string of self
	for _, selfWord := range selfKeywords {
		wordLen := len(selfWord)
		if rest == selfWord {
			rest = user.Character.Name
			break
		} else if len(rest) >= wordLen+1 && rest[len(rest)-(wordLen+1):] == ` `+selfWord {
			rest = rest[:len(rest)-(wordLen+1)] + ` ` + user.Character.Name
			break
		}
	}

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	rest = strings.TrimSpace(rest)

	// Figure out whether it was an exit entered
	exitName, _ := room.FindExitByName(cmd)
	if exitName != `` {
		rest = cmd
		cmd = "go"
	} else {

		if alias := keywords.TryCommandAlias(cmd); alias != cmd {
			if strings.Contains(alias, ` `) {
				parts := strings.Split(alias, ` `)
				cmd = parts[0]                                         // grab the first word as the new cmd
				rest = strings.TrimPrefix(alias, cmd+` `) + ` ` + rest // add the remaining alias to the rest
			} else {
				cmd = alias
			}
		}

		// Cancel any buffs they have that get cancelled based on them doing anything at all
		user.Character.CancelBuffsWithFlag(buffs.CancelOnAction)

		userDisabled = user.Character.IsDisabled()
		isAdmin = user.Permission == users.PermissionAdmin
		isAdmin = isAdmin || user.HasAdminCommand(cmd)

		// Check if the "rest" is an item the character has
		matchingItem, found := user.Character.FindInBackpack(rest)
		if !found {
			matchingItem, found = user.Character.FindOnBody(rest)
		}

		if found {
			// If the item has a script, run it
			if handled, err := scripting.TryItemCommand(cmd, matchingItem, user.UserId); err == nil {
				if handled { // For this event, handled represents whether to reject the move.
					return handled, err
				}
			}
		}

	}

	if cmdInfo, ok := userCommands[cmd]; ok {

		if userDisabled && !cmdInfo.AllowedWhenDowned && !cmdInfo.AdminOnly {
			user.SendText("You are unable to do that while downed.")
			return true, nil
		}

		if isAdmin || !cmdInfo.AdminOnly {

			start := time.Now()
			defer func() {
				util.TrackTime(`usr-cmd[`+cmd+`]`, time.Since(start).Seconds())
			}()

			// Run the command here
			handled, err := cmdInfo.Func(rest, user, room)
			return handled, err

		}
	}

	if _, ok := emoteAliases[cmd]; ok {
		handled, err := Emote(cmd, user, room)
		return handled, err
	}

	if user.Character.HasSpell(cmd) {
		castCmd := cmd
		if len(rest) > 0 {
			castCmd += ` ` + rest
		}
		return Cast(castCmd, user, room)
	}

	// "go" attempt
	start := time.Now()
	defer func() {
		util.TrackTime(`usr-cmd[go]`, time.Since(start).Seconds())
	}()

	if handled, err := Go(cmd, user, room); handled {
		return handled, err
	}
	// end "go" attempt

	return false, nil
}
