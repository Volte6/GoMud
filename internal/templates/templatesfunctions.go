package templates

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/language"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

var (
	funcMap = template.FuncMap{
		"pad":       pad,
		"padLeft":   padLeft,
		"padRight":  padRight,
		"padRightX": padRightX,
		"join":      join,
		"lte":       func(a, b int) bool { return a <= b },
		"gte":       func(a, b int) bool { return a >= b },
		"lt":        func(a, b int) bool { return a < b },
		"uc":        func(s string) string { return strings.Title(s) },
		"firstletter": func(a string) string {
			if len(a) == 0 {
				return ""
			}
			return string(a[0])
		},
		"repeat": func(what string, count int) string {
			if count < 1 {
				return ""
			}
			return strings.Repeat(what, count)
		},
		"add": func(start int, num ...int) int {
			for _, n := range num {
				start += n
			}
			return start
		},
		"sub": func(start int, num ...int) int {
			for _, n := range num {
				start -= n
			}
			return start
		},
		"mult": func(start int, num ...int) int {
			for _, n := range num {
				start *= n
			}
			return start
		},
		"multfloat": func(start float32, num ...float32) float32 {
			for _, n := range num {
				start *= n
			}
			return start
		},
		"cellpad": func(padding int, s string) string {
			if padding < 1 {
				return s
			}
			pad := strings.Repeat(" ", padding)
			return pad + s + pad
		},
		"lowercase": func(s string) string {
			return strings.ToLower(s)
		},
		"mapkeys": func(m map[string]string) []string {
			keys := make([]string, len(m))
			i := 0
			for k := range m {
				keys[i] = k
				i++
			}
			return keys
		},
		"intmapkeys": func(m map[string]int) []string {
			keys := make([]string, len(m))
			i := 0
			for k := range m {
				keys[i] = k
				i++
			}
			return keys
		},
		"uidToCharacter": func(userId int) *characters.Character {
			u := users.GetByUserId(userId)
			return u.Character
		},
		"idToMobCharacter": func(mobId int) string {
			if m := mobs.GetInstance(mobId); m != nil {
				return m.Character.Name
			}
			return "glorble"
		},
		"idsOtherThan": idsOtherThan,
		"tnl":          TNL,
		"pct":          pct,
		"numberFormat": numberFormat,
		"mod":          func(a, b int) int { return a % b },
		"stringor":     stringOr,
		"splitstring":  SplitStringNL,
		"ansiparse":    TplAnsiParse,
		"buffname": func(buffId int) string {
			buffSpec := buffs.GetBuffSpec(buffId)
			if buffSpec == nil {
				return "Unknown"
			}
			return buffSpec.Name
		},
		"buffduration": func(buffId int) string {
			buffSpec := buffs.GetBuffSpec(buffId)
			if buffSpec == nil {
				return "Unknown"
			}

			if buffSpec.RoundInterval == 1 && buffSpec.TriggerCount == 1 {
				return `Activates once`
			}

			var roundCt string
			if buffSpec.RoundInterval > 1 {
				roundCt = fmt.Sprintf(`%d rounds`, buffSpec.RoundInterval)
			} else {
				roundCt = `round`
			}
			return fmt.Sprintf("Activates every %s (%dx total)", roundCt, buffSpec.TriggerCount)
		},
		"formatdiceroll": func(roll string) string {
			a, d, s, b, _ := util.ParseDiceRoll(roll)
			return util.FormatDiceRoll(a, d, s, b, []int{})
		},
		"profession": func(char characters.Character) string {

			allRanks := char.GetAllSkillRanks()
			return skills.GetProfession(allRanks)
		},
		"roundstotime": func(rounds int) string {
			if rounds >= buffs.TriggersLeftUnlimited {
				return `Unlimited`
			}
			return formatDuration(rounds * int(configs.GetTimingConfig().RoundSeconds))
		},
		"secondsFrom": func(t time.Time) int {
			// return the number of seconds unti the given time
			return int(time.Until(t).Seconds())
		},
		"intstrlen": func(i ...int) int {
			totalLen := 0
			for _, n := range i {
				totalLen += len(strconv.Itoa(n))
			}
			return totalLen
		},
		"healthStr": func(h int, hMax int, padTo ...int) string {
			padding := ``
			if len(padTo) > 0 {
				padding = strings.Repeat(" ", padTo[0]-(len(strconv.Itoa(h))+len(strconv.Itoa(hMax))+1))
			}
			hLevel := util.QuantizeTens(h, hMax)
			return fmt.Sprintf(`<ansi fg="health-%d">%d</ansi>/<ansi fg="health-%d">%d</ansi>%s`, hLevel, h, hLevel, hMax, padding)
		},
		"manaStr": func(m int, mMax int, padTo ...int) string {

			padding := ``
			if len(padTo) > 0 {
				padding = strings.Repeat(" ", padTo[0]-(len(strconv.Itoa(m))+len(strconv.Itoa(mMax))+1))
			}

			mLevel := util.QuantizeTens(m, mMax)
			return fmt.Sprintf(`<ansi fg="mana-%d">%d</ansi>/<ansi fg="mana-%d">%d</ansi>%s`, mLevel, m, mLevel, mMax, padding)
		},
		"colorpattern": func(s string, pattern string, colorizeStyle ...string) string {
			style := ``
			if len(colorizeStyle) > 0 {
				style = colorizeStyle[0]
			}

			if style == `words` {
				return colorpatterns.ApplyColorPattern(s, pattern, colorpatterns.Words)
			} else if style == `stretch` {
				return colorpatterns.ApplyColorPattern(s, pattern, colorpatterns.Stretch)
			}
			return colorpatterns.ApplyColorPattern(s, pattern)
		},
		"permadeath": func() bool {
			return bool(configs.GetGamePlayConfig().Death.PermaDeath)
		},
		"zodiac": func(year int) string {
			return gametime.GetZodiac(year)
		},
		"month": func(month int) string {
			return gametime.MonthName(month)
		},
		"map": makeMap,
		"t":   language.T,
	}
)

// Usage:
//
//	{{ implode .items "," }}
//		OUTPUT: "item1,item2,item3"
func join(items []string, sep string) string {
	return strings.Join(items, sep)
}

// Usage:
//
//	{{ padLeft 10 }}
//		OUTPUT: "		  "
//	{{ padLeft 10 "hello" "-" }}
//		OUTPUT: "-----hello"
func padLeft(totalWidth int, stringArgs ...string) string {
	var stringIn string = ""
	var padString string = " "

	if len(stringArgs) > 0 {
		stringIn = stringArgs[0]
		if len(stringArgs) > 1 {
			padString = stringArgs[1]
		}
	}

	stringInWidth := runewidth.StringWidth(stringIn)

	if stringInWidth >= totalWidth {
		return stringIn
	}
	paddingLength := totalWidth - stringInWidth
	if paddingLength < 1 {
		return stringIn
	}
	return strings.Repeat(padString, paddingLength) + stringIn
}

// Usage:
//
//	{{ padRight 10 }}
//		OUTPUT: "		  "
//	{{ padRight 10 "hello" "-" }}
//		OUTPUT: "hello-----"
func padRight(totalWidth int, stringArgs ...string) string {
	var stringIn string = ""
	var padString string = " "

	if len(stringArgs) > 0 {
		stringIn = stringArgs[0]
		if len(stringArgs) > 1 {
			padString = stringArgs[1]
		}
	}

	stringInWidth := runewidth.StringWidth(stringIn)

	if stringInWidth >= totalWidth {
		return stringIn
	}
	paddingLength := totalWidth - stringInWidth
	if paddingLength < 1 {
		return stringIn
	}
	return stringIn + strings.Repeat(padString, paddingLength)
}

func padRightX(input, padding string, length int) string {

	padLen := runewidth.StringWidth(padding)
	inputLen := runewidth.StringWidth(input)

	if length < inputLen {
		length = inputLen
	}

	// Calculate how many times the padding string should be repeated
	paddingRepeats := int(math.Ceil((float64(length) - float64(inputLen)) / float64(padLen)))
	finalPadLength := length - inputLen

	// Repeat the padding string to fill the gap
	pad := strings.Repeat(padding, paddingRepeats)

	if runewidth.StringWidth(pad)+inputLen > length {
		pad = string([]rune(pad)[:finalPadLength])
	}

	// Trim the padded string to the desired length

	// Concatenate the input and the padding
	result := input + pad

	return result
}

// Usage:
//
//	{{ pad 10 }}
//		OUTPUT: "          "
//	{{ pad 11 "hello" "-" }}
//		OUTPUT: "---hello---"
func pad(totalWidth int, stringArgs ...string) string {
	var stringIn string = ""
	var padString string = " "

	if len(stringArgs) > 0 {
		stringIn = stringArgs[0]
		if len(stringArgs) > 1 {
			padString = stringArgs[1]
		}
	}

	stringInWidth := runewidth.StringWidth(stringIn)

	if stringInWidth >= totalWidth {
		return stringIn
	}
	paddingLength := totalWidth - stringInWidth
	leftPad := paddingLength >> 1
	if leftPad < 1 {
		return stringIn
	}
	if paddingLength-leftPad < 1 {
		return strings.Repeat(padString, leftPad) + stringIn
	}
	return strings.Repeat(padString, leftPad) + stringIn + strings.Repeat(padString, paddingLength-leftPad)
}

func idsOtherThan(allIds []uint64, excludeId uint64) []uint64 {
	finalIds := make([]uint64, 0)
	for _, id := range allIds {
		if id != excludeId {
			finalIds = append(finalIds, id)
		}
	}
	return finalIds
}

func TNL(userId int) string {
	user := users.GetByUserId(userId)
	realXPNow, realXPTNL := user.Character.XPTNLActual()
	return fmt.Sprintf(`%d/%d (%d%%)`, realXPNow, realXPTNL, pct(realXPNow, realXPTNL))
}

func pct(a, b int) int {
	return (a * 100) / b
}

func numberFormat(num int) string {
	return util.FormatNumber(num)
}

func TplAnsiParse(input string) string {
	return AnsiParse(input)
}

func stringOr(a string, b string, padding ...int) string {
	str := a
	if str == "" {
		str = b
	}
	if len(padding) > 0 {
		str = padRight(padding[0], str)
	}
	return str
}

// Splits a string by adding line breaks at the end of each line
func SplitStringNL(input string, lineWidth int, nlPrefix ...string) string {

	output := strings.Builder{}

	words := strings.Fields(input) // Split the input into words

	linePrefix := ""
	if len(nlPrefix) > 0 {
		linePrefix = nlPrefix[0]
	}

	currentLine := ""
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= lineWidth { // +1 for the space
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			if linePrefix != "" && output.Len() > 0 {
				output.WriteString(linePrefix)
			}
			output.WriteString(currentLine)
			output.WriteString(term.CRLFStr)
			currentLine = word
		}
	}

	if currentLine != "" {
		if linePrefix != "" && output.Len() > 0 {
			output.WriteString(linePrefix)
		}
		output.WriteString(currentLine)
	}

	return output.String()
}

func formatDuration(seconds int) string {
	days := seconds / (24 * 3600)
	hours := (seconds % (24 * 3600)) / 3600
	minutes := (seconds % 3600) / 60
	seconds = seconds % 60

	result := ""
	if days > 0 {
		if days == 1 {
			result += strconv.Itoa(days) + " day "
		} else {
			result += strconv.Itoa(days) + " days "
		}
	}
	if hours > 0 || days > 0 { // Include hours if there are any days
		if hours == 1 {
			result += strconv.Itoa(hours) + " hour "
		} else {
			result += strconv.Itoa(hours) + " hours "
		}
	}
	if minutes > 0 || hours > 0 || days > 0 { // Include minutes if there are any hours or days
		if minutes == 1 {
			result += strconv.Itoa(minutes) + " minute "
		} else {
			result += strconv.Itoa(minutes) + " minutes "
		}
	}

	if seconds > 0 {
		if seconds == 1 {
			result += strconv.Itoa(seconds) + " second"
		} else {
			result += strconv.Itoa(seconds) + " seconds"
		}
	}

	if result == `` {
		result = `0 seconds`
	}

	return strings.TrimSpace(result)
}

func makeMap(kvs ...any) map[any]any {
	m := make(map[any]any)
	for i := 0; i < len(kvs)-1; i += 2 {
		m[kvs[i]] = kvs[i+1]
	}

	return m
}
