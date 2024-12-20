package mobs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/fileloader"
	"github.com/volte6/gomud/internal/util"
)

func CreateNewMobFile(newMobName string, newMobRaceId int, newMobZone string, newMobDescription string, includeScript bool) (MobId, error) {

	newMobInfo := Mob{
		MobId: getNextMobId(),
		Zone:  newMobZone,
		Character: characters.Character{
			Name:        newMobName,
			RaceId:      newMobRaceId,
			Description: strings.ReplaceAll(newMobDescription, `\n`, "\n"),
		},
	}

	if includeScript {
		newMobInfo.QuestFlags = []string{`1000000-start`}
	}

	if err := newMobInfo.Validate(); err != nil {
		return 0, err
	}

	if newMobInfo.MobId == 0 {
		return 0, errors.New(`Could not find a new mob id to assign.`)
	}

	saveModes := []fileloader.SaveOption{}

	if configs.GetConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	if err := fileloader.SaveFlatFile[*Mob](mobDataFilesFolderPath, &newMobInfo, saveModes...); err != nil {
		return 0, err
	}

	// Save to in-memory cache
	allMobNames = append(allMobNames, newMobInfo.Character.Name)
	mobNameCache[newMobInfo.MobId] = newMobInfo.Character.Name
	mobs[newMobInfo.Id()] = &newMobInfo

	if includeScript {

		newScriptPath := newMobInfo.GetScriptPath()
		os.MkdirAll(filepath.Dir(newScriptPath), os.ModePerm)

		fileloader.CopyFileContents(
			util.FilePath(`_datafiles/mobs/sample-quest-mob-script.js`),
			newMobInfo.GetScriptPath(),
		)
	}

	return newMobInfo.MobId, nil
}

func getNextMobId() MobId {

	lowestFreeId := MobId(0)
	for _, mInfo := range mobs {

		if mInfo.MobId >= lowestFreeId {
			lowestFreeId = mInfo.MobId + 1
		}
	}

	return lowestFreeId
}
