package rooms

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/exit"
	"github.com/volte6/gomud/internal/fileloader"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"

	"log/slog"

	"gopkg.in/yaml.v2"
)

var (
	roomManager = &RoomManager{
		rooms:                make(map[int]*Room),
		zones:                make(map[string]ZoneInfo),
		roomsWithUsers:       make(map[int]int),
		roomsWithMobs:        make(map[int]int),
		roomDescriptionCache: make(map[string]string),
		roomIdToFileCache:    make(map[int]string),
	}
)

type RoomManager struct {
	rooms                map[int]*Room
	zones                map[string]ZoneInfo // a map of zone name to room id
	roomsWithUsers       map[int]int         // key is roomId to # players
	roomsWithMobs        map[int]int         // key is roomId to # mobs
	topRoomItems         []int               // list of the top room items
	roomDescriptionCache map[string]string   // key is a hash, value is the description
	roomIdToFileCache    map[int]string      // key is room id, value is the file path
}

const (
	GoblinZone       = `Endless Trashheap`
	GoblinRoom       = 139
	StartRoomIdAlias = 0
)

type ZoneInfo struct {
	RootRoomId      int
	DefaultBiome    string // city, swamp etc. see biomes.go
	HasZoneMutators bool   // does it have any zone mutators assigned?
	RoomIds         map[int]struct{}
}

func GetNextRoomId() int {
	return int(configs.GetConfig().NextRoomId)
}

func SetNextRoomId(nextRoomId int) {
	configs.SetVal(`nextroomid`, strconv.Itoa(nextRoomId), true)
}

func GetAllRoomIds() []int {

	var roomIds []int = make([]int, len(roomManager.roomIdToFileCache))
	i := 0
	for roomId, _ := range roomManager.roomIdToFileCache {
		roomIds[i] = roomId
		i++
	}

	return roomIds
}

func GetZonesWithMutators() ([]string, []int) {

	zNames := []string{}
	rootRoomIds := []int{}

	for zName, zInfo := range roomManager.zones {
		if zInfo.HasZoneMutators {
			zNames = append(zNames, zName)
			rootRoomIds = append(rootRoomIds, zInfo.RootRoomId)
		}
	}
	return zNames, rootRoomIds
}

func RoomMaintenance() bool {
	start := time.Now()
	defer func() {
		util.TrackTime(`RoomMaintenance()`, time.Since(start).Seconds())
	}()

	roundCount := util.GetRoundCount()
	// Get the current round count
	unloadRoundThreshold := roundCount - roomUnloadTimeoutRounds
	unloadRooms := make([]*Room, 0)

	roomsUpdated := false
	for _, room := range roomManager.rooms {

		for _, fx := range room.GetEffects() {

			if fx.Type == Wildfire {

				if fx.Expired() { // Wildfire spreads on expiration

					room.SendText(`The ` + colorpatterns.ApplyColorPattern(`burning`, `flame`) + ` finally subsides.`)

					for _, exitInfo := range room.Exits {

						if util.RollDice(1, 3) == 1 { // 33% chance of not spreading to this exit
							continue
						}

						events.Requeue(events.RoomAction{
							RoomId: exitInfo.RoomId,
							Action: string(Wildfire),
						})
					}

					for _, exitInfo := range room.ExitsTemp {

						if util.RollDice(1, 3) == 1 { // 33% chance of not spreading to this exit
							continue
						}

						events.Requeue(events.RoomAction{
							RoomId: exitInfo.RoomId,
							Action: string(Wildfire),
						})
					}

				} else {

					for _, uid := range room.GetPlayers() {
						if user := users.GetByUserId(uid); user != nil {
							if user.Character.HasBuff(22) { // burning
								continue
							}
							user.AddBuff(22)
						}
					}

					for _, miid := range room.GetMobs() {
						if mob := mobs.GetInstance(miid); mob != nil {
							if mob.Character.HasBuff(22) { // burning
								continue
							}
							mob.AddBuff(22)
						}
					}

					for _, itm := range room.GetAllFloorItems(false) {
						room.SendText(`The ` + itm.DisplayName() + ` that was laying on the ground is destroyed by ` + colorpatterns.ApplyColorPattern(`flames`, `flame`) + `.`)
						room.RemoveItem(itm, false)
					}

				}
			}

		}

		if visitorCt := room.PruneVisitors(); visitorCt > 0 {
			roomsUpdated = true
		}

		// Notify that room that something happened to the sign?
		if prunedSigns := room.PruneSigns(); len(prunedSigns) > 0 {
			roomsUpdated = true

			if roomPlayers := room.GetPlayers(); len(roomPlayers) > 0 {
				for _, userId := range roomPlayers {
					for _, sign := range prunedSigns {
						if sign.VisibleUserId == 0 {
							if u := users.GetByUserId(userId); u != nil {
								u.SendText("A sign crumbles to dust.\n")
							}
						} else if sign.VisibleUserId == userId {
							if u := users.GetByUserId(userId); u != nil {
								u.SendText("The rune you had enscribed here has faded away.\n")
							}
						}
					}
				}
			}
		}

		// Notify the room that the temp exits disappeared?
		if prunedExits := room.PruneTemporaryExits(); len(prunedExits) > 0 {
			roomsUpdated = true

			if roomPlayers := room.GetPlayers(); len(roomPlayers) > 0 {
				for _, exit := range prunedExits {
					for _, userId := range roomPlayers {
						if u := users.GetByUserId(userId); u != nil {
							u.SendText(fmt.Sprintf("The %s vanishes.\n", exit.Title))
						}
					}
				}
			}
		}

		// If a room is burning, don't clean it up
		if room.IsBurning() {
			continue
		}

		// Consider unloading rooms from memory?
		if roundCount%roomUnloadTimeoutRounds == 0 {
			if room.lastVisited < unloadRoundThreshold {
				unloadRooms = append(unloadRooms, room)
			}
		}

	}

	if len(unloadRooms) > 0 {
		for _, room := range unloadRooms {
			removeRoomFromMemory(room)
		}
		roomsUpdated = true
	}

	return roomsUpdated
}

func GetAllZoneNames() []string {

	var zoneNames []string = make([]string, len(roomManager.zones))
	i := 0
	for zoneName, _ := range roomManager.zones {
		zoneNames[i] = zoneName
		i++
	}

	return zoneNames
}

func MoveToRoom(userId int, toRoomId int, isSpawn ...bool) error {

	user := users.GetByUserId(userId)

	currentRoom := LoadRoom(user.Character.RoomId)
	if currentRoom == nil {
		return fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	cfg := configs.GetConfig()

	if toRoomId == StartRoomIdAlias {

		// If "StartRoom" is set for MiscData on the char, use that.
		if charStartRoomId := user.Character.GetMiscData(`StartRoom`); charStartRoomId != nil {
			if rId, ok := charStartRoomId.(int); ok {
				toRoomId = rId
			}
		}

		// If still StartRoomIdAlias, use config value
		if toRoomId == StartRoomIdAlias && cfg.StartRoom != 0 {
			toRoomId = int(cfg.StartRoom)
		}

		// If toRomoId is zero after all this, default to 1
		if toRoomId == 0 {
			toRoomId = 1
		}
	}

	newRoom := LoadRoom(toRoomId)
	if newRoom == nil {
		return fmt.Errorf(`room %d not found`, toRoomId)
	}

	// r.prepare locks, so do it before the upcoming lock
	if len(newRoom.players) == 0 {
		newRoom.Prepare(true)
	}

	currentRoom.MarkVisited(userId, VisitorUser, 1)

	if len, _ := currentRoom.RemovePlayer(userId); len < 1 {
		delete(roomManager.roomsWithUsers, currentRoom.RoomId)
	}

	newRoom.MarkVisited(userId, VisitorUser)

	//
	// Apply any mutators from the zone or room
	// This will only add mutators that the player
	// doesn't already have.
	//
	for mut := range newRoom.ActiveMutators {
		spec := mut.GetSpec()
		if len(spec.PlayerBuffIds) == 0 {
			continue
		}
		for _, buffId := range spec.PlayerBuffIds {
			if !user.Character.HasBuff(buffId) {
				user.AddBuff(buffId)
			}
		}
	}
	//
	// Done adding mutator buffs
	//

	playerCt := newRoom.addPlayer(userId)
	roomManager.roomsWithUsers[newRoom.RoomId] = playerCt

	formerRoomId := user.Character.RoomId
	user.Character.RoomId = newRoom.RoomId
	user.Character.Zone = newRoom.Zone
	user.Character.RememberRoom(newRoom.RoomId) // Mark this room as remembered.

	roundNow := util.GetRoundCount()

	if user.Character.Level < 5 && toRoomId > -1 {

		if formerRoomId > 0 && (formerRoomId < 900 || formerRoomId > 999) {

			var lastGuideRound uint64 = 0
			tmpLGR := user.GetTempData(`lastGuideRound`)
			if tmpLGRUint, ok := tmpLGR.(uint64); ok {
				lastGuideRound = tmpLGRUint
			}

			spawnGuide := false
			if (roundNow - lastGuideRound) > uint64(cfg.SecondsToRounds(300)) {
				spawnGuide = true
			}

			if spawnGuide {
				for _, miid := range user.Character.GetCharmIds() {
					if testMob := mobs.GetInstance(miid); testMob != nil && testMob.MobId == 38 {
						spawnGuide = false
						break
					}
				}
			}

			if spawnGuide {

				room := LoadRoom(toRoomId)
				guideMob := mobs.NewMobById(38, 1)
				guideMob.Character.Name = fmt.Sprintf(`%s's Guide`, user.Character.Name)
				room.AddMob(guideMob.InstanceId)
				guideMob.Character.Charm(userId, characters.CharmPermanent, characters.CharmExpiredDespawn)
				// Track it
				user.Character.TrackCharmed(guideMob.InstanceId, true)

				room.SendText(`<ansi fg="mobname">` + guideMob.Character.Name + `</ansi> appears in a shower of sparks!`)

				guideMob.Command(`sayto ` + user.ShorthandId() + ` I'll be here to help protect you while you learn the ropes.`)
				guideMob.Command(`sayto ` + user.ShorthandId() + ` I can create a portal to take us back to Town Square any time. Just <ansi fg="command">ask</ansi> me about it.`)

				user.SendText(`Your guide will try and stick around until you reach level 5.`)

				user.SetTempData(`lastGuideRound`, roundNow)

			}
		}
	}

	//
	// Send GMCP Updates
	//
	if connections.GetClientSettings(user.ConnectionId()).GmcpEnabled(`Room`) {

		newRoomPlayers := strings.Builder{}

		// Send to everyone in the new room that a player arrived
		for _, uid := range newRoom.GetPlayers() {

			if uid == user.UserId {
				continue
			}

			if u := users.GetByUserId(uid); u != nil {

				if newRoomPlayers.Len() > 0 {
					newRoomPlayers.WriteString(`, `)
				}

				newRoomPlayers.WriteString(`"` + u.Character.Name + `": ` + `"` + u.Character.Name + `"`)

				if connections.GetClientSettings(u.ConnectionId()).GmcpEnabled(`Room`) {

					bytesOut := []byte(fmt.Sprintf(`Room.AddPlayer {"name": "%s", "fullname": "%s"}`, user.Character.Name, user.Character.Name))
					connections.SendTo(
						term.GmcpPayload.BytesWithPayload(bytesOut),
						user.ConnectionId(),
					)
				}
			}
		}

		// Send to everyone in the old room that a player left
		for _, uid := range currentRoom.GetPlayers() {

			if uid == user.UserId {
				continue
			}

			if u := users.GetByUserId(uid); u != nil {
				if connections.GetClientSettings(u.ConnectionId()).GmcpEnabled(`Room`) {

					bytesOut := []byte(fmt.Sprintf(`Room.RemovePlayer "%s"`, user.Character.Name))
					connections.SendTo(
						term.GmcpPayload.BytesWithPayload(bytesOut),
						user.ConnectionId(),
					)
				}
			}
		}

		roomInfoStr := strings.Builder{}
		roomInfoStr.WriteString(`{ `)
		roomInfoStr.WriteString(`"num": ` + strconv.Itoa(newRoom.RoomId) + `, `)
		roomInfoStr.WriteString(`"name": "` + newRoom.Title + `", `)
		roomInfoStr.WriteString(`"area": "` + newRoom.Zone + `", `)
		roomInfoStr.WriteString(`"environment": "` + newRoom.GetBiome().Name() + `", `)

		// build exits
		roomInfoStr.WriteString(`"exits": {`)
		exitCt := 0
		for name, exitInfo := range newRoom.Exits {
			if exitInfo.Secret {
				continue
			}
			if exitCt > 0 {
				roomInfoStr.WriteString(`, `)
			}

			roomInfoStr.WriteString(`"` + name + `": ` + strconv.Itoa(exitInfo.RoomId))

			exitCt++
		}
		roomInfoStr.WriteString(`}, `)
		// End exits

		// build details
		roomInfoStr.WriteString(`"details": [`)

		detailCt := 0
		if len(newRoom.GetMobs(FindMerchant)) > 0 || len(newRoom.GetPlayers(FindMerchant)) > 0 {
			if detailCt > 0 {
				roomInfoStr.WriteString(`, `)
			}
			detailCt++
			roomInfoStr.WriteString(`"shop"`)
		}
		if len(newRoom.SkillTraining) > 0 {
			if detailCt > 0 {
				roomInfoStr.WriteString(`, `)
			}
			detailCt++
			roomInfoStr.WriteString(`"trainer"`)
		}
		if newRoom.IsBank {
			if detailCt > 0 {
				roomInfoStr.WriteString(`, `)
			}
			detailCt++
			roomInfoStr.WriteString(`"bank"`)
		}
		if newRoom.IsStorage {
			if detailCt > 0 {
				roomInfoStr.WriteString(`, `)
			}
			detailCt++
			roomInfoStr.WriteString(`"storage"`)
		}
		roomInfoStr.WriteString(`]`)
		// end details

		roomInfoStr.WriteString(` }`)
		// End room info

		// send big 'ol room info object
		connections.SendTo(
			term.GmcpPayload.BytesWithPayload([]byte("Room.Info "+roomInfoStr.String())),
			user.ConnectionId(),
		)

		// send player list for room
		connections.SendTo(
			term.GmcpPayload.BytesWithPayload([]byte("Room.Players {"+newRoomPlayers.String()+`}`)),
			user.ConnectionId(),
		)
	}

	return nil
}

// skipRecentlyVisited means ignore rooms with recent visitors
// minimumItemCt is the minimum items in the room to care about it
func GetRoomWithMostItems(skipRecentlyVisited bool, minimumItemCt int, minimumGoldCt int) (roomId int, itemCt int) {

	topItemRoomId, topItemCt := 0, 0
	topGoldRoomId, topGoldCt := 0, 0

	for cRoomId, cRoom := range roomManager.rooms {
		// Don't include goblin trash zone items
		if cRoom.Zone == GoblinZone {
			continue
		}

		iCt := len(cRoom.Items)

		if iCt < minimumItemCt && cRoom.Gold < minimumGoldCt {
			continue
		}

		if iCt > topItemCt {
			if skipRecentlyVisited && cRoom.HasRecentVisitors() {
				continue
			}
			topItemRoomId = cRoomId
			topItemCt = iCt
		}

		if cRoom.Gold > topGoldCt {
			if skipRecentlyVisited && cRoom.HasRecentVisitors() {
				continue
			}
			topGoldRoomId = cRoomId
			topGoldCt = cRoom.Gold
		}
	}

	if topItemRoomId == 0 && topGoldCt > 0 {
		return topGoldRoomId, topGoldCt
	}

	return topItemRoomId, topItemCt
}

func GetRoomsWithPlayers() []int {

	deleteKeys := []int{}
	roomsWithPlayers := []int{}

	for roomId, _ := range roomManager.roomsWithUsers {
		roomsWithPlayers = append(roomsWithPlayers, roomId)
	}

	for i := len(roomsWithPlayers) - 1; i >= 0; i-- {
		roomId := roomsWithPlayers[i]
		if r := LoadRoom(roomId); r != nil {
			if len(r.players) < 1 {
				roomsWithPlayers = append(roomsWithPlayers[:i], roomsWithPlayers[i+1:]...)
				deleteKeys = append(deleteKeys, roomId)
				continue
			}
		}
	}

	if len(deleteKeys) > 0 {

		for _, roomId := range deleteKeys {
			delete(roomManager.roomsWithUsers, roomId)
		}

	}

	return roomsWithPlayers
}

func GetRoomsWithMobs() []int {

	var roomsWithMobs []int = make([]int, len(roomManager.roomsWithMobs))
	i := 0
	for roomId, _ := range roomManager.roomsWithMobs {
		roomsWithMobs[i] = roomId
		i++
	}

	return roomsWithMobs
}

func SaveAllRooms() error {

	// Unhash the descriptions before saving
	for _, loadedRoom := range roomManager.rooms {

		if strings.HasPrefix(loadedRoom.Description, `h:`) {
			hash := strings.TrimPrefix(loadedRoom.Description, `h:`)
			if description, ok := roomManager.roomDescriptionCache[hash]; ok {
				loadedRoom.Description = description
			}
		}
	}

	start := time.Now()

	saveModes := []fileloader.SaveOption{}

	if configs.GetConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	saveCt, err := fileloader.SaveAllFlatFiles[int, *Room](roomDataFilesPath, roomManager.rooms, saveModes...)

	slog.Info("SaveAllRooms()", "savedCount", saveCt, "expectedCt", len(roomManager.rooms), "Time Taken", time.Since(start))

	return err
}

// Goes through all of the rooms and caches key information
func loadAllRoomZones() error {
	start := time.Now()

	nextRoomId := GetNextRoomId()
	defer func() {
		if nextRoomId != GetNextRoomId() {
			SetNextRoomId(nextRoomId)
		}
	}()

	loadedRooms, err := fileloader.LoadAllFlatFiles[int, *Room](roomDataFilesPath)
	if err != nil {
		return err
	}

	roomsWithoutEntrances := map[int]string{}

	for _, loadedRoom := range loadedRooms {

		// Room 75 is the death/shadow realm and gets a pass
		if loadedRoom.RoomId == 75 {
			continue
		}

		// If it has never been set, set it to the filepath
		if _, ok := roomsWithoutEntrances[loadedRoom.RoomId]; !ok {
			roomsWithoutEntrances[loadedRoom.RoomId] = loadedRoom.Filepath()
		}

		for _, exit := range loadedRoom.Exits {
			roomsWithoutEntrances[exit.RoomId] = ``
		}

	}

	for roomId, filePath := range roomsWithoutEntrances {

		if filePath == `` {
			delete(roomsWithoutEntrances, roomId)
			continue
		}

		slog.Warn("No Entrance", "roomId", roomId, "filePath", filePath)
	}

	for _, loadedRoom := range loadedRooms {
		// Keep track of the highest roomId

		if loadedRoom.RoomId >= nextRoomId {
			nextRoomId = loadedRoom.RoomId + 1
		}

		// Cache the file path for every roomId
		roomManager.roomIdToFileCache[loadedRoom.RoomId] = loadedRoom.Filepath()

		// Update the zone info cache
		if _, ok := roomManager.zones[loadedRoom.Zone]; !ok {
			roomManager.zones[loadedRoom.Zone] = ZoneInfo{
				RootRoomId: 0,
				RoomIds:    make(map[int]struct{}),
			}
		}

		// Update the zone info
		zoneInfo := roomManager.zones[loadedRoom.Zone]
		zoneInfo.RoomIds[loadedRoom.RoomId] = struct{}{}

		if loadedRoom.ZoneConfig.RoomId == loadedRoom.RoomId {
			zoneInfo.RootRoomId = loadedRoom.RoomId
			zoneInfo.DefaultBiome = loadedRoom.Biome

			if len(loadedRoom.ZoneConfig.Mutators) > 0 {
				zoneInfo.HasZoneMutators = true
			}
		}

		roomManager.zones[loadedRoom.Zone] = zoneInfo
	}

	slog.Info("rooms.loadAllRoomZones()", "loadedCount", len(loadedRooms), "Time Taken", time.Since(start))

	return nil
}

// Saves a room to disk and unloads it from memory
func removeRoomFromMemory(r *Room) {

	room, ok := roomManager.rooms[r.RoomId]

	if !ok {
		return
	}

	if len(room.players) > 0 {
		return
	}

	for _, mobInstanceId := range room.mobs {
		mobs.DestroyInstance(mobInstanceId)
	}

	for _, spawnDetails := range room.SpawnInfo {
		if spawnDetails.InstanceId > 0 {

			if m := mobs.GetInstance(spawnDetails.InstanceId); m != nil {
				if m.Character.RoomId == room.RoomId {
					mobs.DestroyInstance(spawnDetails.InstanceId)
				}
			}

		}
	}

	beforeCt := len(roomManager.rooms)

	SaveRoom(*room)
	delete(roomManager.rooms, r.RoomId)

	afterCt := len(roomManager.rooms)

	slog.Info("Removing from memory", "RoomId", r.RoomId, "Title", r.Title, "beforeCt", beforeCt, "afterCt", afterCt)
}

// Loads a room from disk and stores in memory
func addRoomToMemory(r *Room) {

	if _, ok := roomManager.rooms[r.RoomId]; ok {
		return
	}

	roomManager.rooms[r.RoomId] = r

	if _, ok := roomManager.roomIdToFileCache[r.RoomId]; !ok {
		roomManager.roomIdToFileCache[r.RoomId] = r.Filepath()
	}

	// Hash the descriptions and store centrally.
	// This saves a lot of memory because many descriptions are duplicates
	hash := util.Hash(r.Description)
	if _, ok := roomManager.roomDescriptionCache[hash]; !ok {
		roomManager.roomDescriptionCache[hash] = r.Description
	}
	r.Description = fmt.Sprintf(`h:%s`, hash)

	// Track whatever the last room id created is so we know what to number the next one.
	if r.RoomId >= GetNextRoomId() {
		SetNextRoomId(r.RoomId + 1)
	}

	if _, ok := roomManager.zones[r.Zone]; !ok {
		roomManager.zones[r.Zone] = ZoneInfo{
			RootRoomId: 0,
			RoomIds:    make(map[int]struct{}),
		}
	}

	// Populate the zone info
	zoneInfo := roomManager.zones[r.Zone]
	zoneInfo.RoomIds[r.RoomId] = struct{}{}

	if r.ZoneConfig.RoomId == r.RoomId {
		zoneInfo.RootRoomId = r.RoomId
	}

	roomManager.zones[r.Zone] = zoneInfo

}

func findRoomFile(roomId int) string {

	foundFilePath := ``
	searchFileName := filepath.FromSlash(fmt.Sprintf(`/%d.yaml`, roomId))

	walkPath := filepath.FromSlash(roomDataFilesPath)

	filepath.Walk(walkPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if strings.HasSuffix(path, searchFileName) {
			foundFilePath = path
			return errors.New(`found`)
		}

		return nil
	})

	return strings.TrimPrefix(foundFilePath, walkPath)
}

func loadRoomFromFile(roomFilePath string) (*Room, error) {

	roomFilePath = util.FilePath(roomFilePath)

	roomPtr, err := fileloader.LoadFlatFile[*Room](roomFilePath)
	if err != nil {
		slog.Error("loadRoomFromFile()", "error", err.Error())
		return roomPtr, err
	}

	// Automatically set the last visitor to now (reset the timer)
	roomPtr.lastVisited = util.GetRoundCount()

	addRoomToMemory(roomPtr)

	return roomPtr, err
}

func GetZoneRoot(zone string) (int, error) {

	if zoneInfo, ok := roomManager.zones[zone]; ok {
		return zoneInfo.RootRoomId, nil
	}

	return 0, fmt.Errorf("zone %s does not exist.", zone)
}

func GetZoneConfig(zone string) *ZoneConfig {

	zoneInfo, ok := roomManager.zones[zone]

	if ok {
		if r := LoadRoom(zoneInfo.RootRoomId); r != nil {
			return &r.ZoneConfig
		}
	}
	return nil
}

func IsRoomLoaded(roomId int) bool {

	_, ok := roomManager.rooms[roomId]
	return ok
}

// Load room grabs the room from memory and returns a pointer to it.
// If the room hasn't been loaded yet, it loads it into memory
func LoadRoom(roomId int) *Room {

	// Room 0 aliases to start room
	if roomId == StartRoomIdAlias {
		if roomId = int(configs.GetConfig().StartRoom); roomId == 0 {
			roomId = 1
		}
	}

	room, ok := roomManager.rooms[roomId]

	if ok {
		return room
	}

	filename := findRoomFile(roomId)
	retRoom, _ := loadRoomFromFile(util.FilePath(roomDataFilesPath, `/`, filename))

	return retRoom
}

func SaveRoom(r Room) error {

	if strings.HasPrefix(r.Description, `h:`) {
		hash := strings.TrimPrefix(r.Description, `h:`)
		if description, ok := roomManager.roomDescriptionCache[hash]; ok {
			r.Description = description
		}
	}

	data, err := yaml.Marshal(&r)
	if err != nil {
		return err
	}

	zone := ZoneToFolder(r.Zone)

	roomFilePath := util.FilePath(roomDataFilesPath, `/`, fmt.Sprintf("%s%d.yaml", zone, r.RoomId))

	if err = os.WriteFile(roomFilePath, data, 0777); err != nil {
		return err
	}

	slog.Info("Saved room", "room", r.RoomId)

	return nil
}

func ZoneStats(zone string) (rootRoomId int, totalRooms int, err error) {

	if zoneInfo, ok := roomManager.zones[zone]; ok {
		return zoneInfo.RootRoomId, len(zoneInfo.RoomIds), nil
	}

	return 0, 0, fmt.Errorf("zone %s does not exist.", zone)
}

func ZoneNameSanitize(zone string) string {
	if zone == "" {
		return ""
	}
	// Convert spaces to underscores
	zone = strings.ReplaceAll(zone, " ", "_")
	// Lowercase it all, and add a slash at the end
	return strings.ToLower(zone)
}

func ZoneToFolder(zone string) string {
	zone = ZoneNameSanitize(zone)
	// Lowercase it all, and add a slash at the end
	return zone + "/"
}

func ValidateZoneName(zone string) error {
	if zone == "" {
		return nil
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_ ]+$`).MatchString(zone) {
		return errors.New("allowable characters in zone name are letters, numbers, spaces, and underscores")
	}

	return nil
}

func FindZoneName(zone string) string {

	if _, ok := roomManager.zones[zone]; ok {
		return zone
	}

	for zoneName, _ := range roomManager.zones {
		if strings.Contains(strings.ToLower(zoneName), strings.ToLower(zone)) {
			return zoneName
		}
	}

	return ""
}

func GetZoneBiome(zone string) string {

	if z, ok := roomManager.zones[zone]; ok {
		return z.DefaultBiome
	}

	return ``
}

func MoveToZone(roomId int, newZoneName string) error {

	room, ok := roomManager.rooms[roomId]

	if !ok {
		return errors.New("room doesn't exist")
	}

	oldZoneName := room.Zone
	oldZoneInfo, ok := roomManager.zones[oldZoneName]
	if !ok {
		return errors.New("old zone doesn't exist")
	}
	oldFilePath := fmt.Sprintf("%s/%s", roomDataFilesPath, room.Filepath())

	newZoneInfo, ok := roomManager.zones[newZoneName]
	if !ok {
		return errors.New("new zone doesn't exist")
	}

	if oldZoneInfo.RootRoomId == roomId {
		return errors.New("can't move the root room of a zone")
	}

	room.Zone = newZoneName
	newFilePath := fmt.Sprintf("%s/%s", roomDataFilesPath, room.Filepath())

	if err := os.Rename(oldFilePath, newFilePath); err != nil {
		return err
	}

	delete(oldZoneInfo.RoomIds, roomId)
	roomManager.zones[oldZoneName] = oldZoneInfo

	newZoneInfo.RoomIds[roomId] = struct{}{}
	roomManager.zones[newZoneName] = newZoneInfo

	SaveRoom(*room)

	return nil
}

// #build zone The Arctic
// Build a zone, popualtes with an empty boring room
func CreateZone(zoneName string) (roomId int, err error) {

	zoneName = strings.TrimSpace(zoneName)

	if len(zoneName) < 2 {
		return 0, errors.New("zone name must be at least 2 characters")
	}

	if zoneInfo, ok := roomManager.zones[zoneName]; ok {

		return zoneInfo.RootRoomId, errors.New("zone already exists")
	}

	zoneFolder := util.FilePath(roomDataFilesPath, "/", ZoneToFolder(zoneName))
	if err := os.Mkdir(zoneFolder, 0755); err != nil {
		return 0, err
	}

	newRoom := NewRoom(zoneName)

	newRoom.ZoneConfig = ZoneConfig{RoomId: newRoom.RoomId}

	if err := newRoom.Validate(); err != nil {
		return 0, err
	}

	addRoomToMemory(newRoom)

	// save to the flat file
	SaveRoom(*newRoom)

	// write room to the folder under the new ID
	return newRoom.RoomId, nil
}

// #build room north
// Build a room to a specific direction, and connect it by exit name
// You still need to visit that room and connect it the opposite way
func BuildRoom(fromRoomId int, exitName string, mapDirection ...string) (room *Room, err error) {

	exitName = strings.TrimSpace(exitName)
	exitMapDirection := exitName

	if len(mapDirection) > 0 {
		exitMapDirection = mapDirection[0]
	}

	fromRoom := LoadRoom(fromRoomId)
	if fromRoom == nil {
		return nil, fmt.Errorf(`room %d not found`, fromRoomId)
	}

	newRoom := NewRoom(fromRoom.Zone)
	if newRoom != nil {
		newRoom.Validate()
	}

	newRoom.Title = fromRoom.Title

	if strings.HasPrefix(fromRoom.Description, `h:`) {
		hash := strings.TrimPrefix(fromRoom.Description, `h:`)
		if description, ok := roomManager.roomDescriptionCache[hash]; ok {
			newRoom.Description = description
		}
	} else {
		newRoom.Description = fromRoom.Description
	}

	newRoom.MapSymbol = fromRoom.MapSymbol
	newRoom.MapLegend = fromRoom.MapLegend
	newRoom.Biome = fromRoom.Biome

	if len(fromRoom.IdleMessages) > 0 {
		//newRoom.IdleMessages = fromRoom.IdleMessages
	}

	slog.Info("Connection room", "fromRoom", fromRoom.RoomId, "newRoom", newRoom.RoomId, "exitName", exitName)

	// connect the old room to the new room
	newExit := exit.RoomExit{RoomId: newRoom.RoomId, Secret: false}
	if exitMapDirection != exitName {
		newExit.MapDirection = exitMapDirection
	}
	fromRoom.Exits[exitName] = newExit

	//if _, ok := roomManager.rooms[newRoom.RoomId]; !ok {
	//	roomManager.rooms[newRoom.RoomId] = newRoom
	//}
	addRoomToMemory(newRoom)

	SaveRoom(*fromRoom)
	SaveRoom(*newRoom)

	return newRoom, nil
}

// #build exit north 1337
// Build an exit in the current room that links to room by id
// You still need to visit that room and connect it the opposite way
func ConnectRoom(fromRoomId int, toRoomId int, exitName string, mapDirection ...string) error {

	// exitname will be "north"
	exitName = strings.TrimSpace(exitName)
	exitMapDirection := exitName
	// Return direction will be "north" or "north-x2"
	if len(mapDirection) > 0 {
		exitMapDirection = mapDirection[0]
	}

	fromRoom := LoadRoom(fromRoomId)
	if fromRoom == nil {
		return fmt.Errorf(`room %d not found`, fromRoomId)
	}

	toRoom := LoadRoom(toRoomId)
	if toRoom == nil {
		return fmt.Errorf(`room %d not found`, toRoomId)
	}

	// connect the old room to the new room
	newExit := exit.RoomExit{RoomId: toRoom.RoomId, Secret: false}
	if exitMapDirection != exitName {
		newExit.MapDirection = exitMapDirection
	}
	fromRoom.Exits[exitName] = newExit

	SaveRoom(*fromRoom)

	return nil
}

func GetMapForDataString(dataStr string) string {
	// roomid:[1]/size:[wide/normal]/secrets:false/height:[18]/name:[Map of Frostfang]
	mapProperties := map[string]string{
		`roomid`:  ``,
		`size`:    `normal`, // wide?
		`height`:  `18`,
		`width`:   `65`,
		`name`:    `A useful map`,
		`secrets`: `false`,
		`markers`: ``,
	}
	mapDetails := strings.Split(dataStr, `/`)

	for _, mapDetail := range mapDetails {
		mapDetailParts := strings.Split(mapDetail, `=`)
		if len(mapDetailParts) == 2 {
			if _, ok := mapProperties[mapDetailParts[0]]; ok {
				mapProperties[mapDetailParts[0]] = mapDetailParts[1]
			}
		}
	}

	if len(mapDetails) > 0 {

		mapRoomId, _ := strconv.Atoi(mapProperties[`roomid`])
		mapSize := mapProperties[`size`]
		mapHeight, _ := strconv.Atoi(mapProperties[`height`])
		mapWidth, _ := strconv.Atoi(mapProperties[`width`])
		mapName := mapProperties[`name`]
		showAll, _ := strconv.ParseBool(mapProperties[`secrets`])

		mapMarkers := []string{mapProperties[`markers`]}

		return GetSpecificMap(mapRoomId, mapSize, mapHeight, mapWidth, mapName, showAll, mapMarkers)

	}
	return ""
}

func GetTinyMap(mapRoomId int) []string {

	result := [][]string{
		{` `, ` `, ` `, ` `, ` `},
		{` `, ` `, ` `, ` `, ` `},
		{` `, ` `, ` `, ` `, ` `},
		{` `, ` `, ` `, ` `, ` `},
		{` `, ` `, ` `, ` `, ` `},
	}

	originX := 2
	originY := 2

	result[originY][originX] = "@"

	if room := LoadRoom(mapRoomId); room != nil {

		var deltas directionDelta
		var ok bool

		for direction, exit := range room.Exits {
			if exit.Secret {
				continue
			}

			targetSymbol := `•`
			if targetRoom := LoadRoom(exit.RoomId); targetRoom != nil {
				if len(targetRoom.MapSymbol) > 0 {
					targetSymbol = targetRoom.MapSymbol
				}
				if len(targetRoom.players) > 0 || len(targetRoom.mobs) > 0 {
					targetSymbol = `⚠`
				}
			}

			if len(exit.MapDirection) > 0 {
				if deltas, ok = DirectionDeltas[exit.MapDirection]; !ok {
					continue
				}
			} else if deltas, ok = DirectionDeltas[direction]; !ok {
				continue
			}

			targetX := originX + (deltas.Dx * 2)
			stepX := 0
			if deltas.Dx < 0 {
				stepX = -1
			} else if deltas.Dx > 0 {
				stepX = 1
			}
			totalXSteps := stepX * (deltas.Dx * 2)

			targetY := originY + (deltas.Dy * 2)
			stepY := 0
			if deltas.Dy < 0 {
				stepY = -1
			} else if deltas.Dy > 0 {
				stepY = 1
			}
			totalYSteps := stepY * (deltas.Dy * 2)

			if stepX == 0 && stepY == 0 {
				continue
			}

			posX := originX
			posY := originY
			for totalXSteps > 0 || totalYSteps > 0 {
				if totalXSteps > 0 {
					totalXSteps--
				}
				if totalYSteps > 0 {
					totalYSteps--
				}
				posX += stepX
				posY += stepY

				if posY == originY && posX == originX {
					continue
				}
				// out of bounds
				if posY < 0 || posX < 0 || posY > len(result)-1 || posX > len(result[posY])-1 {
					continue
				}

				if posX == targetX && posY == targetY {
					result[posY][posX] = targetSymbol
				} else {
					result[posY][posX] = string(deltas.Arrow)
				}
			}

		}
	}

	returnResult := []string{}
	returnResult = append(returnResult, `╔═════╗`)
	for y := 0; y < len(result); y++ {
		returnResult = append(returnResult, `║`+strings.Join(result[y], ``)+`║`)
	}
	returnResult = append(returnResult, `╚═════╝`)

	return returnResult
}

func GetSpecificMap(mapRoomId int, mapSize string, mapHeight int, mapWidth int, mapName string, showSecrets bool, mapMarkers []string) string {

	mapMode := MapModeAllButSecrets
	if showSecrets {
		mapMode = MapModeAll
	}

	var mapData MapData
	var err error

	if mapRoom := LoadRoom(mapRoomId); mapRoom != nil {

		if mapSize == "wide" {

			rGraph := GenerateZoneMap(mapRoom.Zone, mapRoomId, 0, mapWidth, mapHeight, mapMode)

			if len(mapMarkers) > 0 {
				for _, overrideString := range mapMarkers {
					parts := strings.Split(overrideString, `,`)
					if len(parts) == 3 {
						roomId, _ := strconv.Atoi(parts[0])
						symbol := parts[1]
						legend := parts[2]

						if roomId > 0 && len(symbol) > 0 && len(legend) > 0 {
							rGraph.AddRoomSymbolOverrides([]rune(symbol)[0], legend, roomId)
						}
					}
				}
			}

			mapData, err = DrawZoneMap(rGraph, mapName, mapWidth, mapHeight)
		} else {

			rGraph := GenerateZoneMap(mapRoom.Zone, mapRoomId, 0, int(math.Ceil(float64(mapWidth)/2)), int(math.Ceil(float64(mapHeight)/2)), mapMode)

			if len(mapMarkers) > 0 {
				for _, overrideString := range mapMarkers {
					parts := strings.Split(overrideString, `,`)
					if len(parts) == 3 {
						roomId, _ := strconv.Atoi(parts[0])
						symbol := parts[1]
						legend := parts[2]

						if roomId > 0 && len(symbol) > 0 && len(legend) > 0 {
							rGraph.AddRoomSymbolOverrides([]rune(symbol)[0], legend, roomId)
						}
					}
				}
			}

			mapData, err = DrawZoneMap(rGraph, mapName, mapWidth, mapHeight)
		}

		if mapData.LegendWidth < 72 { // 80 - " Legend "
			mapData.LegendWidth = 72
		}

		if err != nil {
			slog.Error("Map Prop", "error", err.Error())
			return ``
		}

		mapTxt, _ := templates.Process("maps/map", mapData)
		return mapTxt
	}
	return ``
}

func GetRoomCount(zoneName string) int {

	zoneInfo, ok := roomManager.zones[zoneName]
	if !ok {
		return 0
	}

	return len(zoneInfo.RoomIds)
}

func LoadDataFiles() {

	if len(roomManager.zones) > 0 {
		slog.Info("rooms.LoadDataFiles()", "msg", "skipping reload of room files, rooms shouldn't be hot reloaded from flatfiles.")
		return
	}

	if err := loadAllRoomZones(); err != nil {
		panic(err)
	}

}
