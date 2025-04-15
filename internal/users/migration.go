package users

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (
	ErrBothFilesExist = errors.New("could not migrate due to both file formats existing")
)

//
// Handles user file migration
//

func DoUserMigrations() {

	DoFilenameMigrationV1()

}

func DoFilenameMigrationV1() error {

	var errorResult error = nil

	SearchOfflineUsers(func(u *UserRecord) bool {

		oldUserPath := util.FilePath(string(configs.GetFilePathsConfig().DataFiles), `/`, `users`, `/`, strings.ToLower(u.Username)+`.yaml`)
		newUserPath := util.FilePath(string(configs.GetFilePathsConfig().DataFiles), `/`, `users`, `/`, strconv.Itoa(u.UserId)+`.yaml`)

		_, err := os.Stat(oldUserPath)
		oldUserPathExists := err == nil

		_, err = os.Stat(newUserPath)
		newUserPathExists := err == nil

		if oldUserPathExists && newUserPathExists {
			errorResult = ErrBothFilesExist
			return false
		}

		if oldUserPathExists {
			err := os.Rename(oldUserPath, newUserPath)
			if err != nil {
				errorResult = err
				return false
			}
		}

		oldAltsFilePath := util.FilePath(string(configs.GetFilePathsConfig().DataFiles), `/users/`, strings.ToLower(u.Username)+`-alts.yaml`)
		newAltsFilePath := util.FilePath(string(configs.GetFilePathsConfig().DataFiles), `/users/`, strconv.Itoa(u.UserId)+`.alts.yaml`)

		_, err = os.Stat(oldAltsFilePath)
		oldAltsPathExists := err == nil

		_, err = os.Stat(newAltsFilePath)
		newAltsPathExists := err == nil

		if oldAltsPathExists && newAltsPathExists {
			errorResult = ErrBothFilesExist
			return false
		}

		if oldAltsPathExists {
			err := os.Rename(oldAltsFilePath, newAltsFilePath)
			if err != nil {
				errorResult = err
				return false
			}
		}

		return true
	})

	return errorResult
}
