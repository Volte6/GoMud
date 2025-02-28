package web

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"text/template"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/mutators"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
)

type ZoneDetails struct {
	ZoneName  string
	RoomCount int
	AutoScale string
}

func roomsIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String()+"/admin/_header.html", configs.GetConfig().FolderHtmlFiles.String()+"/admin/rooms/index.html", configs.GetConfig().FolderHtmlFiles.String()+"/admin/_footer.html")
	if err != nil {
		mudlog.Error("HTML Template", "error", err)
	}

	qsp := r.URL.Query()

	filterType := qsp.Get(`filter-type`)

	type shortRoomInfo struct {
		RoomId          int
		RoomZone        string
		ZoneRoot        bool
		RoomTitle       string
		IsBank          bool
		IsStorage       bool
		IsCharacterRoom bool
		IsSkillTraining bool
		HasContainer    bool
		IsPvp           bool
	}

	allZones := []ZoneDetails{}
	allRooms := []shortRoomInfo{}
	zoneCounter := map[string]int{}

	for _, rId := range rooms.GetAllRoomIds() {
		if room := rooms.LoadRoom(rId); room != nil {

			if _, ok := zoneCounter[room.Zone]; !ok {

				autoScale := ``

				if rootRoomId, err := rooms.GetZoneRoot(room.Zone); err == nil {
					if rootRoom := rooms.LoadRoom(rootRoomId); rootRoom != nil {
						if rootRoom.ZoneConfig.MobAutoScale.Minimum > 0 || rootRoom.ZoneConfig.MobAutoScale.Maximum > 0 {
							autoScale = fmt.Sprintf(`%d to %d`, rootRoom.ZoneConfig.MobAutoScale.Minimum, rootRoom.ZoneConfig.MobAutoScale.Maximum)
						}
					}
				}

				zoneCounter[room.Zone] = 0
				allZones = append(allZones, ZoneDetails{
					ZoneName:  room.Zone,
					RoomCount: 0,
					AutoScale: autoScale,
				})
			}
			zoneCounter[room.Zone] = zoneCounter[room.Zone] + 1

			if filterType != `*` && filterType != room.Zone {
				continue
			}

			hasContainer := false

			for _, cInfo := range room.Containers {
				if cInfo.DespawnRound == 0 {
					hasContainer = true
					break
				}
			}

			allRooms = append(allRooms, shortRoomInfo{
				RoomId:          room.RoomId,
				RoomZone:        room.Zone,
				ZoneRoot:        room.ZoneConfig.RoomId == room.RoomId,
				RoomTitle:       room.Title,
				IsBank:          room.IsBank,
				IsStorage:       room.IsStorage,
				IsCharacterRoom: room.IsCharacterRoom,
				IsSkillTraining: len(room.SkillTraining) > 0,
				HasContainer:    hasContainer,
				IsPvp:           room.IsPvp(),
			})
		}
	}

	for i, zInfo := range allZones {
		zInfo.RoomCount = zoneCounter[zInfo.ZoneName]
		allZones[i] = zInfo
	}

	sort.SliceStable(allRooms, func(i, j int) bool {

		if allRooms[i].RoomZone != allRooms[j].RoomZone {
			return allRooms[i].RoomZone < allRooms[j].RoomZone
		}

		if allRooms[i].ZoneRoot {
			return true
		} else if allRooms[j].ZoneRoot {
			return false
		}

		return allRooms[i].RoomId < allRooms[j].RoomId
	})

	sort.SliceStable(allZones, func(i, j int) bool {
		return allZones[i].ZoneName < allZones[j].ZoneName
	})

	tplData := map[string]any{
		`Zones`:      allZones,
		`Rooms`:      allRooms,
		`FilterType`: filterType,
	}

	if err := tmpl.Execute(w, tplData); err != nil {
		mudlog.Error("HTML Execute", "error", err)
	}

}

func roomData(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("room.data.html").Funcs(funcMap).ParseFiles(configs.GetConfig().FolderHtmlFiles.String() + "/admin/rooms/room.data.html")
	if err != nil {
		mudlog.Error("HTML Template", "error", err)
	}

	urlVals := r.URL.Query()

	roomIdInt, _ := strconv.Atoi(urlVals.Get(`roomid`))

	roomInfo := rooms.LoadRoom(roomIdInt)

	tplData := map[string]any{}
	tplData[`roomInfo`] = roomInfo

	buffSpecs := []buffs.BuffSpec{}
	for _, buffId := range buffs.GetAllBuffIds() {
		if b := buffs.GetBuffSpec(buffId); b != nil {
			if b.Name == `empty` {
				continue
			}
			buffSpecs = append(buffSpecs, *b)
		}
	}
	sort.SliceStable(buffSpecs, func(i, j int) bool {
		return buffSpecs[i].BuffId < buffSpecs[j].BuffId
	})
	tplData[`buffSpecs`] = buffSpecs

	allBiomes := rooms.GetAllBiomes()
	sort.SliceStable(allBiomes, func(i, j int) bool {
		return allBiomes[i].Name() < allBiomes[j].Name()
	})
	tplData[`biomes`] = allBiomes

	allSkillNames := []string{}
	for _, name := range skills.GetAllSkillNames() {
		allSkillNames = append(allSkillNames, string(name))
	}
	sort.SliceStable(allSkillNames, func(i, j int) bool {
		return allSkillNames[i] < allSkillNames[j]
	})
	tplData[`allSkillNames`] = allSkillNames

	tplData[`allSlotTypes`] = characters.GetAllSlotTypes()

	mapDirections := []string{}

	for name := range rooms.DirectionDeltas {
		mapDirections = append(mapDirections, name)
	}
	sort.SliceStable(mapDirections, func(i, j int) bool {
		return mapDirections[i] < mapDirections[j]
	})
	tplData[`mapDirections`] = mapDirections

	mutSpecs := mutators.GetAllMutatorSpecs()
	sort.SliceStable(mutSpecs, func(i, j int) bool {
		return mutSpecs[i].MutatorId < mutSpecs[j].MutatorId
	})
	tplData[`mutSpecs`] = mutSpecs

	if err := tmpl.Execute(w, tplData); err != nil {
		mudlog.Error("HTML Execute", "error", err)
	}

}
