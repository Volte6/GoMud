package rooms

import (
	"fmt"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/exit"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mutators"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

type RoomTemplateDetails struct {
	VisiblePlayers []string
	VisibleMobs    []string
	VisibleCorpses []string
	VisibleExits   map[string]exit.RoomExit
	TemporaryExits map[string]exit.TemporaryRoomExit
	UserId         int
	Character      *characters.Character
	Permission     string
	RoomSymbol     string
	RoomLegend     string
	Nouns          []string
	Zone           string
	Title          string
	Description    string
	IsDark         bool
	IsNight        bool
	TrackingString string
	RoomAlerts     []string // Messages to show below room description as a special alert
	ShowPvp        bool     // Whether to display that the room is PVP
}

func GetDetails(r *Room, user *users.UserRecord) RoomTemplateDetails {

	c := configs.GetGamePlayConfig()

	var roomSymbol string = r.MapSymbol
	var roomLegend string = r.MapLegend

	b := r.GetBiome()

	if b.symbol != 0 {
		roomSymbol = string(b.symbol)
	}
	if b.name != `` {
		roomLegend = b.name
	}

	showPvp := false
	// Don't need to show the PVP flag if Pvp is globally enabled or globally disabled
	if c.PVP == configs.PVPLimited {
		showPvp = r.IsPvp()
	}

	details := RoomTemplateDetails{
		VisiblePlayers: []string{},
		VisibleMobs:    []string{},
		VisibleCorpses: []string{},
		VisibleExits:   make(map[string]exit.RoomExit),
		TemporaryExits: make(map[string]exit.TemporaryRoomExit),
		Zone:           r.Zone,
		Title:          r.Title,
		Description:    r.GetDescription(),
		UserId:         user.UserId,     // Who is viewing the room
		Character:      user.Character,  // The character of the user viewing the room
		Permission:     user.Permission, // The permission level of the user viewing the room
		RoomSymbol:     roomSymbol,
		RoomLegend:     roomLegend,
		IsDark:         b.IsDark(),
		IsNight:        gametime.IsNight(),
		TrackingString: ``,
		ShowPvp:        showPvp,
	}

	//
	// Start Room Alerts
	//

	if len(r.SkillTraining) > 0 {
		details.RoomAlerts = append(details.RoomAlerts, `<ansi fg="yellow-bold">You can train here!</ansi> Type <ansi fg="command">train</ansi> to see what training is available.`)
	}

	if r.IsBank {
		details.RoomAlerts = append(details.RoomAlerts, `          <ansi fg="yellow-bold">This is a bank!</ansi> Type <ansi fg="command">bank</ansi> to deposit/withdraw.`)
	}

	if r.IsStorage {
		details.RoomAlerts = append(details.RoomAlerts, ` <ansi fg="yellow-bold">This is an item storage location!</ansi> Type <ansi fg="command">storage</ansi> to store/unstore.`)
	}

	if r.IsCharacterRoom {
		details.RoomAlerts = append(details.RoomAlerts, `      <ansi fg="yellow-bold">This is a character room!</ansi> Type <ansi fg="command">character</ansi> to interact.`)
	}

	if r.RoomId == -1 {
		details.RoomAlerts = append(details.RoomAlerts, `      <ansi fg="yellow-bold">Type <ansi fg="command">start</ansi> to begin playing.</ansi>`)
	}

	//
	// End Room Alerts
	//

	tinymap := GetTinyMap(r.RoomId)

	renderNouns := user.Permission == users.PermissionAdmin
	if user.Character.Pet.Exists() && user.Character.HasBuffFlag(buffs.SeeNouns) {
		renderNouns = true
	}

	if tinyMapOn := user.GetConfigOption(`tinymap`); tinyMapOn != nil && tinyMapOn.(bool) {
		desclineWidth := 80 - 7 // 7 is the width of the tinymap
		padding := 1
		description := util.SplitString(details.Description, desclineWidth-padding)

		for i := 0; i < len(tinymap); i++ {
			if i > len(description)-1 {
				description = append(description, strings.Repeat(` `, desclineWidth))
			}

			description[i] += strings.Repeat(` `, desclineWidth-len(description[i])) + tinymap[i]
		}

		if renderNouns && len(r.Nouns) > 0 {
			for i := range description {
				for noun, _ := range r.Nouns {
					description[i] = strings.Replace(description[i], noun, `<ansi fg="noun">`+noun+`</ansi>`, 1)
				}
			}
		}

		details.Description = strings.Join(description, "\n")
	} else {

		roomDesc := util.SplitString(details.Description, 80)

		if renderNouns && len(r.Nouns) > 0 {
			for i := range roomDesc {
				for noun, _ := range r.Nouns {
					roomDesc[i] = strings.Replace(roomDesc[i], noun, `<ansi fg="noun">`+noun+`</ansi>`, 1)
				}
			}
		}

		details.Description = strings.Join(roomDesc, "\n")
	}

	for mut := range r.ActiveMutators {
		mutSpec := mut.GetSpec()

		if mutSpec.NameModifier != nil {

			if mutSpec.NameModifier.Behavior == mutators.TextPrepend {

				if mutSpec.NameModifier.Text != `` {
					details.Title = colorpatterns.ApplyColorPattern(mutSpec.NameModifier.Text, mutSpec.NameModifier.ColorPattern) + ` ` + details.Title
				}

			} else if mutSpec.NameModifier.Behavior == mutators.TextReplace {

				if mutSpec.NameModifier.Text != `` {
					details.Title = colorpatterns.ApplyColorPattern(mutSpec.NameModifier.Text, mutSpec.NameModifier.ColorPattern)
				} else {
					details.Title = colorpatterns.ApplyColorPattern(details.Title, mutSpec.NameModifier.ColorPattern)
				}

			} else if mutSpec.NameModifier.Behavior == mutators.TextAppend {

				if mutSpec.NameModifier.Text != `` {
					details.Title = details.Title + ` ` + colorpatterns.ApplyColorPattern(mutSpec.NameModifier.Text, mutSpec.NameModifier.ColorPattern)
				}

			}

		}

		if mutSpec.DescriptionModifier != nil {

			// Handle any text changes
			if mutSpec.DescriptionModifier.Behavior == mutators.TextPrepend {

				if mutSpec.DescriptionModifier.Text != `` {

					details.Description = colorpatterns.ApplyColorPattern(mutSpec.DescriptionModifier.Text, mutSpec.DescriptionModifier.ColorPattern) +
						term.CRLFStr +
						details.Description

				}

			} else if mutSpec.DescriptionModifier.Behavior == mutators.TextReplace {

				if mutSpec.DescriptionModifier.Text != `` {
					details.Description = colorpatterns.ApplyColorPattern(mutSpec.DescriptionModifier.Text, mutSpec.DescriptionModifier.ColorPattern)
				} else {
					details.Description = colorpatterns.ApplyColorPattern(details.Description, mutSpec.DescriptionModifier.ColorPattern)
				}

			} else if mutSpec.DescriptionModifier.Behavior == mutators.TextAppend {

				if mutSpec.DescriptionModifier.Text != `` {

					details.Description = details.Description +
						term.CRLFStr +
						colorpatterns.ApplyColorPattern(mutSpec.DescriptionModifier.Text, mutSpec.DescriptionModifier.ColorPattern)

				}
			}

		}

		// Alert modifiers can only add to the list.
		// No current plans to allow them to overwrite existing alerts.
		if mutSpec.AlertModifier != nil {

			alertText := mutSpec.AlertModifier.Text

			// center the text
			if len(mutSpec.AlertModifier.Text) < 65 {
				padding := (65 - len(mutSpec.AlertModifier.Text)) >> 1
				alertText = strings.Repeat(` `, padding) + alertText
			}

			details.RoomAlerts = append(details.RoomAlerts, colorpatterns.ApplyColorPattern(alertText, mutSpec.AlertModifier.ColorPattern))

		}
	}

	nameFlags := []characters.NameRenderFlag{}
	if user.Character.GetSkillLevel(skills.Peep) > 0 {
		nameFlags = append(nameFlags, characters.RenderHealth)
	}

	if useShortAdjectives := user.GetConfigOption(`shortadjectives`); useShortAdjectives != nil && useShortAdjectives.(bool) {
		nameFlags = append(nameFlags, characters.RenderShortAdjectives)
	}

	for _, playerId := range r.players {
		if playerId != user.UserId {

			renderFlags := append([]characters.NameRenderFlag{}, nameFlags...)

			player := users.GetByUserId(playerId)
			if player != nil {

				if player.Character.HasBuffFlag(buffs.Hidden) { // Don't show them if they are sneaking
					if !user.Character.Pet.Exists() || !user.Character.HasBuffFlag(buffs.SeeHidden) {
						continue
					}
				}

				pName := player.Character.GetPlayerName(user.UserId, renderFlags...)
				details.VisiblePlayers = append(details.VisiblePlayers, pName.String())
			}
		}
	}

	if user.Character.Pet.Exists() && r.RoomId == user.Character.RoomId {
		details.VisiblePlayers = append(details.VisiblePlayers, fmt.Sprintf(`%s (your pet)`, user.Character.Pet.DisplayName()))
	}

	visibleFriendlyMobs := []string{}

	for idx, mobInstanceId := range r.mobs {
		if mob := mobs.GetInstance(mobInstanceId); mob != nil {

			if mob.Character.HasBuffFlag(buffs.Hidden) { // Don't show them if they are sneaking
				if !user.Character.Pet.Exists() || !user.Character.HasBuffFlag(buffs.SeeHidden) {
					continue
				}
			}

			tmpNameFlags := nameFlags

			mobName := mob.Character.GetMobName(user.UserId, tmpNameFlags...)

			for _, qFlag := range mob.QuestFlags {
				if user.Character.HasQuest(qFlag) || (len(qFlag) >= 5 && qFlag[len(qFlag)-5:] == `start`) {
					mobName.QuestAlert = true
					break
				}
			}

			if mob.Character.IsCharmed() {
				visibleFriendlyMobs = append(visibleFriendlyMobs, mobName.String())
			} else {
				details.VisibleMobs = append(details.VisibleMobs, mobName.String())
			}
		} else {
			r.mobs = append(r.mobs[:idx], r.mobs[idx+1:]...)
		}
	}

	// Add the friendly mobs to the end
	details.VisibleMobs = append(details.VisibleMobs, visibleFriendlyMobs...)

	for exitStr, exitInfo := range r.ExitsTemp {
		details.TemporaryExits[exitStr] = exitInfo
	}

	// Do this twice to ensure secrets are last

	for exitStr, exitInfo := range r.Exits {

		// If it's a secret room we need to make sure the player has recently been there before including it in the exits
		if exitInfo.Secret { //&& user.Permission != users.PermissionAdmin {
			if targetRm := LoadRoom(exitInfo.RoomId); targetRm != nil {
				if targetRm.HasVisited(user.UserId, VisitorUser) {
					details.VisibleExits[exitStr] = exitInfo
				}
			}
		} else {
			details.VisibleExits[exitStr] = exitInfo
		}
	}

	// add any corpses present
	mobCorpses := map[string]int{}
	playerCorpses := map[string]int{}

	for _, c := range r.Corpses {
		if c.Prunable {
			continue
		}

		if c.MobId > 0 {
			mobCorpses[c.Character.Name] = mobCorpses[c.Character.Name] + 1
		}

		if c.UserId > 0 {
			playerCorpses[c.Character.Name] = playerCorpses[c.Character.Name] + 1
		}

	}

	for name, qty := range playerCorpses {
		if qty == 1 {
			details.VisibleCorpses = append(details.VisibleCorpses, fmt.Sprintf(`<ansi fg="user-corpse">%s corpse</ansi>`, name))
		} else {
			details.VisibleCorpses = append(details.VisibleCorpses, fmt.Sprintf(`<ansi fg="user-corpse">%d %s corpses</ansi>`, qty, name))
		}
	}

	for name, qty := range mobCorpses {
		if qty == 1 {
			details.VisibleCorpses = append(details.VisibleCorpses, fmt.Sprintf(`<ansi fg="mob-corpse">%s corpse</ansi>`, name))
		} else {
			details.VisibleCorpses = append(details.VisibleCorpses, fmt.Sprintf(`<ansi fg="mob-corpse">%d %s corpses</ansi>`, qty, name))
		}
	}

	// assign mutator exits last so that they can overwrite normal exits
	for mut := range r.ActiveMutators {
		spec := mut.GetSpec()
		for exitName, exitInfo := range spec.Exits {
			details.VisibleExits[exitName] = exitInfo
		}
	}

	if searchMobName := user.Character.GetMiscData(`tracking-mob`); searchMobName != nil {

		if searchMobNameStr, ok := searchMobName.(string); ok {

			if r.isInRoom(searchMobNameStr, ``) {

				details.TrackingString = `Tracking <ansi fg="mobname">` + searchMobNameStr + `</ansi>... They are here!`
				user.Character.RemoveBuff(26)

			} else {

				allNames := []string{}

				for mobInstId, _ := range r.Visitors(VisitorMob) {
					if mob := mobs.GetInstance(mobInstId); mob != nil {
						allNames = append(allNames, mob.Character.Name)
					}
				}

				match, closeMatch := util.FindMatchIn(searchMobNameStr, allNames...)
				if match == `` && closeMatch == `` {

					details.TrackingString = `You lost the trail of <ansi fg="mobname">` + searchMobNameStr + `</ansi>`
					user.Character.RemoveBuff(26)

				} else {

					exitName := r.findMobExit(0, searchMobNameStr)
					if exitName == `` {

						details.TrackingString = `You lost the trail of <ansi fg="username">` + searchMobNameStr + `</ansi>`
						user.Character.RemoveBuff(26)

					} else {

						details.TrackingString = `Tracking <ansi fg="mobname">` + searchMobNameStr + `</ansi>... They went <ansi fg="exit">` + exitName + `</ansi>`
					}

				}
			}
		}

	}

	if searchUserName := user.Character.GetMiscData(`tracking-user`); searchUserName != nil {
		if searchUserNameStr, ok := searchUserName.(string); ok {

			if r.isInRoom(``, searchUserNameStr) {

				details.TrackingString = `Tracking <ansi fg="username">` + searchUserNameStr + `</ansi>... They are here!`
				user.Character.RemoveBuff(26)

			} else {

				allNames := []string{}

				for userId, _ := range r.Visitors(VisitorUser) {
					if u := users.GetByUserId(userId); u != nil {
						allNames = append(allNames, u.Character.Name)
					}
				}

				match, closeMatch := util.FindMatchIn(searchUserNameStr, allNames...)
				if match == `` && closeMatch == `` {

					details.TrackingString = `You lost the trail of <ansi fg="username">` + searchUserNameStr + `</ansi>`
					user.Character.RemoveBuff(26)

				} else {

					exitName := r.findUserExit(0, searchUserNameStr)
					if exitName == `` {

						details.TrackingString = `You lost the trail of <ansi fg="username">` + searchUserNameStr + `</ansi>`
						user.Character.RemoveBuff(26)

					} else {

						details.TrackingString = `Tracking <ansi fg="username">` + searchUserNameStr + `</ansi>... They went <ansi fg="exit">` + exitName + `</ansi>`
					}

				}
			}

		}
	}

	return details

}
