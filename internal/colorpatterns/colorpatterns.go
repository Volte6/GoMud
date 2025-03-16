package colorpatterns

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Volte6/ansitags"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mudlog"
	"gopkg.in/yaml.v2"
)

// TODO: Load patterns from a config file.
type ColorizeStyle uint8

var (
	Default ColorizeStyle = 0
	Words   ColorizeStyle = 1
	Once    ColorizeStyle = 2
	Stretch ColorizeStyle = 3

	colorsCompiled bool = false

	numericPatterns = map[string][]int{}
	// Short tags
	ShortTagPatterns = map[string][]string{}
)

func GetColorPatternNames() []string {

	if !colorsCompiled {
		CompileColorPatterns()
	}

	ret := []string{}

	for name := range numericPatterns {
		ret = append(ret, name)
	}

	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })

	return ret
}

func ApplyColorPattern(input string, pattern string, method ...ColorizeStyle) string {
	if pattern == `` {
		return input
	}
	patternValues, ok := numericPatterns[pattern]
	if !ok {
		return input
	}

	return ApplyColors(input, patternValues, method...)
}

func ApplyColors(input string, patternValues []int, method ...ColorizeStyle) string {

	patternValueLength := len(patternValues)

	newString := strings.Builder{}

	patternDir := 1
	patternPosition := 0
	inTagPlaceholder := false

	//
	// Tokenize existing ansi tags to avoid colorizing them
	//
	// Regular expression to match <ansi ...>...</ansi> tags
	re := regexp.MustCompile(`<ansi[^>]*>.*?</ansi>`)
	// Counter to keep track of the unique numbers
	counter := 0
	preExistingTags := map[string]string{}
	// Function to replace each match with a unique number
	input = re.ReplaceAllStringFunc(input, func(match string) string {
		counter++
		tag := `:` + strconv.Itoa(counter)
		preExistingTags[tag] = match
		return tag
	})
	//
	// End tokenization
	//

	if len(method) == 0 || method[0] == Default {
		// Color change on a per character basis (not spaces), reverses at the end
		for _, runeChar := range input {

			// Handle placeholder tags that look like :123
			if inTagPlaceholder {
				if runeChar != 32 {
					newString.WriteString(string(runeChar))
					continue
				}
				inTagPlaceholder = false
			} else {
				if runeChar == ':' {
					inTagPlaceholder = true
					newString.WriteString(string(runeChar))
					continue
				}
			}

			newString.WriteString(fmt.Sprintf(`<ansi fg="%d">%s</ansi>`, patternValues[patternPosition], string(runeChar)))
			if runeChar != 32 { // space
				if patternPosition == patternValueLength-1 {
					patternDir = -1
				} else if patternPosition == 0 {
					patternDir = 1
				}
				patternPosition += patternDir // advance the color token position
			}
		}
	} else if method[0] == Words {
		// Color change on a per word basis

		newString.WriteString(`<ansi>`)
		for i, runeChar := range input {

			// Handle placeholder tags that look like :123
			if inTagPlaceholder {
				if runeChar != 32 {
					newString.WriteString(string(runeChar))
					continue
				}
				inTagPlaceholder = false
			} else {
				if runeChar == ':' {
					inTagPlaceholder = true
					newString.WriteString(string(runeChar))
					continue
				}
			}
			// End handling placeholder tags

			if i == 0 || runeChar == 32 { // space
				newString.WriteString(fmt.Sprintf(`</ansi><ansi fg="%d">`, patternValues[patternPosition%patternValueLength]))
				patternPosition++ // advance the color token position
			}
			newString.WriteRune(runeChar) // Write whatever the next character is
		}
		newString.WriteString(`</ansi>`)
	} else if method[0] == Once {
		// Color stops changing and stays on the final color
		newString.WriteString(`<ansi>`)
		for _, runeChar := range input {

			// Handle placeholder tags that look like :123
			if inTagPlaceholder {
				if runeChar != 32 {
					newString.WriteString(string(runeChar))
					continue
				}
				inTagPlaceholder = false
			} else {
				if runeChar == ':' {
					inTagPlaceholder = true
					newString.WriteString(string(runeChar))
					continue
				}
			}
			// End handling placeholder tags

			newString.WriteString(fmt.Sprintf(`<ansi fg="%d">%s</ansi>`, patternValues[patternPosition], string(runeChar)))
			if patternPosition < patternValueLength-1 && runeChar != 32 { // space
				patternPosition += 1 // advance the color token position
			}
		}
		newString.WriteString(`</ansi>`)
	} else if method[0] == Stretch {
		// Spread the whole pattern to fit the string
		subCounter := 0
		stretchAmount := int(math.Floor(float64(runewidth.StringWidth(input)) / float64(len(patternValues))))
		if stretchAmount < 1 {
			stretchAmount = 1
		}
		newString.WriteString(`<ansi>`)
		for _, runeChar := range input {

			// Handle placeholder tags that look like :123
			if inTagPlaceholder {
				if runeChar != 32 {
					newString.WriteString(string(runeChar))
					continue
				}
				inTagPlaceholder = false
			} else {
				if runeChar == ':' {
					inTagPlaceholder = true
					newString.WriteString(string(runeChar))
					continue
				}
			}
			// End handling placeholder tags

			newString.WriteString(fmt.Sprintf(`<ansi fg="%d">%s</ansi>`, patternValues[patternPosition], string(runeChar)))
			subCounter++
			if patternPosition < patternValueLength-1 && runeChar != 32 { // space
				if subCounter%stretchAmount == 0 {
					patternPosition += 1 // advance the color token position
				}
			}
		}
		newString.WriteString(`</ansi>`)
	}

	finalString := newString.String()

	for tmp, replacement := range preExistingTags {
		finalString = strings.Replace(finalString, tmp, replacement, -1)
	}

	return finalString
}

func CompileColorPatterns() {

	if colorsCompiled {
		return
	}

	for name, numbers := range numericPatterns {
		cPatterns := []string{}

		for _, num := range numbers {
			cPatterns = append(cPatterns, fmt.Sprintf(`{%d}`, num))
		}
		ShortTagPatterns[name] = cPatterns
	}

	colorsCompiled = true
}

func IsValidPattern(pName string) bool {
	if _, ok := numericPatterns[pName]; ok {
		return true
	}
	return false
}

func LoadColorPatterns() {

	start := time.Now()

	path := string(configs.GetFilePathsConfig().FolderDataFiles) + `/color-patterns.yaml`

	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(errors.Wrap(err, `filepath: `+path))
	}

	clear(numericPatterns)
	clear(ShortTagPatterns)
	colorsCompiled = false

	err = yaml.Unmarshal(bytes, &numericPatterns)
	if err != nil {
		panic(errors.Wrap(err, `filepath: `+path))
	}

	CompileColorPatterns()

	mudlog.Info("...LoadColorPatterns()", "loadedCount", len(numericPatterns), "Time Taken", time.Since(start))

	for _, name := range GetColorPatternNames() {
		mudlog.Info("Color Test (Patterns)", "name", name,
			"(default)", ansitags.Parse(ApplyColorPattern(`Color test pattern`, name)),
			"Stretch", ansitags.Parse(ApplyColorPattern(`Color test pattern`, name, Stretch)),
			"Words", ansitags.Parse(ApplyColorPattern(`Color test pattern color test pattern`, name, Words)),
		)
	}
}
