package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/volte6/mud/auctions"
	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/colorpatterns"
	"github.com/volte6/mud/combat"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func (w *World) roundTick() {
	tStart := time.Now()

	c := configs.GetConfig()

	gdBefore := gametime.GetDate()

	util.IncrementRoundCount()
	roundNumber := util.GetRoundCount()

	gdNow := gametime.GetDate()

	if gdBefore.Night != gdNow.Night {
		if gdNow.Night {
			sunsetTxt, _ := templates.Process("generic/sunset", nil)

			events.AddToQueue(events.Broadcast{
				Text: sunsetTxt,
			})

		} else {
			sunriseTxt, _ := templates.Process("generic/sunrise", nil)

			events.AddToQueue(events.Broadcast{
				Text: sunriseTxt,
			})

		}
	}

	//
	// Disconnect players that have been inactive too long
	//
	w.HandleInactivePlayers(c.SecondsToRounds(int(c.MaxIdleSeconds)))

	//
	// Do auction maintenance
	//
	w.ProcessAuction(tStart)

	if roundNumber%100 == 0 {
		scripting.PruneVMs()
	}

	if c.LogIntervalRoundCount > 0 && roundNumber%uint64(c.LogIntervalRoundCount) == 0 {
		slog.Info("World::RoundTick()", "roundNumber", roundNumber)
	}

	if roundNumber%uint64(c.LootGoblinRoundCount) == 0 {
		if room := rooms.LoadRoom(rooms.GoblinRoom); room != nil { // loot goblin room
			slog.Info(`Loot Goblin Spawn`, `roundNumber`, roundNumber)
			room.Prepare(false) // Make sure the loot goblin spawns.
		}
	}

	//
	// Reduce existing hostility (if any)
	//
	mobs.ReduceHostility()

	//
	// Player round ticks
	//
	w.HandlePlayerRoundTicks()
	//
	// Player round ticks
	//
	w.HandleMobRoundTicks()

	//
	// Respawn any enemies that have been missing for too long
	//
	w.HandleRespawns()

	//
	// Combat rounds
	//
	affectedPlayers1, affectedMobs1 := w.HandlePlayerCombat()

	affectedPlayers2, affectedMobs2 := w.HandleMobCombat()

	// Do any resolution or extra checks based on everyone that has been involved in combat this round.
	w.HandleAffected(append(affectedPlayers1, affectedPlayers2...), append(affectedMobs1, affectedMobs2...))

	//
	// Healing
	//
	w.HandleAutoHealing(roundNumber)

	//
	// Idle mobs
	//
	w.HandleIdleMobs()

	//
	// Shadow/death realm
	//
	w.HandleShadowRealm(roundNumber)

	util.TrackTime(`World::RoundTick()`, time.Since(tStart).Seconds())
}

func (w *World) HandleInactivePlayers(maxIdleRounds int) {

	if maxIdleRounds == 0 {
		return
	}

	roundNumber := util.GetRoundCount()
	if roundNumber < uint64(maxIdleRounds) {
		return
	}

	cutoffRound := roundNumber - uint64(maxIdleRounds)

	for _, user := range users.GetAllActiveUsers() {
		li := user.GetLastInputRound()

		//slog.Info("HandleInactivePlayers", "roundNumber", roundNumber, "maxIdleRounds", maxIdleRounds, "cutoffRound", cutoffRound, "GetLastInputRound", li)

		if li == 0 {
			continue
		}

		if li-cutoffRound == 5 {
			user.SendText(`<ansi fg="203">WARNING:</ansi> <ansi fg="208">You are about to be kicked for inactivity!</ansi>`)
		}

		if li < cutoffRound {
			slog.Info(`Inactive Kick`, `userId`, user.UserId)
			w.Kick(user.UserId)
		}

	}

}

// Round ticks for players
func (w *World) HandlePlayerRoundTicks() {

	roomsWithPlayers := rooms.GetRoomsWithPlayers()
	for _, roomId := range roomsWithPlayers {
		// Get rooom
		if room := rooms.LoadRoom(roomId); room != nil {
			room.RoundTick()

			allowIdleMessages := true
			if handled, err := scripting.TryRoomIdleEvent(roomId); err == nil {
				if handled { // For this event, handled represents whether to reject the move.
					allowIdleMessages = false
				}
			}

			if allowIdleMessages {
				chanceIn100 := 5
				if room.RoomId == -1 {
					chanceIn100 = 20
				}

				idleMsgs := room.IdleMessages
				idleMsgCt := len(room.IdleMessages)
				if idleMsgCt > 0 && util.Rand(100) < chanceIn100 {

					if targetRoomId, err := strconv.Atoi(idleMsgs[0]); err == nil {
						idleMsgCt = 0
						if tgtRoom := rooms.LoadRoom(targetRoomId); tgtRoom != nil {
							idleMsgs = tgtRoom.IdleMessages
							idleMsgCt = len(idleMsgs)
						}
					}

					if idleMsgCt > 0 {
						// pick a random message
						idleMsgIndex := uint8(util.Rand(idleMsgCt))

						// If it's a repeating message, treat it as a non-message
						// (Unless it's the only one)
						if idleMsgIndex != room.LastIdleMessage || idleMsgCt == 1 {

							room.LastIdleMessage = idleMsgIndex

							msg := idleMsgs[idleMsgIndex]
							if msg != `` {
								room.SendText(msg)
							}

						}
					}

				}
			}

			for _, uId := range room.GetPlayers() {

				user := users.GetByUserId(uId)
				if user == nil {
					continue
				}

				// Roundtick any cooldowns
				user.Character.Cooldowns.RoundTick()

				if user.Character.Charmed != nil && user.Character.Charmed.RoundsRemaining > 0 {
					user.Character.Charmed.RoundsRemaining--
				}

				if triggeredBuffs := user.Character.Buffs.Trigger(); len(triggeredBuffs) > 0 {

					//
					// Fire onTrigger for buff script
					//
					for _, buff := range triggeredBuffs {
						if !buff.Expired() {
							scripting.TryBuffScriptEvent(`onTrigger`, uId, 0, buff.BuffId)
						}
					}

				}

				// Recalculate all stats at the end of the round tick
				user.Character.Validate()
			}

		}
	}

}

// Round ticks for players
func (w *World) HandleMobRoundTicks() {

	for _, mobInstanceId := range mobs.GetAllMobInstanceIds() {

		mob := mobs.GetInstance(mobInstanceId)

		if mob == nil {
			continue
		}

		// Roundtick any cooldowns
		mob.Character.Cooldowns.RoundTick()

		if mob.Character.Charmed != nil && mob.Character.Charmed.RoundsRemaining > 0 {
			mob.Character.Charmed.RoundsRemaining--
		}

		if triggeredBuffs := mob.Character.Buffs.Trigger(); len(triggeredBuffs) > 0 {

			//
			// Fire onTrigger for buff script
			//
			for _, buff := range triggeredBuffs {
				scripting.TryBuffScriptEvent(`onTrigger`, 0, mobInstanceId, buff.BuffId)
			}

		}

		// Do charm cleanup
		if mob.Character.IsCharmed() && mob.Character.Charmed.RoundsRemaining == 0 {
			cmd := mob.Character.Charmed.ExpiredCommand
			if charmedUserId := mob.Character.RemoveCharm(); charmedUserId > 0 {
				if charmedUser := users.GetByUserId(charmedUserId); charmedUser != nil {
					charmedUser.Character.TrackCharmed(mob.InstanceId, false)
				}
			}
			if cmd != `` {
				cmds := strings.Split(cmd, `;`)
				for _, cmd := range cmds {
					cmd = strings.TrimSpace(cmd)
					if len(cmd) > 0 {
						mob.Command(cmd)
					}
				}
			}
		}

		// Recalculate all stats at the end of the round tick
		mob.Character.Validate()

		if mob.Character.Health <= 0 {
			// Mob died
			mob.Command(`suicide`)
		}

	}

}

func (w *World) LogOff(userId int) {

	user := users.GetByUserId(userId)

	users.SaveUser(*user)

	worldManager.LeaveWorld(userId)

	connId := user.ConnectionId()

	tplTxt, _ := templates.Process("goodbye", nil, templates.AnsiTagsPreParse)

	worldManager.GetConnectionPool().SendTo([]byte(tplTxt), connId)

	if err := users.LogOutUserByConnectionId(connId); err != nil {
		slog.Error("Log Out Error", "connectionId", connId, "error", err)
	}

	worldManager.GetConnectionPool().Remove(connId)

}

func (w *World) PruneBuffs() {

	roomsWithPlayers := rooms.GetRoomsWithPlayers()
	for _, roomId := range roomsWithPlayers {
		// Get rooom
		if room := rooms.LoadRoom(roomId); room != nil {

			// Handle outstanding player buffs
			logOff := false
			for _, uId := range room.GetPlayers(rooms.FindBuffed) {

				user := users.GetByUserId(uId)

				logOff = false
				if buffsToPrune := user.Character.Buffs.Prune(); len(buffsToPrune) > 0 {
					for _, buffInfo := range buffsToPrune {
						scripting.TryBuffScriptEvent(`onEnd`, uId, 0, buffInfo.BuffId)

						if buffInfo.BuffId == 0 { // Log them out // logoff // logout
							logOff = true
						}
					}

					user.Character.Validate()

					if logOff {
						slog.Info("DOING LOGOFF")
						w.LogOff(uId)
					}
				}

			}
		}
	}

	// Handle outstanding mob buffs
	for _, mobInstanceId := range mobs.GetAllMobInstanceIds() {

		mob := mobs.GetInstance(mobInstanceId)

		if buffsToPrune := mob.Character.Buffs.Prune(); len(buffsToPrune) > 0 {
			for _, buffInfo := range buffsToPrune {
				scripting.TryBuffScriptEvent(`onEnd`, 0, mobInstanceId, buffInfo.BuffId)
			}

			mob.Character.Validate()
		}

	}

}

func (w *World) HandleRespawns() {

	//
	// Handle any respawns pending
	//
	for _, roomId := range rooms.GetRoomsWithPlayers() {

		// Get rooom
		room := rooms.LoadRoom(roomId)
		if room == nil {
			continue
		}

		for idx, spawnInfo := range room.SpawnInfo {

			if spawnInfo.InstanceId == 0 {

				if spawnInfo.CooldownLeft < 1 {
					// Spawn a new one.
					if spawnInfo.MobId > 0 {

						if mob := mobs.NewMobById(mobs.MobId(spawnInfo.MobId), room.RoomId); mob != nil {
							spawnInfo.InstanceId = mob.InstanceId
							room.AddMob(mob.InstanceId)
							room.SpawnInfo[idx] = spawnInfo

							if len(spawnInfo.Message) > 0 {
								room.SendText(spawnInfo.Message)
							}
						}
					}
				}

			}
		}
	}

}

// WHere combat happens
func (w *World) HandlePlayerCombat() (affectedPlayerIds []int, affectedMobInstanceIds []int) {

	tStart := time.Now()

	c := configs.GetConfig()

	for _, userId := range users.GetOnlineUserIds() {

		user := users.GetByUserId(userId)

		// If has a buff that prevents combat, skip the player
		if user.Character.HasBuffFlag(buffs.NoCombat) {
			continue
		}

		if user == nil || user.Character.Aggro == nil {
			continue
		}

		// Disable any buffs that are cancelled by combat
		user.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

		roomId := user.Character.RoomId

		uRoom := rooms.LoadRoom(roomId)
		if uRoom == nil {
			continue
		}

		if user.Character.Aggro.Type == characters.Flee {

			// Revert to Default combat regardless of outcome
			user.Character.SetAggro(user.Character.Aggro.UserId, user.Character.Aggro.MobInstanceId, characters.DefaultAttack)

			blockedByMob := ``
			for _, mobInstId := range uRoom.GetMobs(rooms.FindFighting) {
				if mob := mobs.GetInstance(mobInstId); mob != nil {
					if mob.Character.Aggro == nil || mob.Character.Aggro.UserId != userId {
						continue
					}

					// if equal, 20% chance of fleeing... at best, 40% chance
					chanceIn100 := int(float64(user.Character.Stats.Speed.ValueAdj) / (float64(user.Character.Stats.Speed.ValueAdj) + float64(mob.Character.Stats.Speed.ValueAdj)) * 40)
					roll := util.Rand(100)

					util.LogRoll(`Flee`, roll, chanceIn100)

					if roll >= chanceIn100 {
						blockedByMob = mob.Character.Name
						break
					}
				}
			}

			blockedByPlayer := ``
			blockedByPlayerId := 0
			for _, userId := range uRoom.GetPlayers(rooms.FindFighting) {
				if u := users.GetByUserId(userId); u != nil {
					if u.Character.Aggro == nil || u.Character.Aggro.UserId != userId {
						continue
					}

					// if equal, 20% chance of fleeing... at best, 40% chance
					chanceIn100 := int(float64(user.Character.Stats.Speed.ValueAdj) / (float64(user.Character.Stats.Speed.ValueAdj) + float64(u.Character.Stats.Speed.ValueAdj)) * 40)
					roll := util.Rand(100)

					util.LogRoll(`Flee`, roll, chanceIn100)

					if roll >= chanceIn100 {
						blockedByPlayer = u.Character.Name
						blockedByPlayerId = u.UserId
						break
					}
				}
			}

			if blockedByMob != `` {
				user.SendText(fmt.Sprintf(`<ansi fg="red-bold"><ansi fg="mobname">%s</ansi> blocks you from fleeing!</ansi>`, blockedByMob))
				uRoom.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is blocked from fleeing by <ansi fg="mobname">%s</ansi>!`, user.Character.Name, blockedByMob), user.UserId)
				continue
			}

			if blockedByPlayer != `` {
				user.SendText(fmt.Sprintf(`<ansi fg="red-bold"><ansi fg="username">%s</ansi> blocks you from fleeing!</ansi>`, blockedByPlayer))
				uRoom.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is blocked from fleeing by <ansi fg="username">%s</ansi>!`, user.Character.Name, blockedByPlayer), userId, blockedByPlayerId)
				continue
			}

			// Success!
			exitName, exitRoomId := uRoom.GetRandomExit()

			if exitRoomId == 0 {
				user.SendText(`You can't find an exit!`)
				continue
			}

			user.SendText(fmt.Sprintf(`You flee to the <ansi fg="exit">%s</ansi> exit!`, exitName))
			uRoom.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> flees to the <ansi fg="exit">%s</ansi> exit!`, user.Character.Name, exitName), userId)

			rooms.MoveToRoom(userId, exitRoomId)

			for _, instId := range uRoom.GetMobs(rooms.FindCharmed) {
				if mob := mobs.GetInstance(instId); mob != nil {
					// Charmed mobs assist
					if mob.Character.IsCharmed(userId) {
						mob.Command(exitName)
					}
				}
			}

			continue
		}

		if user.Character.Aggro != nil && user.Character.Aggro.Type == characters.SpellCast {

			if user.Character.Aggro.RoundsWaiting > 0 {
				user.Character.Aggro.RoundsWaiting--

				scripting.TrySpellScriptEvent(`onWait`, user.UserId, 0, user.Character.Aggro.SpellInfo)

				continue
			}

			roll := util.RollDice(1, 100)
			successChance := user.Character.GetBaseCastSuccessChance(user.Character.Aggro.SpellInfo.SpellId)
			if roll >= successChance {

				// fail
				user.SendText(fmt.Sprintf(`<ansi fg="magenta">***</ansi> Your spell fizzles! <ansi fg="magenta">***</ansi> (Rolled %d on %d%% chance of success)`, roll, successChance))
				uRoom.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to cast a spell but it <ansi fg="magenta">fizzles</ansi>!`, user.Character.Name), userId)
				user.Character.Aggro = nil

				continue

			}

			allowRetaliation := true
			if handled, err := scripting.TrySpellScriptEvent(`onMagic`, user.UserId, 0, user.Character.Aggro.SpellInfo); err == nil {
				if handled {
					allowRetaliation = false
				}
			}

			user.Character.TrackSpellCast(user.Character.Aggro.SpellInfo.SpellId)

			if allowRetaliation {
				if spellData := spells.GetSpell(user.Character.Aggro.SpellInfo.SpellId); spellData != nil {

					if spellData.Type == spells.HarmSingle || spellData.Type == spells.HarmMulti || spellData.Type == spells.HarmArea {

						for _, mobId := range user.Character.Aggro.SpellInfo.TargetMobInstanceIds {
							if defMob := mobs.GetInstance(mobId); defMob != nil {

								defMob.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

								if defMob.Character.Aggro == nil {
									defMob.PreventIdle = true
									defMob.Command(fmt.Sprintf("attack @%d", user.UserId)) // @ means player
								}
							}
						}

					}
				}
			}

			user.Character.Aggro = nil

			continue

		}

		// In combat with another player
		if user.Character.Aggro != nil && user.Character.Aggro.UserId > 0 {

			defUser := users.GetByUserId(user.Character.Aggro.UserId)

			uRoom := rooms.LoadRoom(roomId)

			if uRoom == nil {
				user.Character.Aggro = nil
				continue
			}

			targetFound := true
			if defUser == nil {
				targetFound = false
			} else if defUser.Character.RoomId != user.Character.RoomId {

				if user.Character.Aggro.ExitName == `` {
					targetFound = false
				} else {
					// If the exitId doesn't match the target room id, can't find em
					if _, exitRoomId := uRoom.FindExitByName(user.Character.Aggro.ExitName); exitRoomId != defUser.Character.RoomId {
						targetFound = false
					}
				}

			}

			if !targetFound {
				user.SendText(`Your target can't be found.`)
				user.Character.Aggro = nil
				continue
			}

			defRoom := rooms.LoadRoom(defUser.Character.RoomId)
			if defRoom == nil {
				user.Character.Aggro = nil
				continue
			}

			defUser.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

			if defUser.Character.Health < 1 {
				user.SendText(`Your rage subsides.`)
				user.Character.Aggro = nil
				continue
			}

			if user.Character.Aggro.RoundsWaiting > 0 {
				slog.Info(`RoundsWaiting`, `User`, user.Character.Name, `Rounds`, user.Character.Aggro.RoundsWaiting)

				user.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, user.Character, defUser.Character, combat.User, combat.User)

				for _, msg := range roundResult.MessagesToSource {
					user.SendText(msg)
				}

				for _, msg := range roundResult.MessagesToTarget {
					defUser.SendText(msg)
				}

				if len(roundResult.MessagesToSourceRoom) > 0 {
					for _, msg := range roundResult.MessagesToSourceRoom {
						uRoom.SendText(msg, user.UserId, defUser.UserId)
					}
				}

				if len(roundResult.MessagesToTargetRoom) > 0 {
					for _, msg := range roundResult.MessagesToTargetRoom {
						defRoom.SendText(msg, user.UserId, defUser.UserId)
					}
				}

				continue
			}

			// Can't see them, can't fight them.
			if defUser.Character.HasBuffFlag(buffs.Hidden) {
				user.SendText("You can't seem to find your target.")
				continue
			}

			affectedPlayerIds = append(affectedPlayerIds, user.Character.Aggro.UserId)

			var roundResult combat.AttackResult

			roundResult = combat.AttackPlayerVsPlayer(user, defUser)

			// If a mob attacks a player, check whether player has a charmed mob helping them, and if so, they will move to attack back
			room := rooms.LoadRoom(roomId)
			for _, instanceId := range room.GetMobs(rooms.FindCharmed) {
				if charmedMob := mobs.GetInstance(instanceId); charmedMob != nil {
					if charmedMob.Character.IsCharmed(defUser.UserId) && charmedMob.Character.Aggro == nil {

						// Set aggro to something to prevent multiple attack triggers on this conditional
						charmedMob.Character.Aggro = &characters.Aggro{
							Type: characters.DefaultAttack,
						}

						charmedMob.Command(fmt.Sprintf("attack @%d", user.UserId))

					}
				}
			}

			for _, buffId := range roundResult.BuffSource {

				events.AddToQueue(events.Buff{
					UserId:        user.UserId,
					MobInstanceId: 0,
					BuffId:        buffId,
				})

			}

			for _, buffId := range roundResult.BuffTarget {

				events.AddToQueue(events.Buff{
					UserId:        defUser.UserId,
					MobInstanceId: 0,
					BuffId:        buffId,
				})

			}

			for _, msg := range roundResult.MessagesToSource {
				user.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToTarget {
				defUser.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToSourceRoom {
				uRoom.SendText(msg, user.UserId, defUser.UserId)
			}

			for _, msg := range roundResult.MessagesToTargetRoom {
				defRoom.SendText(msg, user.UserId, defUser.UserId)
			}

			// If the attack connected, check for damage to equipment.
			if roundResult.Hit {
				// For now, only focus on offhand items.
				if defUser.Character.Equipment.Offhand.ItemId > 0 {

					modifier := 0
					if roundResult.Crit { // Crits double the chance of breakage for offhand items.
						modifier = int(defUser.Character.Equipment.Offhand.GetSpec().BreakChance)
					}

					if defUser.Character.Equipment.Offhand.BreakTest(modifier) {
						// Send message about the break

						defUser.SendText(`<ansi fg="202">***</ansi>`)
						defUser.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> Your <ansi fg="item">%s</ansi> breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple()))
						defUser.SendText(`<ansi fg="202">***</ansi>`)

						defRoom.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="username">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple(), defUser.Character.Name), defUser.UserId)

						defUser.Character.RemoveFromBody(defUser.Character.Equipment.Offhand)
						itm := items.New(20) // Broken item
						if !defUser.Character.StoreItem(itm) {
							room.AddItem(itm, false)
						}
					}
				}
			}

			if user.Character.Health <= 0 || defUser.Character.Health <= 0 {
				defUser.Character.EndAggro()
				user.Character.EndAggro()
			} else {
				user.Character.SetAggro(defUser.UserId, 0, characters.DefaultAttack)
			}

		}

		// In combat with a mob
		if user.Character.Aggro != nil && user.Character.Aggro.MobInstanceId > 0 {

			affectedMobInstanceIds = append(affectedMobInstanceIds, user.Character.Aggro.MobInstanceId)

			defMob := mobs.GetInstance(user.Character.Aggro.MobInstanceId)

			targetFound := true
			if defMob == nil {
				targetFound = false
			} else if defMob.Character.RoomId != user.Character.RoomId {

				if user.Character.Aggro.ExitName == `` {
					targetFound = false
				} else {
					// Make sure the target is still at the exit

					uRoom := rooms.LoadRoom(roomId)
					if uRoom == nil {
						user.Character.Aggro = nil
						continue
					}

					// If the exitId doesn't match the target room id, can't find em
					if _, exitRoomId := uRoom.FindExitByName(user.Character.Aggro.ExitName); exitRoomId != defMob.Character.RoomId {
						targetFound = false
					}

				}

			}

			if !targetFound {
				user.SendText("Your target can't be found.")
				user.Character.Aggro = nil
				continue
			}

			defRoom := rooms.LoadRoom(defMob.Character.RoomId)

			defMob.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

			if defMob.Character.Health < 1 {
				user.SendText("Your rage subsides.")
				user.Character.Aggro = nil
				continue
			}

			if user.Character.Aggro.RoundsWaiting > 0 {
				slog.Info(`RoundsWaiting`, `User`, user.Character.Name, `Rounds`, user.Character.Aggro.RoundsWaiting)

				user.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, user.Character, &defMob.Character, combat.User, combat.Mob)

				for _, msg := range roundResult.MessagesToSource {
					user.SendText(msg)
				}

				for _, msg := range roundResult.MessagesToSourceRoom {
					uRoom.SendText(msg, user.UserId)
				}

				for _, msg := range roundResult.MessagesToTargetRoom {
					defRoom.SendText(msg, user.UserId)
				}

				continue
			}

			// Can't see them, can't fight them.
			if defMob.Character.HasBuffFlag(buffs.Hidden) {
				user.SendText("You can't seem to find your target.")
				continue
			}

			affectedPlayerIds = append(affectedPlayerIds, user.Character.Aggro.UserId)

			var roundResult combat.AttackResult

			roundResult = combat.AttackPlayerVsMob(user, defMob)

			for _, buffId := range roundResult.BuffSource {

				events.AddToQueue(events.Buff{
					UserId:        user.UserId,
					MobInstanceId: 0,
					BuffId:        buffId,
				})

			}

			for _, buffId := range roundResult.BuffTarget {

				events.AddToQueue(events.Buff{
					UserId:        0,
					MobInstanceId: defMob.InstanceId,
					BuffId:        buffId,
				})

			}

			for _, msg := range roundResult.MessagesToSource {
				user.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToSourceRoom {
				uRoom.SendText(msg, user.UserId)
			}

			for _, msg := range roundResult.MessagesToTargetRoom {
				defRoom.SendText(msg, user.UserId)
			}

			// Handle any scripted behavior now.
			if roundResult.Hit {
				scripting.TryMobScriptEvent(`onHurt`, defMob.InstanceId, user.UserId, `user`, map[string]any{`damage`: roundResult.DamageToTarget, `crit`: roundResult.Crit})
			}

			//
			// Special mob-only reaction/behavior
			//
			// Hostility default to 5 minutes
			for _, groupName := range defMob.Groups {
				mobs.MakeHostile(groupName, user.UserId, c.MinutesToRounds(2)-user.Character.Stats.Perception.ValueAdj)
			}

			// Mobs get aggro when attacked
			if defMob.Character.Aggro == nil {
				defMob.PreventIdle = true
				// If not in the same room,
				// find an exit to the room of the player to move to
				if user.Character.RoomId != defMob.Character.RoomId {
					if mobRoom := rooms.LoadRoom(defMob.Character.RoomId); mobRoom != nil {
						for exitName, exitInfo := range mobRoom.Exits {
							if exitInfo.RoomId == user.Character.RoomId {
								defMob.Command(fmt.Sprintf(`go %s`, exitName))
								if actionStr := defMob.GetAngryCommand(); actionStr != `` {
									defMob.Command(actionStr)
								}
								break
							}
						}
					}
				}

				defMob.Command(fmt.Sprintf("attack @%d", user.UserId)) // @ means player
			}

			if user.Character.Health <= 0 || defMob.Character.Health <= 0 {
				defMob.Character.EndAggro()
				user.Character.EndAggro()
			} else {
				user.Character.SetAggro(0, defMob.InstanceId, characters.DefaultAttack)
			}

		}

	}

	util.TrackTime(`World::HandlePlayerCombat()`, time.Since(tStart).Seconds())

	return affectedPlayerIds, affectedMobInstanceIds
}

// Mob combat operations may happen when players are not present.
func (w *World) HandleMobCombat() (affectedPlayerIds []int, affectedMobInstanceIds []int) {

	c := configs.GetConfig()

	tStart := time.Now()

	// Handle mob round of combat
	for _, mobId := range mobs.GetAllMobInstanceIds() {

		mob := mobs.GetInstance(mobId)

		// Only handling combat functions here, so ditch out if not in combat
		if mob == nil || mob.Character.Aggro == nil {
			continue
		}

		// If has a buff that prevents combat, skip the player
		if mob.Character.HasBuffFlag(buffs.NoCombat) {
			continue
		}

		mobRoom := rooms.LoadRoom(mob.Character.RoomId)

		if mobRoom == nil {
			mob.Character.Aggro = nil
			continue
		}

		// Disable any buffs that are cancelled by combat
		mob.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

		if mob.Character.Aggro != nil && mob.Character.Aggro.Type == characters.SpellCast {

			if mob.Character.Aggro.RoundsWaiting > 0 {
				mob.Character.Aggro.RoundsWaiting--

				scripting.TrySpellScriptEvent(`onWait`, 0, mob.InstanceId, mob.Character.Aggro.SpellInfo)

				continue
			}

			successChance := mob.Character.GetBaseCastSuccessChance(mob.Character.Aggro.SpellInfo.SpellId)
			if util.RollDice(1, 100) >= successChance {

				// fail
				mobRoom.SendText(fmt.Sprintf(`<ansi fg="mobnamme">%s</ansi> tries to cast a spell but it <ansi fg="magenta">fizzles</ansi>!`, mob.Character.Name))
				mob.Character.Aggro = nil

				continue

			}

			allowRetaliation := true
			if handled, err := scripting.TrySpellScriptEvent(`onMagic`, 0, mob.InstanceId, mob.Character.Aggro.SpellInfo); err == nil {
				if handled {
					allowRetaliation = false
				}
			}

			if allowRetaliation {
				if spellData := spells.GetSpell(mob.Character.Aggro.SpellInfo.SpellId); spellData != nil {

					if spellData.Type == spells.HarmSingle || spellData.Type == spells.HarmMulti || spellData.Type == spells.HarmArea {

						for _, mobId := range mob.Character.Aggro.SpellInfo.TargetMobInstanceIds {
							if defMob := mobs.GetInstance(mobId); defMob != nil {

								defMob.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

								if defMob.Character.Aggro == nil {
									defMob.PreventIdle = true
									defMob.Command(fmt.Sprintf("attack #%d", mob.InstanceId)) // # means mob
								}
							}
						}

					}
				}
			}

			mob.Character.Aggro = nil

			continue

		}

		// H2H is the base level combat, can do combat commands then
		if mob.Character.Aggro.Type == characters.DefaultAttack {

			// If they have idle commands, maybe do one of them?
			cmdCt := len(mob.CombatCommands)
			if cmdCt > 0 {

				// Each mob has a 10% chance of doing an idle action.
				if util.Rand(10) < mob.ActivityLevel {

					combatAction := mob.CombatCommands[util.Rand(cmdCt)]

					if combatAction == `` { // blank is a no-op
						continue
					}

					allCmds := strings.Split(combatAction, `;`)
					if len(allCmds) >= c.TurnsPerRound() {
						mob.Command(`say I have a CombatAction that is too long. Please notify an admin.`)
					} else {
						for turnDelay, action := range strings.Split(combatAction, `;`) {
							mob.Command(action, turnDelay)
						}
					}
					continue
				}
			}

		}
		roomId := mob.Character.RoomId

		affectedMobInstanceIds = append(affectedMobInstanceIds, mob.InstanceId)

		// mob attacks player
		if mob.Character.Aggro != nil && mob.Character.Aggro.UserId > 0 {

			defUser := users.GetByUserId(mob.Character.Aggro.UserId)
			if defUser == nil || mob.Character.RoomId != defUser.Character.RoomId {
				mob.Character.Aggro = nil
				continue
			}

			defRoom := rooms.LoadRoom(defUser.Character.RoomId)
			if defRoom == nil {
				mob.Character.Aggro = nil
				continue
			}

			defUser.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

			if defUser.Character.Health < 1 {
				mob.Character.Aggro = nil
				continue
			}

			// Can't see them, can't fight them.
			if defUser.Character.HasBuffFlag(buffs.Hidden) {
				continue
			}

			affectedPlayerIds = append(affectedPlayerIds, mob.Character.Aggro.UserId)

			// If no weapon but has stuff in the backpack, look for a weapon
			// Especially useful for when they get disarmed
			if mob.Character.Equipment.Weapon.ItemId == 0 && len(mob.Character.Items) > 0 {

				roll := util.Rand(100)

				util.LogRoll(`Look for weapon`, roll, mob.Character.Stats.Perception.ValueAdj)

				if roll < mob.Character.Stats.Perception.ValueAdj {
					possibleWeapons := []string{}
					for _, itm := range mob.Character.Items {
						iSpec := itm.GetSpec()
						if iSpec.Type == items.Weapon {
							possibleWeapons = append(possibleWeapons, itm.DisplayName())
						}
					}

					if len(possibleWeapons) > 0 {
						mob.Command(fmt.Sprintf("equip %s", possibleWeapons[util.Rand(len(possibleWeapons))]))
					}

				}
			}

			if mob.Character.Aggro.RoundsWaiting > 0 {
				slog.Info(`RoundsWaiting`, `User`, mob.Character.Name, `Rounds`, mob.Character.Aggro.RoundsWaiting)

				mob.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, &mob.Character, defUser.Character, combat.Mob, combat.User)

				for _, msg := range roundResult.MessagesToTarget {
					defUser.SendText(msg)
				}

				for _, msg := range roundResult.MessagesToSourceRoom {
					mobRoom.SendText(msg, defUser.UserId)
				}

				for _, msg := range roundResult.MessagesToTargetRoom {
					defRoom.SendText(msg, defUser.UserId)
				}

				continue
			}

			var roundResult combat.AttackResult

			roundResult = combat.AttackMobVsPlayer(mob, defUser)

			// If a mob attacks a player, check whether player has a charmed mob helping them, and if so, they will move to attack back
			room := rooms.LoadRoom(roomId)
			for _, instanceId := range room.GetMobs(rooms.FindCharmed) {
				if charmedMob := mobs.GetInstance(instanceId); charmedMob != nil {
					if charmedMob.Character.IsCharmed(defUser.UserId) && charmedMob.Character.Aggro == nil {
						// This is set to prevent it from triggering more than once
						charmedMob.Character.Aggro = &characters.Aggro{
							Type: characters.DefaultAttack,
						}

						charmedMob.Command(fmt.Sprintf("attack #%d", mob.InstanceId))

					}
				}
			}

			for _, buffId := range roundResult.BuffSource {

				events.AddToQueue(events.Buff{
					UserId:        0,
					MobInstanceId: mob.InstanceId,
					BuffId:        buffId,
				})

			}

			for _, buffId := range roundResult.BuffTarget {

				events.AddToQueue(events.Buff{
					UserId:        defUser.UserId,
					MobInstanceId: 0,
					BuffId:        buffId,
				})

			}

			for _, msg := range roundResult.MessagesToTarget {
				defUser.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToSourceRoom {
				mobRoom.SendText(msg, defUser.UserId)
			}

			for _, msg := range roundResult.MessagesToTargetRoom {
				defRoom.SendText(msg, defUser.UserId)
			}

			// If the attack connected, check for damage to equipment.
			if roundResult.Hit {

				// For now, only focus on offhand items.
				if defUser.Character.Equipment.Offhand.ItemId > 0 {

					modifier := 0
					if roundResult.Crit { // Crits double the chance of breakage for offhand items.
						modifier = int(defUser.Character.Equipment.Offhand.GetSpec().BreakChance)
					}

					if defUser.Character.Equipment.Offhand.BreakTest(modifier) {
						// Send message about the break

						defUser.SendText(`<ansi fg="202">***</ansi>`)
						defUser.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> Your <ansi fg="item">%s</ansi> breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple()))
						defUser.SendText(`<ansi fg="202">***</ansi>`)

						defRoom.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="username">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple(), defUser.Character.Name), defUser.UserId)

						defUser.Character.RemoveFromBody(defUser.Character.Equipment.Offhand)
						itm := items.New(20) // Broken item
						if !defUser.Character.StoreItem(itm) {
							room.AddItem(itm, false)
						}
					}
				}
			}

			if mob.Character.Health <= 0 || defUser.Character.Health <= 0 {
				mob.Character.EndAggro()
				defUser.Character.EndAggro()
			} else {
				mob.Character.SetAggro(defUser.UserId, 0, characters.DefaultAttack)
			}
		}

		// mob attacks mob
		if mob.Character.Aggro != nil && mob.Character.Aggro.MobInstanceId > 0 {

			affectedMobInstanceIds = append(affectedMobInstanceIds, mob.Character.Aggro.MobInstanceId)

			defMob := mobs.GetInstance(mob.Character.Aggro.MobInstanceId)

			if defMob == nil || mob.Character.RoomId != defMob.Character.RoomId {
				mob.Character.Aggro = nil
				continue
			}

			defRoom := rooms.LoadRoom(defMob.Character.RoomId)

			defMob.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

			if defMob.Character.Health < 1 {
				mob.Character.Aggro = nil
				continue
			}

			if mob.Character.Aggro.RoundsWaiting > 0 {
				slog.Info(`RoundsWaiting`, `User`, mob.Character.Name, `Rounds`, mob.Character.Aggro.RoundsWaiting)

				mob.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, &mob.Character, &defMob.Character, combat.Mob, combat.Mob)

				for _, msg := range roundResult.MessagesToSourceRoom {
					mobRoom.SendText(msg)
				}

				for _, msg := range roundResult.MessagesToTargetRoom {
					defRoom.SendText(msg)
				}

				continue
			}

			// Can't see them, can't fight them.
			if defMob.Character.HasBuffFlag(buffs.Hidden) {
				continue
			}

			var roundResult combat.AttackResult

			roundResult = combat.AttackMobVsMob(mob, defMob)

			for _, buffId := range roundResult.BuffSource {

				events.AddToQueue(events.Buff{
					UserId:        0,
					MobInstanceId: mob.InstanceId,
					BuffId:        buffId,
				})

			}

			for _, buffId := range roundResult.BuffTarget {

				events.AddToQueue(events.Buff{
					UserId:        0,
					MobInstanceId: defMob.InstanceId,
					BuffId:        buffId,
				})
			}

			for _, msg := range roundResult.MessagesToSourceRoom {
				mobRoom.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToTargetRoom {
				defRoom.SendText(msg)
			}

			// Handle any scripted behavior now.
			if roundResult.Hit {
				scripting.TryMobScriptEvent(`onHurt`, defMob.InstanceId, mob.InstanceId, `mob`, map[string]any{`damage`: roundResult.DamageToTarget, `crit`: roundResult.Crit})
			}

			// Mobs get aggro when attacked
			if defMob.Character.Aggro == nil {
				defMob.PreventIdle = true
				defMob.Character.Aggro = &characters.Aggro{
					Type: characters.DefaultAttack,
				}
				defMob.Command(fmt.Sprintf("attack #%d", mob.InstanceId)) // # means mob
			}

			// If the attack connected, check for damage to equipment.
			if roundResult.Hit {
				// For now, only focus on offhand items.
				if defMob.Character.Equipment.Offhand.ItemId > 0 {

					modifier := 0
					if roundResult.Crit { // Crits double the chance of breakage for offhand items.
						modifier = int(defMob.Character.Equipment.Offhand.GetSpec().BreakChance)
					}

					if defMob.Character.Equipment.Offhand.BreakTest(modifier) {
						// Send message about the break

						if defRoom := rooms.LoadRoom(defMob.Character.RoomId); defRoom != nil {

							defRoom.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="mobname">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, defMob.Character.Equipment.Offhand.NameSimple(), defMob.Character.Name))

							defMob.Character.RemoveFromBody(defMob.Character.Equipment.Offhand)
							itm := items.New(20) // Broken item
							if !defMob.Character.StoreItem(itm) {
								defRoom.AddItem(itm, false)
							}
						}
					}
				}
			}

			if mob.Character.Health <= 0 || defMob.Character.Health <= 0 {
				mob.Character.EndAggro()
				defMob.Character.EndAggro()
			} else {
				mob.Character.SetAggro(0, defMob.InstanceId, characters.DefaultAttack)
			}

		}

	}

	util.TrackTime(`World::HandleMobCombat()`, time.Since(tStart).Seconds())

	return affectedPlayerIds, affectedMobInstanceIds
}

func (w *World) HandleAffected(affectedPlayerIds []int, affectedMobInstanceIds []int) {

	playersHandled := map[int]struct{}{}
	for _, userId := range affectedPlayerIds {
		if _, ok := playersHandled[userId]; ok {
			continue
		}
		playersHandled[userId] = struct{}{}

		if user := users.GetByUserId(userId); user != nil {

			if user.Character.Health <= -10 {
				user.Command(`suicide`) // suicide drops all money/items and transports to land of the dead.
			} else if user.Character.Health < 1 {

				user.SendText(`<ansi fg="red">you drop to the ground!</ansi>`)

				if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> <ansi fg="red">drops to the ground!</ansi>`, user.Character.Name),
						user.UserId)
				}

			}
		}
	}

	mobsHandled := map[int]struct{}{}
	for _, mobId := range affectedMobInstanceIds {
		if _, ok := mobsHandled[mobId]; ok {
			continue
		}
		mobsHandled[mobId] = struct{}{}

		if mob := mobs.GetInstance(mobId); mob != nil {
			if mob.Character.Health < 1 {

				mob.Command(`suicide`)

			}
		}

	}

}

// Idle Mobs
func (w *World) HandleIdleMobs() {

	// c := configs.GetConfig()

	maxBoredom := uint8(configs.GetConfig().MaxMobBoredom)

	// Handle idle mob behavior
	tStart := time.Now()
	for _, mobId := range mobs.GetAllMobInstanceIds() {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		if mob.BoredomCounter >= maxBoredom {
			if mob.Despawns() {
				mob.Command(`despawn` + fmt.Sprintf(` depression %d/%d`, mob.BoredomCounter, maxBoredom))
			} else {
				mob.BoredomCounter = 0
			}
			continue
		}

		// If idle prevented, it's a one round interrupt (until another comes along)
		if mob.PreventIdle {
			mob.PreventIdle = false
			continue
		}

		// If they are doing some sort of combat thing,
		// Don't do idle actions
		if mob.Character.Aggro != nil {
			if mob.Character.Aggro.UserId > 0 {
				user := users.GetByUserId(mob.Character.Aggro.UserId)
				if user == nil || user.Character.RoomId != mob.Character.RoomId {
					mob.Command(`emote mumbles about losing their quarry.`)
					mob.Character.Aggro = nil
				}
			}
			continue
		}

		// If they have idle commands, maybe do one of them?
		handled, _ := scripting.TryMobScriptEvent("onIdle", mob.InstanceId, 0, ``, nil)
		if !handled {
			if !mob.Character.IsCharmed() { // Won't do this stuff if befriended

				if mob.MaxWander > -1 && len(mob.RoomStack) > mob.MaxWander {
					mob.GoingHome = true
				}
				if mob.GoingHome {
					mob.Command(`go home`)
					continue
				}

			}
		}

		//
		// Look for trouble
		//
		if !handled {
			if mob.Character.IsCharmed() {
				// Only some mobs can apply first aid
				if mob.Character.KnowsFirstAid() {
					mob.Command(`lookforaid`)
				}
			} else {

				idleCmd := `lookfortrouble`
				if util.Rand(10) < mob.ActivityLevel {
					idleCmd = mob.GetIdleCommand()
					if idleCmd == `` {
						idleCmd = `lookfortrouble`
					}
				}
				mob.Command(idleCmd)
			}
		}

	}

	util.TrackTime(`HandleIdleMobs()`, time.Since(tStart).Seconds())

}

// Healing
func (w *World) HandleAutoHealing(roundNumber uint64) {

	// Every 3 rounds.
	if roundNumber%3 != 0 {
		return
	}

	onlineIds := users.GetOnlineUserIds()
	for _, userId := range onlineIds {
		user := users.GetByUserId(userId)

		// Only heal if not in combat
		if user.Character.Aggro != nil {
			continue
		}

		if user.Character.Health < 1 {
			if user.Character.RoomId == 75 {
				if user.Character.Health < user.Character.HealthMax.Value {
					user.Character.Health++
				}
			} else {
				if user.Character.Health <= -10 {

					user.Command(`suicide`) // suicide drops all money/items and transports to land of the dead.

				} else {
					user.Character.Health--
					user.SendText(`<ansi fg="red">you are bleeding out!</ansi>`)
					if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
						room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is <ansi fg="red">bleeding out</ansi>! Somebody needs to provide aid!`, user.Character.Name), userId)
					}
				}
			}
		} else {

			if user.Character.Health > 0 || user.Character.RoomId == 75 {
				healingFactor := 1
				if user.Character.RoomId == 75 {
					healingFactor = 5
				}

				user.Character.Heal(
					//1*healingFactor, 1*healingFactor,
					user.Character.HealthPerRound()*healingFactor,
					user.Character.ManaPerRound()*healingFactor,
				)

			}
		}

		newcmdprompt := user.GetCommandPrompt(true)
		oldcmdprompt := user.GetTempData(`cmdprompt`)

		// If the prompt hasn't changed, skip redrawing
		if oldcmdprompt != nil && oldcmdprompt.(string) == newcmdprompt {
			continue
		}

		// save the new prompt for next time we want to check
		user.SetTempData(`cmdprompt`, newcmdprompt)

		if w.connectionPool.IsWebsocket(user.ConnectionId()) {
			w.connectionPool.SendTo([]byte(newcmdprompt), user.ConnectionId())
		} else {
			w.connectionPool.SendTo([]byte(templates.AnsiParse(newcmdprompt)), user.ConnectionId())
		}

	}

}

// Special shadow realm stuff
func (w *World) HandleShadowRealm(roundNumber uint64) {

	if roundNumber%uint64(configs.GetConfig().MinutesToRounds(1)) == 0 {

		// room 75 is the Shadow Realm
		deadRoom := rooms.LoadRoom(75)

		players := deadRoom.GetPlayers()
		if len(players) > 0 {

			type TemporaryRoomExit struct {
				RoomId  int       // Where does it lead to?
				Title   string    // Does this exist have a special title?
				UserId  int       // Who created it?
				Expires time.Time // When will it be auto-cleaned up?
			}

			tmpExit := rooms.TemporaryRoomExit{
				RoomId:  1,
				Title:   colorpatterns.ApplyColorPattern(`shimmering portal`, `cyan`),
				UserId:  0,
				Expires: time.Now().Add(time.Second * 10),
			}
			// Spawn a portal in the room that leads to the portal location
			deadRoom.AddTemporaryExit("shimmering portal", tmpExit)
			deadRoom.SendText(`<ansi fg="magenta-bold">A shimmering portal appears in the room.</ansi>`)
		}
	}

}

// Handle dropped players
func (w *World) HandleDroppedPlayers(droppedPlayers []int) {

	if len(droppedPlayers) == 0 {
		return
	}

	for _, userId := range droppedPlayers {
		if user := users.GetByUserId(userId); user != nil {

			user.SendText(`<ansi fg="red">you drop to the ground!</ansi>`)

			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> <ansi fg="red">drops to the ground!</ansi>`, user.Character.Name),
					user.UserId)
			}
		}
	}

	return
}

// Levelups
func (w *World) CheckForLevelUps() {

	onlineIds := users.GetOnlineUserIds()
	for _, userId := range onlineIds {
		user := users.GetByUserId(userId)

		for {

			if newLevel, statsDelta := user.Character.LevelUp(); newLevel {

				livesBefore := user.Character.ExtraLives

				c := configs.GetConfig()
				if c.PermaDeath && c.LevelUpLives > 0 {
					user.Character.ExtraLives += int(c.LevelUpLives)
					if user.Character.ExtraLives > int(c.MaxLives) {
						user.Character.ExtraLives = int(c.MaxLives)
					}
				}

				levelUpData := map[string]interface{}{
					"level":          user.Character.Level,
					"statsDelta":     statsDelta,
					"trainingPoints": 1,
					"statPoints":     1,
					"livesUp":        user.Character.ExtraLives - livesBefore,
				}
				levelUpStr, _ := templates.Process("character/levelup", levelUpData)

				user.SendText(levelUpStr)

				events.AddToQueue(events.Broadcast{
					Text: fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> <ansi fg="yellow">has leveled up to level %d!</ansi> <ansi fg="magenta-bold">***</ansi>%s`, user.Character.Name, user.Character.Level, term.CRLFStr),
				})

				if user.Character.Level >= 5 {
					for _, mobInstanceId := range user.Character.CharmedMobs {
						if mob := mobs.GetInstance(mobInstanceId); mob != nil {

							if mob.MobId == 38 {
								mob.Command(`say I see you have grown much stronger and more experienced. My assistance is now needed elsewhere. I wish you good luck!`)
								mob.Command(`emote clicks their heels together and disappears in a cloud of smoke.`, 10)
								mob.Command(`suicide vanish`, 10)
							}
						}
					}
				}

				users.SaveUser(*user)

				continue
			}

			break

		}

	}

}

// Checks for current auction and handles updates/communication
func (w *World) ProcessAuction(tNow time.Time) {

	a := auctions.GetCurrentAuction()
	if a == nil {
		return
	}

	c := configs.GetConfig()

	if a.IsEnded() {

		auctions.EndAuction()

		a.LastUpdate = tNow
		auctionTxt, _ := templates.Process("auctions/auction-end", a)

		for _, uid := range users.GetOnlineUserIds() {
			if u := users.GetByUserId(uid); u != nil {
				auctionOn := u.GetConfigOption(`auction`)
				if auctionOn == nil || auctionOn.(bool) {
					u.SendText(auctionTxt)
				}
			}
		}

		// Give the item to the winner and let them know
		if a.HighestBidUserId > 0 {

			if user := users.GetByUserId(a.HighestBidUserId); user != nil {
				if user.Character.StoreItem(a.ItemData) {
					msg := fmt.Sprintf(`<ansi fg="yellow">You have won the auction for the <ansi fg="item">%s</ansi>! It has been added to your backpack.</ansi>%s`, a.ItemData.DisplayName(), term.CRLFStr)
					user.SendText(msg)
				}
			}

		} else {
			if user := users.GetByUserId(a.SellerUserId); user != nil {
				if user.Character.StoreItem(a.ItemData) {
					msg := fmt.Sprintf(`<ansi fg="yellow">The auction for the <ansi fg="item">%s</ansi> has ended without a winner. It has been returned to you.</ansi>%s`, a.ItemData.DisplayName(), term.CRLFStr)
					user.SendText(msg)
				}
			}
		}

	} else if a.LastUpdate.IsZero() {

		a.LastUpdate = tNow
		auctionTxt, _ := templates.Process("auctions/auction-start", a)

		for _, uid := range users.GetOnlineUserIds() {
			if u := users.GetByUserId(uid); u != nil {
				auctionOn := u.GetConfigOption(`auction`)
				if auctionOn == nil || auctionOn.(bool) {
					u.SendText(auctionTxt)
				}
			}
		}

	} else if time.Since(a.LastUpdate) > time.Second*time.Duration(c.AuctionUpdateSeconds) {

		a.LastUpdate = tNow
		auctionTxt, _ := templates.Process("auctions/auction-update", a)

		for _, uid := range users.GetOnlineUserIds() {
			if u := users.GetByUserId(uid); u != nil {
				auctionOn := u.GetConfigOption(`auction`)
				if auctionOn == nil || auctionOn.(bool) {
					u.SendText(auctionTxt)
				}
			}
		}

	}

}
