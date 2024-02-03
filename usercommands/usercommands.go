package usercommands

import (
	"strings"
	"time"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
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
		`aid`:        {Aid, false, false},
		`alias`:      {Alias, true, false},
		`appraise`:   {Appraise, false, false},
		`ask`:        {Ask, false, false},
		`attack`:     {Attack, false, false},
		`backstab`:   {Backstab, false, false},
		`biome`:      {Biome, true, false},
		`broadcast`:  {Broadcast, true, false},
		`tackle`:     {Tackle, false, false},
		`bank`:       {Bank, false, false},
		`break`:      {Break, false, false},
		`build`:      {Build, false, true}, // Admin only
		`buff`:       {Buff, false, true},  // Admin only
		`bump`:       {Bump, false, false},
		`buy`:        {Buy, false, false},
		`cast`:       {Cast, false, false},
		`cooldowns`:  {Cooldowns, true, false},
		`command`:    {Command, false, true}, // Admin only
		`conditions`: {Conditions, true, false},
		`consider`:   {Consider, true, false},
		`disarm`:     {Disarm, false, false},
		`drop`:       {Drop, true, false},
		`drink`:      {Drink, false, false},
		`eat`:        {Eat, false, false},
		`emote`:      {Emote, true, false},
		`enchant`:    {Enchant, false, false},
		`exits`:      {Exits, true, false},
		`experience`: {Experience, true, false},
		`equip`:      {Equip, false, false},
		`flee`:       {Flee, false, false},
		`follow`:     {Follow, false, false},
		`gearup`:     {Gearup, false, false},
		`get`:        {Get, false, false},
		`give`:       {Give, false, false},
		`go`:         {Go, false, false},
		`help`:       {Help, true, false},
		`hire`:       {Hire, false, false},
		`keyring`:    {KeyRing, true, false},
		`inspect`:    {Inspect, false, false},
		`inventory`:  {Inventory, true, false},
		`jobs`:       {Jobs, true, false},
		`list`:       {List, false, false},
		`locate`:     {Locate, true, true},
		`lock`:       {Lock, false, false},
		`look`:       {Look, true, false},
		`map`:        {Map, false, false},
		`macros`:     {Macros, true, false},
		`motd`:       {Motd, true, false},
		`offer`:      {Offer, false, false},
		`online`:     {Online, true, false},
		`party`:      {Party, true, false},
		`peep`:       {Peep, false, false},
		`picklock`:   {Picklock, false, false},
		`pickpocket`: {Pickpocket, false, false},
		`prepare`:    {Prepare, true, true}, // Admin only
		`portal`:     {Portal, false, false},
		`pray`:       {Pray, false, false},
		`print`:      {Print, true, false},
		`quests`:     {Quests, true, false},
		`questtoken`: {QuestToken, false, true}, // Admin only
		`rank`:       {Rank, false, false},
		`read`:       {Read, false, false},
		`recover`:    {Recover, false, false},
		`reload`:     {Reload, true, true}, // Admin only
		`remove`:     {Remove, false, false},
		`rename`:     {Rename, false, true}, // Admin only
		`room`:       {Room, false, true},   // Admin only
		`save`:       {Save, true, false},
		`say`:        {Say, true, false},
		`scribe`:     {Scribe, false, false},
		`search`:     {Search, false, false},
		`sell`:       {Sell, false, false},
		`server`:     {Server, false, true}, // Admin only
		`set`:        {Set, true, false},
		`share`:      {Share, false, false},
		`shoot`:      {Shoot, false, false},
		`shout`:      {Shout, true, false},
		`skills`:     {Skills, true, false},
		`skillset`:   {Skillset, false, true}, // Admin only
		`sneak`:      {Sneak, false, false},
		`spawn`:      {Spawn, false, true}, // Admin only
		`stash`:      {Stash, false, false},
		`status`:     {Status, true, false},
		`storage`:    {Storage, false, false},
		`suicide`:    {Suicide, true, false},
		`throw`:      {Throw, false, false},
		`track`:      {Track, false, false},
		`trash`:      {Trash, false, false},
		`train`:      {Train, false, false},
		`unenchant`:  {Unenchant, false, false},
		`uncurse`:    {Uncurse, false, false},
		`unlock`:     {Unlock, false, false},
		`use`:        {Use, false, false},
		`dual-wield`: {DualWield, true, false},
		`whisper`:    {Whisper, true, false},
		`who`:        {Who, true, false},
		`zap`:        {Zap, false, true}, // Admin only
		// Special command only used upon creating a new account
		`start`: {Start, false, false},
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
type UserCommand func(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error)

func TryCommand(cmd string, rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	cmd = strings.ToLower(cmd)

	if alias := keywords.TryCommandAlias(cmd); alias != cmd {
		if strings.Contains(alias, ` `) {
			parts := strings.Split(alias, ` `)
			cmd = parts[0]                                         // grab the first word as the new cmd
			rest = strings.TrimPrefix(alias, cmd+` `) + ` ` + rest // add the remaining alias to the rest
		} else {
			cmd = alias
		}
	}

	rest = strings.TrimSpace(rest)

	userDisabled := false
	isAdmin := false
	if user := users.GetByUserId(userId); user != nil {
		// Cancel any buffs they have that get cancelled based on them doing anything at all
		user.Character.CancelBuffsWithFlag(buffs.CancelOnAction)

		userDisabled = user.Character.IsDisabled()
		isAdmin = user.Permission == users.PermissionAdmin
		isAdmin = isAdmin || user.HasAdminCommand(cmd)
	}

	// Try any room props, only return if the response indicates it was handled
	if !userDisabled {
		if response, err := RoomProps(cmd, rest, userId, cmdQueue); err != nil {
			return response, err
		} else if response.Handled {
			return response, err
		}
	}

	if cmdInfo, ok := userCommands[cmd]; ok {

		if userDisabled && !cmdInfo.AllowedWhenDowned && !cmdInfo.AdminOnly {
			response := NewUserCommandResponse(userId)
			response.SendUserMessage(userId, "You are unable to do that while downed.", true)
			response.Handled = true
			return response, nil
		}

		if isAdmin || !cmdInfo.AdminOnly {

			start := time.Now()
			defer func() {
				util.TrackTime(`usr-cmd[`+cmd+`]`, time.Since(start).Seconds())
			}()

			response, err := cmdInfo.Func(rest, userId, cmdQueue)
			return response, err

		}
	}

	// Try moving if they aren't disabled
	if !userDisabled {

		start := time.Now()
		defer func() {
			util.TrackTime(`usr-cmd[go]`, time.Since(start).Seconds())
		}()

		if response, err := Go(cmd, userId, cmdQueue); err != nil {
			return response, err
		} else if response.Handled {
			return response, err
		}

	}

	if emoteText, ok := emoteAliases[cmd]; ok {
		response, err := Emote(emoteText, userId, cmdQueue)
		return response, err
	}

	return NewUserCommandResponse(userId), nil
}
