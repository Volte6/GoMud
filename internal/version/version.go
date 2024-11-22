package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/configs"
)

// Responsible for version relate operations such as
// upgrading config files

const (
	MAJOR = 0
	MINOR = 1
	PATCH = 2
)

var (
	ErrIncompatibleVersion error = errors.New(`incompatible version`)
	ErrUpgradePossible     error = errors.New(`upgrade possible`)
	ErrCannotUpgrade       error = errors.New(`upgrade not possible`)
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v *Version) String() string {
	return fmt.Sprintf(`%d.%d.%d`, v.Major, v.Minor, v.Patch)
}

func (v *Version) Parse(versionValue string) error {
	versionParts := strings.Split(versionValue, `.`)

	if len(versionParts) != 3 {
		return errors.New(`invalid version number: ` + versionValue)
	}

	v.Major, _ = strconv.Atoi(versionParts[MAJOR])
	v.Minor, _ = strconv.Atoi(versionParts[MINOR])
	v.Patch, _ = strconv.Atoi(versionParts[PATCH])

	return nil
}

func (v *Version) Compatible(other Version) bool {
	return v.Major == other.Major
}

func (v *Version) Upgradable(other Version) bool {
	return v.Minor <= other.Minor || v.Patch < other.Patch
}

func VersionCheck(version string) error {

	binVersion := Version{}
	if err := binVersion.Parse(version); err != nil {
		panic(err)
	}

	cfg := configs.GetConfig()

	cfgVersion := Version{}
	if err := cfgVersion.Parse(string(cfg.Version)); err != nil {
		panic(err)
	}

	if !binVersion.Compatible(cfgVersion) {
		return ErrIncompatibleVersion
	}

	if binVersion.Upgradable(cfgVersion) {
		return ErrUpgradePossible
	}

	return nil
}

func UpgradeDatafiles(v string) error {

	targetVersion := Version{}
	if err := targetVersion.Parse(v); err != nil {
		return err
	}

	cfg := configs.GetConfig()

	currentVersion := Version{}
	if err := currentVersion.Parse(string(cfg.Version)); err != nil {
		return err
	}

	if !currentVersion.Upgradable(targetVersion) {
		return ErrCannotUpgrade
	}

	// TODO: upgrade datafiles

	return nil
}
