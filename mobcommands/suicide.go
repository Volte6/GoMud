package mobcommands

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/combat"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Suicide(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if rest != `vanish` && mob.Character.HasBuffFlag(buffs.ReviveOnDeath) {

		mob.Character.Health = mob.Character.HealthMax.Value

		room.SendText(`<ansi fg="mobname">` + mob.Character.Name + `</ansi> is suddenly revived in a shower of sparks!`)

		mob.Character.CancelBuffsWithFlag(buffs.ReviveOnDeath)

		return true, nil
	}

	slog.Info(`Mob Death`, `name`, mob.Character.Name, `rest`, rest)

	// Make sure to clean up any charm stuff if it's being removed
	if charmedUserId := mob.Character.RemoveCharm(); charmedUserId > 0 {
		if charmedUser := users.GetByUserId(charmedUserId); charmedUser != nil {
			charmedUser.Character.TrackCharmed(mob.InstanceId, false)
		}
	}

	// vanish is meant to remove the mob without any rewards/drops/etc.
	if rest == `vanish` {

		// Destroy any record of this mob.
		mobs.DestroyInstance(mob.InstanceId)

		// Clean up mob from room...
		if r := rooms.LoadRoom(mob.HomeRoomId); r != nil {
			r.CleanupMobSpawns(false)
		}

		// Remove from current room
		room.RemoveMob(mob.InstanceId)

		return true, nil
	}

	// Send a death msg to everyone in the room.
	room.SendText(
		fmt.Sprintf(`<ansi fg="mobname">%s</ansi> has died.`, mob.Character.Name),
	)

	mobXP := mob.Character.XPTL(mob.Character.Level - 1)

	xpVal := mobXP / 125

	xpVariation := xpVal / 100
	if xpVariation < 1 {
		xpVariation = 1
	}

	partyTracker := map[int]int{} // key is party leader ID, value is how much will be shared.

	if len(mob.DamageTaken) > 0 {

		xpVal = xpVal / len(mob.DamageTaken)        // Div by number of players that beat him up
		xpVal += ((util.Rand(3) - 1) * xpVariation) // a little bit of variation

		totalPlayerLevels := 0
		for uId, _ := range mob.DamageTaken {
			if user := users.GetByUserId(uId); user != nil {
				totalPlayerLevels += user.Character.Level
			}
		}

		attackerCt := len(mob.DamageTaken)

		xpMsg := `You gained <ansi fg="experience">%d experience points</ansi>%s!`
		for uId, _ := range mob.DamageTaken {
			if user := users.GetByUserId(uId); user != nil {

				scripting.TryMobScriptEvent(`onDie`, mob.InstanceId, uId, `user`, map[string]any{`attackerCount`: attackerCt})

				p := parties.Get(user.UserId)

				// Not in a party? Great give them the xp.
				if p == nil {

					if mob.Character.Zone != `Training` { // Don't track any kills in the training zone
						user.Character.KD.AddMobKill(int(mob.MobId))
					}

					xpScaler := float64(mob.Character.Level) / float64(totalPlayerLevels)
					//if xpScaler > 1 {
					xpVal = int(math.Ceil(float64(xpVal) * xpScaler))
					//}

					grantXP, xpScale := user.Character.GrantXP(xpVal)

					xpMsgExtra := ``
					if xpScale != 100 {
						xpMsgExtra = fmt.Sprintf(` <ansi fg="yellow">(%d%% scale)</ansi>`, xpScale)
					}

					user.SendText(
						fmt.Sprintf(xpMsg, grantXP, xpMsgExtra),
					)

					// Apply alignment changes
					alignmentBefore := user.Character.AlignmentName()
					alignmentAdj := combat.AlignmentChange(user.Character.Alignment, mob.Character.Alignment)
					user.Character.UpdateAlignment(alignmentAdj)
					alignmentAfter := user.Character.AlignmentName()

					slog.Debug("Alignment", "user Alignment", user.Character.Alignment, "mob Alignment", mob.Character.Alignment, `alignmentAdj`, alignmentAdj, `alignmentBefore`, alignmentBefore, `alignmentAfter`, alignmentAfter)

					if alignmentBefore != alignmentAfter {
						alignmentBefore = fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentBefore, alignmentBefore)
						alignmentAfter = fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentAfter, alignmentAfter)
						updateTxt := fmt.Sprintf(`<ansi fg="231">Your alignment has shifted from %s to %s!</ansi>`, alignmentBefore, alignmentAfter)
						user.SendText(updateTxt)
					}

					// Chance to learn to tame the creature.
					levelDelta := user.Character.Level - mob.Character.Level
					if levelDelta < 0 {
						levelDelta = 0
					}
					skillsDelta := int((float64(user.Character.Stats.Perception.ValueAdj-mob.Character.Stats.Perception.ValueAdj) + float64(user.Character.Stats.Smarts.ValueAdj-mob.Character.Stats.Smarts.ValueAdj)) / 2)
					if skillsDelta < 0 {
						skillsDelta = 0
					}
					targetNumber := levelDelta + skillsDelta
					if targetNumber < 1 {
						targetNumber = 1
					}

					slog.Info("Tame Chance", "levelDelta", levelDelta, "skillsDelta", skillsDelta, "targetNumber", targetNumber)

					if util.Rand(1000) < targetNumber {
						if mob.IsTameable() && user.Character.GetSkillLevel(skills.Tame) > 0 {

							currentSkill := user.Character.MobMastery.GetTame(int(mob.MobId))
							if currentSkill < 50 {
								user.Character.MobMastery.SetTame(int(mob.MobId), currentSkill+1)
								if currentSkill == -1 {
									user.SendText(fmt.Sprintf(`<ansi fg="magenta">***</ansi> You've learned how to tame a <ansi fg="mobname">%s</ansi>! <ansi fg="magenta">***</ansi>`, mob.Character.Name))
								} else {
									user.SendText(fmt.Sprintf(`<ansi fg="magenta">***</ansi> Your <ansi fg="mobname">%s</ansi> taming skills get a little better! <ansi fg="magenta">***</ansi>`, mob.Character.Name))
								}
							}

						}
					}

					continue
				}

				if _, ok := partyTracker[p.LeaderUserId]; !ok {
					partyTracker[p.LeaderUserId] = 0
				}
				partyTracker[p.LeaderUserId] += xpVal

			}
		}

	}

	if len(partyTracker) > 0 {

		xpMsg := `You gained <ansi fg="yellow-bold">%d experience points</ansi>%s!`
		for leaderId, xp := range partyTracker {
			if p := parties.Get(leaderId); p != nil {

				allMembers := p.GetMembers()
				xpSplit := xp / len(allMembers)

				slog.Info(`Party XP`, `totalXP`, xp, `splitXP`, xpSplit, `memberCt`, len(allMembers))

				for _, memberId := range allMembers {

					if user := users.GetByUserId(memberId); user != nil {

						if mob.Character.Zone != `Training` { // Don't track any kills in the training zone
							user.Character.KD.AddMobKill(int(mob.MobId))
						}

						grantXP, xpScale := user.Character.GrantXP(xpSplit)

						xpMsgExtra := ``
						if xpScale != 100 {
							xpMsgExtra = fmt.Sprintf(` <ansi fg="yellow">(%d%% scale)</ansi>`, xpScale)
						}

						user.SendText(
							fmt.Sprintf(xpMsg, grantXP, xpMsgExtra),
						)

						// Apply alignment changes
						alignmentBefore := user.Character.AlignmentName()
						alignmentAdj := combat.AlignmentChange(user.Character.Alignment, mob.Character.Alignment)
						user.Character.UpdateAlignment(alignmentAdj)
						alignmentAfter := user.Character.AlignmentName()

						slog.Debug("Alignment", "user Alignment", user.Character.Alignment, "mob Alignment", mob.Character.Alignment, `alignmentAdj`, alignmentAdj, `alignmentBefore`, alignmentBefore, `alignmentAfter`, alignmentAfter)

						if alignmentBefore != alignmentAfter {
							alignmentBefore = fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentBefore, alignmentBefore)
							alignmentAfter = fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentAfter, alignmentAfter)
							updateTxt := fmt.Sprintf(`<ansi fg="231">Your alignment has shifted from %s to %s!</ansi>`, alignmentBefore, alignmentAfter)
							user.SendText(updateTxt)
						}

						// Chance to learn to tame the creature.
						levelDelta := user.Character.Level - mob.Character.Level
						if levelDelta < 0 {
							levelDelta = 0
						}
						skillsDelta := int((float64(user.Character.Stats.Perception.ValueAdj-mob.Character.Stats.Perception.ValueAdj) + float64(user.Character.Stats.Smarts.ValueAdj-mob.Character.Stats.Smarts.ValueAdj)) / 2)
						if skillsDelta < 0 {
							skillsDelta = 0
						}
						targetNumber := levelDelta + skillsDelta
						if targetNumber < 1 {
							targetNumber = 1
						}

						slog.Info("Tame Chance", "levelDelta", levelDelta, "skillsDelta", skillsDelta, "targetNumber", targetNumber)

						if util.Rand(1000) < targetNumber {
							if mob.IsTameable() && user.Character.GetSkillLevel(skills.Tame) > 0 {

								currentSkill := user.Character.MobMastery.GetTame(int(mob.MobId))
								if currentSkill < 50 {
									user.Character.MobMastery.SetTame(int(mob.MobId), currentSkill+1)

									if currentSkill == -1 {
										user.SendText(fmt.Sprintf(`<ansi fg="magenta">***</ansi> You've learned how to tame a <ansi fg="mobname">%s</ansi>! <ansi fg="magenta">***</ansi>`, mob.Character.Name))
									} else {
										user.SendText(fmt.Sprintf(`<ansi fg="magenta">***</ansi> Your <ansi fg="mobname">%s</ansi> taming skills get a little better! <ansi fg="magenta">***</ansi>`, mob.Character.Name))
									}
								}

							}
						}
					}

				}

			}
		}

	}

	// Check for any dropped loot...
	for _, item := range mob.Character.Items {
		msg := fmt.Sprintf(`<ansi fg="item">%s</ansi> drops to the ground.`, item.DisplayName())
		room.SendText(msg)
		room.AddItem(item, false)
	}

	allWornItems := mob.Character.Equipment.GetAllItems()

	for _, item := range allWornItems {

		roll := util.Rand(100)

		util.LogRoll(`Drop Item`, roll, mob.ItemDropChance)

		if roll >= mob.ItemDropChance {
			continue
		}

		msg := fmt.Sprintf(`<ansi fg="item">%s</ansi> drops to the ground.`, item.DisplayName())
		room.SendText(msg)
		room.AddItem(item, false)
	}

	if mob.Character.Gold > 0 {
		msg := fmt.Sprintf(`<ansi fg="yellow-bold">%d gold</ansi> drops to the ground.`, mob.Character.Gold)
		room.SendText(msg)
		room.Gold += mob.Character.Gold
	}

	// Destroy any record of this mob.
	mobs.DestroyInstance(mob.InstanceId)

	// Clean up mob from room...
	if r := rooms.LoadRoom(mob.HomeRoomId); r != nil {
		r.CleanupMobSpawns(false)
	}

	// Remove from current room
	room.RemoveMob(mob.InstanceId)

	return true, nil
}
