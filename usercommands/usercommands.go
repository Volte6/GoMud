package usercommands

import (
	"strings"
	"time"

	"github.com/volte6/mud/buffs"
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
	helpCommands = []CommandHelpItem{
		{`alias`, `command`, `configuration`, false},
		{`alignment`, `command`, `character`, false},
		{`armor`, `command`, `character`, false},
		{`appraise`, `command`, `shops`, false},
		{`ask`, `command`, `quests`, false},
		{`attack`, `command`, `combat`, false},
		{`bank`, `command`, `shops`, false},
		{`biome`, `command`, `information`, false},
		{`break`, `command`, `combat`, false},
		{`build`, `command`, ``, true},
		{`buff`, `command`, ``, true},
		{`buy`, `command`, `shops`, false},
		{`command`, `command`, ``, true},
		{`conditions`, `command`, `character`, false},
		{`consider`, `command`, `combat`, false},
		{`cooldowns`, `command`, `character`, false},
		{`deposit`, `command`, `shops`, false},
		{`drop`, `command`, `items`, false},
		{`drink`, `command`, `items`, false},
		{`eat`, `command`, `items`, false},
		{`emote`, `command`, `general`, false},
		{`exits`, `command`, `information`, false},
		{`experience`, `command`, `character`, false},
		{`equip`, `command`, `items`, false},
		{`follow`, `command`, `parties`, false},
		{`flee`, `command`, `combat`, false},
		{`get`, `command`, `items`, false},
		{`give`, `command`, `items`, false},
		{`go`, `command`, `general`, false},
		{`help`, `command`, `information`, false},
		{`hire`, `command`, `shops`, false},
		{`inventory`, `command`, `character`, false},
		{`jobs`, `command`, `character`, false},
		{`keyring`, `command`, `character`, false},
		{`list`, `command`, `shops`, false},
		{`locate`, `command`, ``, true},
		{`lock`, `command`, `locks`, false},
		{`look`, `command`, `information`, false},
		{`macros`, `command`, `configuration`, false},
		{`offer`, `command`, `shops`, false},
		{`online`, `command`, `information`, false},
		{`party`, `command`, `parties`, false},
		{`picklock`, `command`, `locks`, false},
		{`prepare`, `command`, ``, true},
		{`quests`, `command`, `quests`, false},
		{`questtoken`, `command`, ``, true},
		{`races`, `command`, `information`, false},
		{`read`, `command`, `general`, false},
		{`reload`, `command`, ``, true},
		{`remove`, `command`, `items`, false},
		{`rename`, `command`, ``, true},
		{`room`, `command`, ``, true},
		{`say`, `command`, `general`, false},
		{`sell`, `command`, `shops`, false},
		{`server`, `command`, ``, true},
		{`set`, `command`, `configuration`, false},
		{`share`, `command`, `parties`, false},
		{`shoot`, `command`, `combat`, false},
		{`shout`, `command`, `general`, false},
		{`skills`, `command`, `character`, false},
		{`skillset`, `command`, ``, true},
		{`spawn`, `command`, ``, true},
		{`stash`, `command`, `items`, false},
		{`status`, `command`, `character`, false},
		{`store`, `command`, `shops`, false},
		{`throw`, `command`, `items`, false},
		{`train`, `command`, `general`, false},
		{`trash`, `command`, `items`, false},
		{`unlock`, `command`, `locks`, false},
		{`unstore`, `command`, `shops`, false},
		{`use`, `command`, `items`, false},
		{`withdraw`, `command`, `shops`, false},
		{`whisper`, `command`, `general`, false},
		{`who`, `command`, `information`, false},
		{`zap`, `command`, ``, true},
		// skills
		{`aid`, `skill`, ``, false},
		{`backstab`, `skill`, ``, false},
		{`brawling`, `skill`, ``, false},
		{`bump`, `skill`, ``, false},
		{`dual-wield`, `skill`, ``, false},
		{`tackle`, `skill`, ``, false},
		{`disarm`, `skill`, ``, false},
		{`recover`, `skill`, ``, false},
		{`enchant`, `skill`, ``, false},
		{`inspect`, `skill`, ``, false},
		{`map`, `skill`, ``, false},
		{`rank`, `skill`, ``, false},
		{`peep`, `skill`, ``, false},
		{`pickpocket`, `skill`, ``, false},
		{`portal`, `skill`, ``, false},
		{`pray`, `skill`, ``, false},
		{`rank`, `skill`, ``, false},
		{`scribe`, `skill`, ``, false},
		{`search`, `skill`, ``, false},
		{`skulduggery`, `skill`, ``, false},
		{`sneak`, `skill`, ``, false},
		{`track`, `skill`, ``, false},
		{`unenchant`, `skill`, ``, false},
		{`uncurse`, `skill`, ``, false},
	}

	// remaps for help
	helpAliases = map[string]string{
		`brawl`:        `brawling`,
		`tackle`:       `brawling`,
		`disarm`:       `brawling`,
		`recover`:      `brawling`,
		`throw`:        `brawling`,
		`unenchant`:    `enchant`,
		`uncurse`:      `enchant`,
		`sneak`:        `skulduggery`,
		`bump`:         `skulduggery`,
		`backstab`:     `skulduggery`,
		`pickpocket`:   `skulduggery`,
		`deposit`:      `bank`,
		`withdraw`:     `bank`,
		`store`:        `storage`,
		`unstore`:      `storage`,
		`str`:          `strength`,
		`vit`:          `vitality`,
		`spd`:          `speed`,
		`spe`:          `speed`,
		`mys`:          `mysticism`,
		`myst`:         `mysticism`,
		`smt`:          `smarts`,
		`sma`:          `smarts`,
		`per`:          `perception`,
		`percep`:       `perception`,
		`percept`:      `perception`,
		`hp`:           `health`,
		`mp`:           `mana`,
		`race`:         `races`,
		`rank`:         `protection`,
		`backrank`:     `protection`,
		`frontrank`:    `protection`,
		`aid`:          `protection`,
		`pick`:         `picklock`,
		`pick-example`: `picklock-example`,
		`keys`:         `keyring`,
		`key`:          `keyring`,
		`wield`:        `equip`,
		`wear`:         `equip`,
		`score`:        `status`,
	}

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

	aliases map[string]string = map[string]string{
		`.`:         `say`,
		"`":         `broadcast`,
		`sta`:       `status`,
		`stat`:      `status`,
		`stats`:     `status`,
		`score`:     `status`,
		`i`:         `inventory`,
		`inv`:       `inventory`,
		`eq`:        `inventory`,
		`l`:         `look`,
		`examine`:   `look`,
		`enter`:     `go`,
		`exp`:       `experience`,
		`xp`:        `experience`,
		`tnl`:       `experience`,
		`m`:         `map`,
		`cond`:      `conditions`,
		`condition`: `conditions`,
		`c`:         `conditions`,
		`sk`:        `skills`,
		`skill`:     `skills`,
		`scribble`:  `scribe`,
		`write`:     `scribe`,
		`wield`:     `equip`,
		`wear`:      `equip`,
		`rem`:       `remove`,
		`unequip`:   `remove`,
		`unwear`:    `remove`,
		`unwield`:   `remove`,
		`toss`:      `throw`,
		`a`:         `attack`,
		`fight`:     `attack`,
		`kill`:      `attack`,
		`k`:         `attack`,
		`g`:         `get`,
		`cmd`:       `command`,
		`=?`:        `macros`,
		`pchat`:     `party chat`,
		`psay`:      `party chat`,
		`deposit`:   `bank deposit`,
		`withdraw`:  `bank withdraw`,
		`store`:     `storage add`,
		`unstore`:   `storage remove`,
		`sn`:        `sneak`,
		`bs`:        `backstab`,
		`q`:         `quests`,
		`quest`:     `quests`,
		`backrank`:  `rank back`,
		`frontrank`: `rank front`,
		`yell`:      `shout`,
		`scream`:    `shout`,
		`holler`:    `shout`,
		`pick`:      `picklock`,
		`lockpick`:  `picklock`,
		`lockpicks`: `picklock`,
		`keys`:      `keyring`,
		`key`:       `keyring`,
		`/w`:        `whisper`,
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

	missingHelp = getMissingHelp()
)

func getMissingHelp() []string {

	missing := []string{}
	lut := map[string]struct{}{}
	for i := 0; i < len(helpCommands); i++ {
		lut[helpCommands[i].Command] = struct{}{}
	}

	for cmd, _ := range userCommands {
		if _, ok := lut[cmd]; !ok {
			missing = append(missing, cmd)
		}
	}

	return missing
}

// Signature of user command
type UserCommand func(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error)

func TryCommand(cmd string, rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	cmd = strings.ToLower(cmd)

	if alias, ok := aliases[cmd]; ok {
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

func init() {

	// Put directional aliases into the alias map
	for cmd, alias := range directionalAliases {
		aliases[cmd] = alias
	}

}
