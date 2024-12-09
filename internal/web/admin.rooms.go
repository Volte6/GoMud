package web

import (
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"text/template"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/races"
	"github.com/volte6/gomud/internal/rooms"
)

func roomsIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("_datafiles/html/admin/_header.html", "_datafiles/html/admin/rooms/index.html", "_datafiles/html/admin/_footer.html")
	if err != nil {
		slog.Error("HTML Template", "error", err)
	}

	type shortRoomInfo struct {
		RoomId    int
		RoomZone  string
		ZoneRoot  bool
		RoomTitle string
	}

	allRooms := []shortRoomInfo{}

	for _, rId := range rooms.GetAllRoomIds() {
		if room := rooms.LoadRoom(rId); room != nil {
			allRooms = append(allRooms, shortRoomInfo{
				RoomId:    room.RoomId,
				RoomZone:  room.Zone,
				ZoneRoot:  room.ZoneConfig.RoomId == room.RoomId,
				RoomTitle: room.Title,
			})
		}
	}

	sort.SliceStable(allRooms, func(i, j int) bool {
		if allRooms[i].RoomZone != allRooms[j].RoomZone {
			return allRooms[i].RoomZone < allRooms[j].RoomZone
		}

		return allRooms[i].RoomId < allRooms[j].RoomId
	})

	if err := tmpl.Execute(w, allRooms); err != nil {
		slog.Error("HTML Execute", "error", err)
	}

}

func roomData(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("room.data.html").Funcs(funcMap).ParseFiles("_datafiles/html/admin/rooms/room.data.html")
	if err != nil {
		slog.Error("HTML Template", "error", err)
	}

	urlVals := r.URL.Query()

	raceIdInt, _ := strconv.Atoi(urlVals.Get(`raceid`))

	raceInfo := races.GetRace(raceIdInt)
	if raceInfo == nil {
		raceInfo = &races.Race{}
	}

	tplData := map[string]any{}
	tplData[`raceInfo`] = *raceInfo

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

	tplData[`allSlotTypes`] = characters.GetAllSlotTypes()

	if err := tmpl.Execute(w, tplData); err != nil {
		slog.Error("HTML Execute", "error", err)
	}

}
