package colorpatterns

import (
	"strings"

	"github.com/Volte6/ansitags"
	"github.com/volte6/mud/util"
)

// TODO: This introduces unnecessary steps by using an intermediate encoding before the ANSITAG encoding.
// This should be optimized.

var (
	// Color patterns that repeat as they are applied
	ColorPatterns = map[string][]string{
		`glowing`: {`{184}`, `{226}`, `{227}`, `{228}`, `{229}`, `{230}`, `{231}`, `{230}`, `{229}`, `{228}`, `{227}`, `{226}`, `{184}`, `{142}`, `{100}`, `{58}`},
		`coupon`:  {`{147}`, `{231}`},
		`rainbow`: {`{196}`, `{214}`, `{226}`, `{118}`, `{51}`, `{21}`, `{93}`},
	}
)

func GetDebugColorPatternOutput() map[string]string {

	output := map[string]string{}

	for name, _ := range ColorPatterns {
		output[name] = ansitags.Parse(util.ConvertColorShortTags(ApplyColorPattern(name+` color pattern test`, name)))
	}

	return output

}

func ApplyColorPattern(input string, pattern string) string {

	patternValues, ok := ColorPatterns[pattern]
	if !ok {
		return input
	}
	patternValueLength := len(patternValues)
	patternText := []byte(input)

	newString := strings.Builder{}

	patternPosition := 0
	for i := 0; i < len(patternText); i++ {
		if patternText[i] != 32 { // space
			newString.WriteString(patternValues[patternPosition%patternValueLength])
			patternPosition++
		}
		newString.Write(patternText[i : i+1])
	}

	return util.ConvertColorShortTags(newString.String())

}
