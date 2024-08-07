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
	"github.com/volte6/mud/combat"
	"github.com/volte6/mud/configs"
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
			w.Broadcast(sunsetTxt)
		} else {
			sunriseTxt, _ := templates.Process("generic/sunrise", nil)
			w.Broadcast(sunriseTxt)
		}
	}

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

	messageQueue := util.NewMessageQueue(0, 0)

	//
	// Player round ticks
	//
	messageQueue.AbsorbMessages(w.HandlePlayerRoundTicks())
	//
	// Player round ticks
	//
	messageQueue.AbsorbMessages(w.HandleMobRoundTicks())

	//
	// Respawn any enemies that have been missing for too long
	//
	messageQueue.AbsorbMessages(w.HandleRespawns())

	//
	// Combat rounds
	//
	msgQ, affectedPlayers1, affectedMobs1 := w.HandlePlayerCombat()
	messageQueue.AbsorbMessages(msgQ)

	msgQ, affectedPlayers2, affectedMobs2 := w.HandleMobCombat()
	messageQueue.AbsorbMessages(msgQ)

	// Do any resolution or extra checks based on everyone that has been involved in combat this round.
	msgQ = w.HandleAffected(append(affectedPlayers1, affectedPlayers2...), append(affectedMobs1, affectedMobs2...))
	messageQueue.AbsorbMessages(msgQ)

	//
	// Healing
	//
	msgQ = w.HandleAutoHealing(roundNumber)
	messageQueue.AbsorbMessages(msgQ)

	//
	// Prune buffs - happens at the end of the round
	//
	// This is now handled in the tick loop
	// messageQueue.AbsorbMessages(w.PruneBuffs())

	//
	// Idle mobs
	//
	messageQueue.AbsorbMessages(w.HandleIdleMobs())

	//
	// Shadow/death realm
	//
	messageQueue.AbsorbMessages(w.HandleShadowRealm(roundNumber))

	if messageQueue.Pending() {
		w.DispatchMessages(messageQueue)
	}

	util.TrackTime(`World::RoundTick()`, time.Since(tStart).Seconds())
}

// Round ticks for players
func (w *World) HandlePlayerRoundTicks() util.MessageQueue {

	messageQueue := util.NewMessageQueue(0, 0)
	roomsWithPlayers := rooms.GetRoomsWithPlayers()
	for _, roomId := range roomsWithPlayers {
		// Get rooom
		if room := rooms.LoadRoom(roomId); room != nil {
			room.RoundTick()

			allowIdleMessages := true
			if scriptResponse, err := scripting.TryRoomIdleEvent(roomId, w); err == nil {
				messageQueue.AbsorbMessages(scriptResponse)
				if scriptResponse.Handled { // For this event, handled represents whether to reject the move.
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
								messageQueue.SendRoomMessage(roomId,
									msg,
									true)
							}

						}
					}

				}
			}

			for _, uId := range room.GetPlayers() {

				user := users.GetByUserId(uId)

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
							if response, err := scripting.TryBuffScriptEvent(`onTrigger`, uId, 0, buff.BuffId, w); err == nil {
								messageQueue.AbsorbMessages(response)
							}
						}
					}

				}

				// Recalculate all stats at the end of the round tick
				user.Character.Validate()
			}

		}
	}

	return messageQueue
}

// Round ticks for players
func (w *World) HandleMobRoundTicks() util.MessageQueue {

	messageQueue := util.NewMessageQueue(0, 0)

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
				if response, err := scripting.TryBuffScriptEvent(`onTrigger`, 0, mobInstanceId, buff.BuffId, w); err == nil {
					messageQueue.AbsorbMessages(response)
				}
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
						w.QueueCommand(0, mob.InstanceId, cmd)
					}
				}
			}
		}

		// Recalculate all stats at the end of the round tick
		mob.Character.Validate()

		if mob.Character.Health <= 0 {
			// Mob died
			w.QueueCommand(0, mob.InstanceId, `suicide`)
		}

	}

	return messageQueue
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

func (w *World) PruneBuffs() util.MessageQueue {

	messageQueue := util.NewMessageQueue(0, 0)

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
						if response, err := scripting.TryBuffScriptEvent(`onEnd`, uId, 0, buffInfo.BuffId, w); err == nil {
							messageQueue.AbsorbMessages(response)
						}
						if buffInfo.BuffId == 0 {
							logOff = true
						}
					}

					user.Character.Validate()

					if logOff {
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
				if response, err := scripting.TryBuffScriptEvent(`onEnd`, 0, mobInstanceId, buffInfo.BuffId, w); err == nil {
					messageQueue.AbsorbMessages(response)
				}
			}

			mob.Character.Validate()
		}

	}

	return messageQueue
}

func (w *World) HandleRespawns() (messageQueue util.MessageQueue) {

	messageQueue = util.NewMessageQueue(0, 0)

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
								messageQueue.SendRoomMessage(room.RoomId, spawnInfo.Message, true)
							}
						}
					}
				}

			}
		}
	}

	return messageQueue
}

// WHere combat happens
func (w *World) HandlePlayerCombat() (messageQueue util.MessageQueue, affectedPlayerIds []int, affectedMobInstanceIds []int) {

	tStart := time.Now()

	c := configs.GetConfig()

	messageQueue = util.NewMessageQueue(0, 0)

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

		if user.Character.Aggro.Type == characters.Flee {

			// Revert to Default combat regardless of outcome
			user.Character.SetAggro(user.Character.Aggro.UserId, user.Character.Aggro.MobInstanceId, characters.DefaultAttack)

			// The test to flee is performed against every mob attacking the player.
			uRoom := rooms.LoadRoom(roomId)
			if uRoom == nil {
				continue
			}

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
						break
					}
				}
			}

			if blockedByMob != `` {
				messageQueue.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="red-bold"><ansi fg="mobname">%s</ansi> blocks you from fleeing!</ansi>`, blockedByMob), true)
				messageQueue.SendRoomMessage(roomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> is blocked from fleeing by <ansi fg="mobname">%s</ansi>!`, user.Character.Name, blockedByMob), true, userId)
				continue
			}

			if blockedByPlayer != `` {
				messageQueue.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="red-bold"><ansi fg="username">%s</ansi> blocks you from fleeing!</ansi>`, blockedByPlayer), true)
				messageQueue.SendRoomMessage(roomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> is blocked from fleeing by <ansi fg="username">%s</ansi>!`, user.Character.Name, blockedByPlayer), true, userId)
				continue
			}

			// Success!
			exitName, exitRoomId := uRoom.GetRandomExit()

			if exitRoomId == 0 {
				messageQueue.SendUserMessage(userId, `You can't find an exit!`, true)
				continue
			}

			messageQueue.SendUserMessage(userId, fmt.Sprintf(`You flee to the <ansi fg="exit">%s</ansi> exit!`, exitName), true)
			messageQueue.SendRoomMessage(roomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> flees to the <ansi fg="exit">%s</ansi> exit!`, user.Character.Name, exitName), true, userId)

			rooms.MoveToRoom(userId, exitRoomId)

			for _, instId := range uRoom.GetMobs(rooms.FindCharmed) {
				if mob := mobs.GetInstance(instId); mob != nil {
					// Charmed mobs assist
					if mob.Character.IsCharmed(userId) {
						w.QueueCommand(0, instId, exitName)
					}
				}
			}

			continue
		}

		if user.Character.Aggro != nil && user.Character.Aggro.Type == characters.SpellCast {

			if user.Character.Aggro.RoundsWaiting > 0 {
				user.Character.Aggro.RoundsWaiting--

				if res, err := scripting.TrySpellScriptEvent(`onWait`, user.UserId, 0, user.Character.Aggro.SpellInfo, w); err == nil {
					messageQueue.AbsorbMessages(res)
				}

				continue
			}

			roll := util.RollDice(1, 100)
			successChance := user.Character.GetBaseCastSuccessChance(user.Character.Aggro.SpellInfo.SpellId)
			if roll >= successChance {

				// fail
				messageQueue.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="magenta">***</ansi> Your spell fizzles! <ansi fg="magenta">***</ansi> (Rolled %d on %d%% chance of success)`, roll, successChance), true)
				messageQueue.SendRoomMessage(roomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to cast a spell but it <ansi fg="magenta">fizzles</ansi>!`, user.Character.Name), true, userId)
				user.Character.Aggro = nil

				continue

			}

			allowRetaliation := true
			if res, err := scripting.TrySpellScriptEvent(`onMagic`, user.UserId, 0, user.Character.Aggro.SpellInfo, w); err == nil {
				messageQueue.AbsorbMessages(res)

				if res.Handled {
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
									w.QueueCommand(0, defMob.InstanceId, fmt.Sprintf("attack @%d", user.UserId)) // @ means player
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

			targetFound := true
			if defUser == nil {
				targetFound = false
			} else if defUser.Character.RoomId != user.Character.RoomId {

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
					if _, exitRoomId := uRoom.FindExitByName(user.Character.Aggro.ExitName); exitRoomId != defUser.Character.RoomId {
						targetFound = false
					}

				}

			}

			if !targetFound {
				messageQueue.SendUserMessage(user.UserId, "Your target can't be found.", true)
				user.Character.Aggro = nil
				continue
			}

			defUser.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

			if defUser.Character.Health < 1 {
				messageQueue.SendUserMessage(user.UserId, "Your rage subsides.", true)
				user.Character.Aggro = nil
				continue
			}

			if user.Character.Aggro.RoundsWaiting > 0 {
				slog.Info(`RoundsWaiting`, `User`, user.Character.Name, `Rounds`, user.Character.Aggro.RoundsWaiting)

				user.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, user.Character, defUser.Character, combat.User, combat.User)

				messageQueue.SendUserMessages(user.UserId, roundResult.MessagesToSource, true)
				messageQueue.SendUserMessages(defUser.UserId, roundResult.MessagesToTarget, true)

				if len(roundResult.MessagesToSourceRoom) > 0 {
					messageQueue.SendRoomMessages(user.Character.RoomId, roundResult.MessagesToSourceRoom, true, user.UserId, defUser.UserId)
				}

				if len(roundResult.MessagesToTargetRoom) > 0 {
					messageQueue.SendRoomMessages(defUser.Character.RoomId, roundResult.MessagesToTargetRoom, true, user.UserId, defUser.UserId)
				}

				continue
			}

			// Can't see them, can't fight them.
			if defUser.Character.HasBuffFlag(buffs.Hidden) {
				messageQueue.SendUserMessage(user.UserId, "You can't seem to find your target.", true)
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

						w.QueueCommand(0, instanceId, fmt.Sprintf("attack @%d", user.UserId)) // # denotes a specific user id
					}
				}
			}

			for _, buffId := range roundResult.BuffSource {
				w.QueueBuff(user.UserId, 0, buffId)
			}

			for _, buffId := range roundResult.BuffTarget {
				w.QueueBuff(defUser.UserId, 0, buffId)
			}

			messageQueue.SendUserMessages(user.UserId, roundResult.MessagesToSource, true)
			messageQueue.SendUserMessages(defUser.UserId, roundResult.MessagesToTarget, true)

			if len(roundResult.MessagesToSourceRoom) > 0 {
				messageQueue.SendRoomMessages(user.Character.RoomId, roundResult.MessagesToSourceRoom, true, user.UserId, defUser.UserId)
			}

			if len(roundResult.MessagesToTargetRoom) > 0 {
				messageQueue.SendRoomMessages(defUser.Character.RoomId, roundResult.MessagesToTargetRoom, true, user.UserId, defUser.UserId)
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

						messageQueue.SendUserMessage(defUser.UserId, `<ansi fg="202">***</ansi>`, true)
						messageQueue.SendUserMessage(defUser.UserId, fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> Your <ansi fg="item">%s</ansi> breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple()), true)
						messageQueue.SendUserMessage(defUser.UserId, `<ansi fg="202">***</ansi>`, true)

						messageQueue.SendRoomMessage(defUser.Character.RoomId, fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="username">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple(), defUser.Character.Name), true, defUser.UserId)

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
				messageQueue.SendUserMessage(user.UserId, "Your target can't be found.", true)
				user.Character.Aggro = nil
				continue
			}

			defMob.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

			if defMob.Character.Health < 1 {
				messageQueue.SendUserMessage(user.UserId, "Your rage subsides.", true)
				user.Character.Aggro = nil
				continue
			}

			if user.Character.Aggro.RoundsWaiting > 0 {
				slog.Info(`RoundsWaiting`, `User`, user.Character.Name, `Rounds`, user.Character.Aggro.RoundsWaiting)

				user.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, user.Character, &defMob.Character, combat.User, combat.Mob)

				messageQueue.SendUserMessages(user.UserId, roundResult.MessagesToSource, true)

				if len(roundResult.MessagesToSourceRoom) > 0 {
					messageQueue.SendRoomMessages(user.Character.RoomId, roundResult.MessagesToSourceRoom, true, user.UserId)
				}

				if len(roundResult.MessagesToTargetRoom) > 0 {
					messageQueue.SendRoomMessages(defMob.Character.RoomId, roundResult.MessagesToTargetRoom, true, user.UserId)
				}

				continue
			}

			// Can't see them, can't fight them.
			if defMob.Character.HasBuffFlag(buffs.Hidden) {
				messageQueue.SendUserMessage(user.UserId, "You can't seem to find your target.", true)
				continue
			}

			affectedPlayerIds = append(affectedPlayerIds, user.Character.Aggro.UserId)

			var roundResult combat.AttackResult

			roundResult = combat.AttackPlayerVsMob(user, defMob)

			for _, buffId := range roundResult.BuffSource {
				w.QueueBuff(user.UserId, 0, buffId)
			}

			for _, buffId := range roundResult.BuffTarget {
				w.QueueBuff(0, defMob.InstanceId, buffId)
			}

			messageQueue.SendUserMessages(user.UserId, roundResult.MessagesToSource, true)
			// messageQueue.SendUserMessages(defMob.InstanceId, roundResult.MessagesToTarget, true) // mobs don't get messages

			if len(roundResult.MessagesToSourceRoom) > 0 {
				messageQueue.SendRoomMessages(user.Character.RoomId, roundResult.MessagesToSourceRoom, true, user.UserId)
			}

			if len(roundResult.MessagesToTargetRoom) > 0 {
				messageQueue.SendRoomMessages(defMob.Character.RoomId, roundResult.MessagesToTargetRoom, true, user.UserId)
			}

			// Handle any scripted behavior now.
			if roundResult.Hit {
				if res, err := scripting.TryMobScriptEvent(`onHurt`, defMob.InstanceId, user.UserId, `user`, map[string]any{`damage`: roundResult.DamageToTarget, `crit`: roundResult.Crit}, w); err == nil {
					messageQueue.AbsorbMessages(res)
				}
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
								w.QueueCommand(0, defMob.InstanceId, fmt.Sprintf(`go %s`, exitName))

								if actionStr := defMob.GetAngryCommand(); actionStr != `` {
									w.QueueCommand(0, defMob.InstanceId, actionStr)
								}
								break
							}
						}
					}
				}
				w.QueueCommand(0, defMob.InstanceId, fmt.Sprintf("attack @%d", user.UserId)) // @ means player
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

	return messageQueue, affectedPlayerIds, affectedMobInstanceIds
}

// Mob combat operations may happen when players are not present.
func (w *World) HandleMobCombat() (messageQueue util.MessageQueue, affectedPlayerIds []int, affectedMobInstanceIds []int) {

	c := configs.GetConfig()

	tStart := time.Now()

	messageQueue = util.NewMessageQueue(0, 0)

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

		// Disable any buffs that are cancelled by combat
		mob.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

		if mob.Character.Aggro != nil && mob.Character.Aggro.Type == characters.SpellCast {

			if mob.Character.Aggro.RoundsWaiting > 0 {
				mob.Character.Aggro.RoundsWaiting--

				if res, err := scripting.TrySpellScriptEvent(`onWait`, 0, mob.InstanceId, mob.Character.Aggro.SpellInfo, w); err == nil {
					messageQueue.AbsorbMessages(res)
				}

				continue
			}

			successChance := mob.Character.GetBaseCastSuccessChance(mob.Character.Aggro.SpellInfo.SpellId)
			if util.RollDice(1, 100) >= successChance {

				// fail
				messageQueue.SendRoomMessage(mob.Character.RoomId, fmt.Sprintf(`<ansi fg="mobnamme">%s</ansi> tries to cast a spell but it <ansi fg="magenta">fizzles</ansi>!`, mob.Character.Name), true)
				mob.Character.Aggro = nil

				continue

			}

			allowRetaliation := true
			if res, err := scripting.TrySpellScriptEvent(`onMagic`, 0, mob.InstanceId, mob.Character.Aggro.SpellInfo, w); err == nil {
				messageQueue.AbsorbMessages(res)

				if res.Handled {
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
									w.QueueCommand(0, defMob.InstanceId, fmt.Sprintf("attack #%d", mob.InstanceId)) // # means mob
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
						w.QueueCommand(0, mob.InstanceId, `say I have a CombatAction that is too long. Please notify an admin.`)
					} else {
						for turnDelay, action := range strings.Split(combatAction, `;`) {
							w.QueueCommand(0, mob.InstanceId, action, turnDelay)
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
						w.QueueCommand(0, mob.InstanceId, fmt.Sprintf("equip %s", possibleWeapons[util.Rand(len(possibleWeapons))]))
					}

				}
			}

			if mob.Character.Aggro.RoundsWaiting > 0 {
				slog.Info(`RoundsWaiting`, `User`, mob.Character.Name, `Rounds`, mob.Character.Aggro.RoundsWaiting)

				mob.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, &mob.Character, defUser.Character, combat.Mob, combat.User)

				messageQueue.SendUserMessages(defUser.UserId, roundResult.MessagesToTarget, true)

				if len(roundResult.MessagesToSourceRoom) > 0 {
					messageQueue.SendRoomMessages(mob.Character.RoomId, roundResult.MessagesToSourceRoom, true, defUser.UserId)
				}

				if len(roundResult.MessagesToTargetRoom) > 0 {
					messageQueue.SendRoomMessages(defUser.Character.RoomId, roundResult.MessagesToTargetRoom, true, defUser.UserId)
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
						w.QueueCommand(0, instanceId, fmt.Sprintf("attack #%d", mob.InstanceId)) // # denotes a specific mob id
					}
				}
			}

			for _, buffId := range roundResult.BuffSource {
				w.QueueBuff(0, mob.InstanceId, buffId)
			}

			for _, buffId := range roundResult.BuffTarget {
				w.QueueBuff(defUser.UserId, 0, buffId)
			}

			messageQueue.SendUserMessages(defUser.UserId, roundResult.MessagesToTarget, true)

			if len(roundResult.MessagesToSourceRoom) > 0 {
				messageQueue.SendRoomMessages(mob.Character.RoomId, roundResult.MessagesToSourceRoom, true, defUser.UserId)
			}

			if len(roundResult.MessagesToTargetRoom) > 0 {
				messageQueue.SendRoomMessages(defUser.Character.RoomId, roundResult.MessagesToTargetRoom, true, defUser.UserId)
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

						messageQueue.SendUserMessage(defUser.UserId, `<ansi fg="202">***</ansi>`, true)
						messageQueue.SendUserMessage(defUser.UserId, fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> Your <ansi fg="item">%s</ansi> breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple()), true)
						messageQueue.SendUserMessage(defUser.UserId, `<ansi fg="202">***</ansi>`, true)

						messageQueue.SendRoomMessage(defUser.Character.RoomId, fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="username">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple(), defUser.Character.Name), true, defUser.UserId)

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

			defMob.Character.CancelBuffsWithFlag(buffs.CancelIfCombat)

			if defMob.Character.Health < 1 {
				mob.Character.Aggro = nil
				continue
			}

			if mob.Character.Aggro.RoundsWaiting > 0 {
				slog.Info(`RoundsWaiting`, `User`, mob.Character.Name, `Rounds`, mob.Character.Aggro.RoundsWaiting)

				mob.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, &mob.Character, &defMob.Character, combat.Mob, combat.Mob)

				if len(roundResult.MessagesToSourceRoom) > 0 {
					messageQueue.SendRoomMessages(mob.Character.RoomId, roundResult.MessagesToSourceRoom, true)
				}

				if len(roundResult.MessagesToTargetRoom) > 0 {
					messageQueue.SendRoomMessages(defMob.Character.RoomId, roundResult.MessagesToTargetRoom, true)
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
				w.QueueBuff(0, mob.InstanceId, buffId)
			}

			for _, buffId := range roundResult.BuffTarget {
				w.QueueBuff(0, defMob.InstanceId, buffId)
			}

			if len(roundResult.MessagesToSourceRoom) > 0 {
				messageQueue.SendRoomMessages(mob.Character.RoomId, roundResult.MessagesToSourceRoom, true)
			}

			if len(roundResult.MessagesToTargetRoom) > 0 {
				messageQueue.SendRoomMessages(defMob.Character.RoomId, roundResult.MessagesToTargetRoom, true)
			}

			// Handle any scripted behavior now.
			if roundResult.Hit {
				if res, err := scripting.TryMobScriptEvent(`onHurt`, defMob.InstanceId, mob.InstanceId, `mob`, map[string]any{`damage`: roundResult.DamageToTarget, `crit`: roundResult.Crit}, w); err == nil {
					messageQueue.AbsorbMessages(res)
				}
			}

			// Mobs get aggro when attacked
			if defMob.Character.Aggro == nil {
				defMob.PreventIdle = true
				defMob.Character.Aggro = &characters.Aggro{
					Type: characters.DefaultAttack,
				}
				w.QueueCommand(0, defMob.InstanceId, fmt.Sprintf("attack #%d", mob.InstanceId)) // # means mob
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

						if room := rooms.LoadRoom(roomId); room != nil {

							messageQueue.SendRoomMessage(roomId, fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="mobname">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, defMob.Character.Equipment.Offhand.NameSimple(), defMob.Character.Name), true)

							defMob.Character.RemoveFromBody(defMob.Character.Equipment.Offhand)
							itm := items.New(20) // Broken item
							if !defMob.Character.StoreItem(itm) {
								room.AddItem(itm, false)
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

	return messageQueue, affectedPlayerIds, affectedMobInstanceIds
}

func (w *World) HandleAffected(affectedPlayerIds []int, affectedMobInstanceIds []int) (messageQueue util.MessageQueue) {
	messageQueue = util.NewMessageQueue(0, 0)

	playersHandled := map[int]struct{}{}
	for _, userId := range affectedPlayerIds {
		if _, ok := playersHandled[userId]; ok {
			continue
		}
		playersHandled[userId] = struct{}{}

		if user := users.GetByUserId(userId); user != nil {

			if user.Character.Health <= -10 {

				w.QueueCommand(userId, 0, "suicide") // suicide drops all money/items and transports to land of the dead.

			} else if user.Character.Health < 1 {

				messageQueue.SendUserMessage(userId,
					`<ansi fg="red">you drop to the ground!</ansi>`,
					true)

				messageQueue.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> <ansi fg="red">drops to the ground!</ansi>`, user.Character.Name),
					true,
					user.UserId)

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
				w.QueueCommand(0, mobId, "suicide") // suicide drops all money/items and transports to land of the dead.
			}
		}

	}

	return messageQueue
}

// Idle Mobs
func (w *World) HandleIdleMobs() util.MessageQueue {

	// c := configs.GetConfig()

	maxBoredom := uint8(configs.GetConfig().MaxMobBoredom)

	messageQueue := util.NewMessageQueue(0, 0)

	// Handle idle mob behavior
	tStart := time.Now()
	for _, mobId := range mobs.GetAllMobInstanceIds() {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		if mob.BoredomCounter >= maxBoredom {
			if mob.Despawns() {
				w.QueueCommand(0, mob.InstanceId, `despawn`+fmt.Sprintf(` depression %d/%d`, mob.BoredomCounter, maxBoredom))
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
					w.QueueCommand(0, mob.InstanceId, `emote mumbles about losing their quarry.`)
					mob.Character.Aggro = nil
				}
			}
			continue
		}

		// If they have idle commands, maybe do one of them?
		result, _ := scripting.TryMobScriptEvent("onIdle", mob.InstanceId, 0, ``, nil, w)
		messageQueue.AbsorbMessages(result)

		if !result.Handled {
			if !mob.Character.IsCharmed() { // Won't do this stuff if befriended

				if mob.MaxWander > -1 && len(mob.RoomStack) > mob.MaxWander {
					mob.GoingHome = true
				}
				if mob.GoingHome {
					w.QueueCommand(0, mob.InstanceId, `go home`)
					continue
				}

			}
		}

		//
		// Look for trouble
		//
		if !result.Handled {
			if mob.Character.IsCharmed() {
				// Only some mobs can apply first aid
				if mob.Character.KnowsFirstAid() {
					w.QueueCommand(0, mob.InstanceId, `lookforaid`)
				}
			} else {
				w.QueueCommand(0, mob.InstanceId, `lookfortrouble`)
			}
		}

	}

	util.TrackTime(`HandleIdleMobs()`, time.Since(tStart).Seconds())

	return messageQueue
}

// Healing
func (w *World) HandleAutoHealing(roundNumber uint64) util.MessageQueue {

	messageQueue := util.NewMessageQueue(0, 0)

	// Every 3 rounds.
	if roundNumber%3 != 0 {
		return messageQueue
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
				if user.Character.Health > -10 {
					user.Character.Health--
					messageQueue.SendUserMessage(userId, `<ansi fg="red">you are bleeding out!</ansi>`, true)
					messageQueue.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> is <ansi fg="red">bleeding out</ansi>! Somebody needs to provide aid!`, user.Character.Name), true, userId)
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

		w.connectionPool.SendTo([]byte(templates.AnsiParse(user.GetCommandPrompt(true))), user.ConnectionId())
	}

	return messageQueue
}

// Special shadow realm stuff
func (w *World) HandleShadowRealm(roundNumber uint64) util.MessageQueue {

	messageQueue := util.NewMessageQueue(0, 0)

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
				Title:   "shimmering portal",
				UserId:  0,
				Expires: time.Now().Add(time.Second * 10),
			}
			// Spawn a portal in the room that leads to the portal location
			deadRoom.AddTemporaryExit("shimmering portal", tmpExit)
			messageQueue.SendRoomMessage(75, `<ansi fg="magenta-bold">A shimmering portal appears in the room.</ansi>`, true)
		}
	}

	return messageQueue
}

// Handle dropped players
func (w *World) HandleDroppedPlayers(droppedPlayers []int) util.MessageQueue {

	messageQueue := util.NewMessageQueue(0, 0)

	if len(droppedPlayers) == 0 {
		return messageQueue
	}

	for _, userId := range droppedPlayers {
		if user := users.GetByUserId(userId); user != nil {

			messageQueue.SendUserMessage(userId,
				`<ansi fg="red">you drop to the ground!</ansi>`,
				true)
			messageQueue.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> <ansi fg="red">drops to the ground!</ansi>`, user.Character.Name),
				true,
				user.UserId)
		}
	}

	return messageQueue
}

// Levelups
func (w *World) CheckForLevelUps() util.MessageQueue {

	messageQueue := util.NewMessageQueue(0, 0)

	onlineIds := users.GetOnlineUserIds()
	for _, userId := range onlineIds {
		user := users.GetByUserId(userId)

		for {

			if newLevel, statsDelta := user.Character.LevelUp(); newLevel {

				levelUpData := map[string]interface{}{
					"level":          user.Character.Level,
					"statsDelta":     statsDelta,
					"trainingPoints": 1,
					"statPoints":     1,
				}
				levelUpStr, _ := templates.Process("character/levelup", levelUpData)

				messageQueue.SendUserMessage(user.UserId, levelUpStr, true)

				w.Broadcast(
					templates.AnsiParse(fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> <ansi fg="yellow">has leveled up to level %d!</ansi> <ansi fg="magenta-bold">***</ansi>%s`, user.Character.Name, user.Character.Level, term.CRLFStr)),
				)

				go users.SaveUser(*user)

				continue
			}

			break

		}

	}

	return messageQueue
}

// Checks for current auction and handles updates/communication
func (w *World) ProcessAuction(tNow time.Time) {

	a := auctions.GetCurrentAuction()
	if a == nil {
		return
	}

	auctionOnConnectionIds := []uint64{}
	for _, uid := range users.GetOnlineUserIds() {
		if u := users.GetByUserId(uid); u != nil {
			auctionOn := u.GetConfigOption(`auction`)
			if auctionOn == nil || auctionOn.(bool) {
				auctionOnConnectionIds = append(auctionOnConnectionIds, u.ConnectionId())
			}
		}
	}

	if len(auctionOnConnectionIds) == 0 {
		return
	}

	c := configs.GetConfig()

	if a.IsEnded() {

		auctions.EndAuction()

		a.LastUpdate = tNow
		auctionTxt, _ := templates.Process("auctions/auction-end", a)
		w.GetConnectionPool().SendTo([]byte(auctionTxt), auctionOnConnectionIds...)

		// Give the item to the winner and let them know
		if a.HighestBidUserId > 0 {

			if user := users.GetByUserId(a.HighestBidUserId); user != nil {
				if user.Character.StoreItem(a.ItemData) {
					msg := templates.AnsiParse(fmt.Sprintf(`<ansi fg="yellow">You have won the auction for the <ansi fg="item">%s</ansi>! It has been added to your backpack.</ansi>%s`, a.ItemData.DisplayName(), term.CRLFStr))
					w.GetConnectionPool().SendTo([]byte(msg), user.ConnectionId())
				}
			}

		} else {
			if user := users.GetByUserId(a.SellerUserId); user != nil {
				if user.Character.StoreItem(a.ItemData) {
					msg := templates.AnsiParse(fmt.Sprintf(`<ansi fg="yellow">The auction for the <ansi fg="item">%s</ansi> has ended without a winner. It has been returned to you.</ansi>%s`, a.ItemData.DisplayName(), term.CRLFStr))
					w.GetConnectionPool().SendTo([]byte(msg), user.ConnectionId())
				}
			}
		}

	} else if a.LastUpdate.IsZero() {

		a.LastUpdate = tNow
		auctionTxt, _ := templates.Process("auctions/auction-start", a)
		w.GetConnectionPool().SendTo([]byte(auctionTxt), auctionOnConnectionIds...)

	} else if time.Since(a.LastUpdate) > time.Second*time.Duration(c.AuctionUpdateSeconds) {

		a.LastUpdate = tNow
		auctionTxt, _ := templates.Process("auctions/auction-update", a)
		w.GetConnectionPool().SendTo([]byte(auctionTxt), auctionOnConnectionIds...)

	}

}
