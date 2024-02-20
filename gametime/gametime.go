package gametime

import (
	"fmt"
	"math"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/util"
)

var (
	dayResetOffset int = 0

	cachedRound uint64
	dateCache   GameDate
)

type GameDate struct {
	Day    int
	Hour   int
	Hour24 int
	Minute int
	AmPm   string
	Night  bool

	DayStart   int
	NightStart int
}

func (gd GameDate) String(symbolOnly ...bool) string {

	dayNight := `day`
	if gd.Night {
		dayNight = `night`
	} else {
		hoursLeft := int(math.Abs(float64(gd.Hour24) - float64(gd.NightStart)))
		if hoursLeft < 3 {
			dayNight = `day-dusk`
		}
	}

	if len(symbolOnly) > 0 && symbolOnly[0] {

		if gd.Night {
			return `<ansi fg="night">•</ansi>`
		}
		return fmt.Sprintf(`<ansi fg="%s">⚙</ansi>`, dayNight)
	}

	return fmt.Sprintf("<ansi fg=\"%s\">%d:%02d%s</ansi>", dayNight, gd.Hour, gd.Minute, gd.AmPm)
}

// Jumps the clock foward to the next night
// If a roundAdjustment is provided, it will be added to the offset
// This is useful to set to the round right before the rollover
func SetToNight(roundAdjustment ...int) {
	c := configs.GetConfig()

	roundsPerHour := float64(c.RoundsPerDay) / 24
	halfNight := math.Floor(float64(c.NightHours) / 2)

	dayResetOffset = int((24 - halfNight) * roundsPerHour)

	roundOfDay := int(util.GetRoundCount() % uint64(c.RoundsPerDay))
	dayResetOffset -= roundOfDay
	if len(roundAdjustment) > 0 {
		dayResetOffset += roundAdjustment[0]
	}

	// Reset the cache
	cachedRound = 0
}

// Jumps the clock forward to the next day
// If a roundAdjustment is provided, it will be added to the offset
// This is useful to set to the round right before the rollover
func SetToDay(roundAdjustment ...int) {
	c := configs.GetConfig()

	roundsPerHour := float64(c.RoundsPerDay) / 24
	halfNight := int(math.Ceil(float64(c.NightHours) / 2))

	dayResetOffset = int(float64(halfNight) * roundsPerHour)

	roundOfDay := int(util.GetRoundCount() % uint64(c.RoundsPerDay))
	dayResetOffset -= roundOfDay
	if len(roundAdjustment) > 0 {
		dayResetOffset += roundAdjustment[0]
	}

	// Reset the cache
	cachedRound = 0
}

// Jumps the clock forward a specific hour/minutes
// Between 0 and 23
func SetTime(setToHour int, setToMinutes ...int) {

	c := configs.GetConfig()

	setToHour = setToHour % 24
	roundsPerHour := float64(c.RoundsPerDay) / 24
	dayResetOffset = int(math.Floor(float64(setToHour) * roundsPerHour))
	if len(setToMinutes) > 0 {
		dayResetOffset += int(math.Ceil((float64(setToMinutes[0]) / 60) * roundsPerHour))
	}

	roundOfDay := int(util.GetRoundCount() % uint64(c.RoundsPerDay))
	dayResetOffset -= roundOfDay

	// Reset the cache
	cachedRound = 0
}

func IsNight() bool {
	gd := GetDate()
	return gd.Night
}

// Gets the details of the current date
func GetDate() GameDate {

	currentRound := util.GetRoundCount()

	if cachedRound == 0 || currentRound != cachedRound {
		// Update the cache info
		dateCache = getDate(currentRound)
		cachedRound = currentRound
	}

	return dateCache
}

func getDate(currentRound uint64) GameDate {

	c := configs.GetConfig()

	currentRoundAdjusted := (currentRound + uint64(dayResetOffset))
	roundOfDay := int(currentRoundAdjusted % uint64(c.RoundsPerDay))

	hourFloat, minutesFloat := math.Modf(float64(roundOfDay) / float64(c.RoundsPerDay) * 24)

	hour := int(hourFloat)
	hour24 := hour

	night := false
	halfNight := int(math.Floor(float64(c.NightHours) / 2))
	nightStart := 24 - halfNight
	nightEnd := c.NightHours - halfNight
	if hour >= nightStart || hour < nightEnd {
		night = true
	}

	ampm := `AM`
	if hour >= 12 {
		ampm = `PM`
		hour -= 12
	}

	if hour == 0 {
		hour = 12
	}

	minute := int(math.Floor(minutesFloat * 60))

	day := 1 + int(math.Floor(float64(currentRoundAdjusted)/float64(c.RoundsPerDay)))

	return GameDate{
		Day:    day,
		Hour:   hour,
		Hour24: hour24,
		Minute: minute,
		AmPm:   ampm,
		Night:  night,

		NightStart: nightStart,
		DayStart:   nightEnd,
	}
}
