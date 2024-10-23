package gametime

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/volte6/gomud/configs"
	"github.com/volte6/gomud/util"
)

var (
	dayResetOffset int = 0

	cachedRound uint64
	dateCache   GameDate
)

type GameDate struct {
	// The round number this GameDate represents
	RoundNumber      uint64
	RoundsPerDay     int
	NightHoursPerDay int

	Year        int
	Day         int
	Hour        int
	Hour24      int
	Minute      int
	MinuteFloat float64
	AmPm        string
	Night       bool

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
			return `<ansi fg="night">☾</ansi>` // •
		}
		return fmt.Sprintf(`<ansi fg="%s">☀️</ansi>`, dayNight) //
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

	gd := GameDate{
		RoundNumber:      currentRound,
		RoundsPerDay:     int(c.RoundsPerDay),
		NightHoursPerDay: int(c.NightHours),
	}

	gd.ReCalculate()

	return gd
}

func (g *GameDate) ReCalculate() {

	currentRoundAdjusted := (g.RoundNumber + uint64(dayResetOffset))
	roundOfDay := int(currentRoundAdjusted % uint64(g.RoundsPerDay))

	hourFloat, minutesFloat := math.Modf(float64(roundOfDay) / float64(g.RoundsPerDay) * 24)

	hour := int(hourFloat)
	hour24 := hour

	night := false
	halfNight := int(math.Floor(float64(g.NightHoursPerDay) / 2))
	nightStart := 24 - halfNight
	nightEnd := int(g.NightHoursPerDay) - halfNight
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

	minute := math.Floor(minutesFloat * 60)

	day := 1 + math.Floor(float64(currentRoundAdjusted)/float64(g.RoundsPerDay))
	year := math.Floor(float64(day) / 365)
	day -= math.Floor(year * 365)

	g.Day = int(day)
	g.Year = int(year)
	g.Hour = hour
	g.Hour24 = hour24
	g.Minute = int(minute)
	g.MinuteFloat = minutesFloat * 60
	g.AmPm = ampm
	g.Night = night

	g.NightStart = nightStart
	g.DayStart = nightEnd
}

func (g GameDate) AdjustTo(str string, adjustHours int, adjustDays int, adjustYears int) GameDate {

	if str == `hour` { // Start of the current hour

		g.RoundNumber -= uint64(math.Ceil(float64(g.MinuteFloat) * (float64(g.RoundsPerDay) / 24 / 60)))

	} else if str == `day` { // Start of current day

		g.RoundNumber -= uint64(math.Floor(float64(g.Hour24) * (float64(g.RoundsPerDay) / 24)))
		g.RoundNumber -= uint64(math.Ceil(float64(g.MinuteFloat) * (float64(g.RoundsPerDay) / 24 / 60)))

	} else if str == `week` { // Start of current week

		g.RoundNumber -= uint64(math.Floor(float64(g.Hour24) * (float64(g.RoundsPerDay) / 24)))
		g.RoundNumber -= uint64(math.Ceil(float64(g.MinuteFloat) * (float64(g.RoundsPerDay) / 24 / 60)))

	} else if str == `noon` { // 12pm of current day

		g.RoundNumber -= uint64(math.Floor(float64(g.Hour24) * (float64(g.RoundsPerDay) / 24)))
		g.RoundNumber -= uint64(math.Ceil(float64(g.MinuteFloat) * (float64(g.RoundsPerDay) / 24 / 60)))

		g.RoundNumber += uint64(math.Floor(float64(g.RoundsPerDay) / 2))

	}

	if adjustYears != 0 {
		if adjustYears < 1 {
			g.RoundNumber -= uint64(-1 * adjustYears * g.RoundsPerDay * 365)
		} else {
			g.RoundNumber += uint64(adjustYears * g.RoundsPerDay * 365)
		}
	}

	if adjustDays != 0 {
		if adjustDays < 1 {
			g.RoundNumber -= uint64(-1 * adjustDays * g.RoundsPerDay)
		} else {
			g.RoundNumber += uint64(adjustDays * g.RoundsPerDay)
		}
	}

	if adjustHours != 0 {
		if adjustHours < 1 {
			g.RoundNumber -= uint64(-1 * adjustHours * g.RoundsPerDay)
		} else {
			g.RoundNumber += uint64(adjustHours * g.RoundsPerDay)
		}
	}

	g.ReCalculate()

	return g
}

func StringToPeriods(currentRound uint64, str string) (lastRoundNum uint64, nextRoundNum uint64) {

	qty := 1
	timeStr := ``

	parts := strings.Split(str, ` `)
	if len(parts) == 1 {
		timeStr = parts[0]
	} else {
		if qty, _ = strconv.Atoi(parts[0]); qty == 0 {
			qty = 1
		}
		timeStr = parts[1]
	}

	g := getDate(currentRound)

	if timeStr == `year` || timeStr == `years` || timeStr == `yearly` {

		gLast := g.AdjustTo(`day`, 0, 0, -1)
		gNext := g.AdjustTo(`day`, 0, 0, 1)

		return gLast.RoundNumber, gNext.RoundNumber

	} else if timeStr == `week` || timeStr == `weeks` || timeStr == `weekly` {

		gLast := g.AdjustTo(`day`, 0, 0, -1)
		gNext := g.AdjustTo(`day`, 0, 0, 1)

		return gLast.RoundNumber, gNext.RoundNumber

	} else if timeStr == `day` || timeStr == `days` || timeStr == `daily` {

	}

	// assume hour/hours

	gLast := g.AdjustTo(`day`, 0, 0, -1)
	gNext := g.AdjustTo(`day`, 0, 0, 1)

	return gLast.RoundNumber, gNext.RoundNumber

}
