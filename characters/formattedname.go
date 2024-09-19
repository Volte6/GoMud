package characters

import (
	"fmt"

	"github.com/volte6/mud/colorpatterns"
)

var (
	// -short suffix should also be defined in case shorthand symbols are preferred
	adjectiveSwaps = map[string]string{

		// Are they charmed/friendly?
		`charmed`:       colorpatterns.ApplyColorPattern(`♥friend`, `pink`),
		`charmed-short`: colorpatterns.ApplyColorPattern(`♥`, `pink`),

		// Are they downed?
		`downed`:       colorpatterns.ApplyColorPattern(`☠downed`, `red`),
		`downed-short`: colorpatterns.ApplyColorPattern(`☠`, `red`),

		// Are they hiding?
		`hidden`:       colorpatterns.ApplyColorPattern(`hidden`, `gray`),
		`hidden-short`: colorpatterns.ApplyColorPattern(`?`, `gray`),

		// Does light come from this character?
		`lit`:       colorpatterns.ApplyColors(`⚙Lit`, []int{187, 229, 228, 227}),
		`lit-short`: colorpatterns.ApplyColors(`⚙`, []int{187, 229, 228, 227}),

		// Are they hiding?
		`sleeping`:       colorpatterns.ApplyColorPattern(`asleep`, `gray`),
		`sleeping-short`: colorpatterns.ApplyColorPattern(`zZz`, `gray`),

		// Have they disconnected and are zombie status?
		`zombie`:       colorpatterns.ApplyColorPattern(`zOmBie`, `zombie`),
		`zombie-short`: colorpatterns.ApplyColorPattern(`z`, `zombie`),

		// Have they disconnected and are zombie status?
		`poisoned`:       colorpatterns.ApplyColorPattern(`☠poisoned`, `purple`),
		`poisoned-short`: colorpatterns.ApplyColorPattern(`☠`, `purple`),
	}
)

type FormattedName struct {
	Name               string
	Type               string // mob/user
	Suffix             string // What ansi alias suffix to use (if any)
	Adjectives         []string
	UseShortAdjectives bool // Whether to failover to short adjectives
	QuestAlert         bool // Whether this mob is relevant to a current quest
}

func (f FormattedName) String() string {

	ansiAlias := f.Type
	if f.Suffix != `` {
		ansiAlias = fmt.Sprintf(`%s-%s`, ansiAlias, f.Suffix)
	}

	output := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, ansiAlias, f.Name)

	adjectives := f.Adjectives

	shortSuffix := ``
	if f.UseShortAdjectives || len(adjectives) > 3 {
		shortSuffix = `-short`
	}

	if adjLen := len(adjectives); adjLen > 0 {
		output += ` <ansi fg="black-bold">(`
		for i, adj := range adjectives {
			if newAdj, ok := adjectiveSwaps[adj+shortSuffix]; ok {
				output += newAdj
			} else {
				output += adj
			}
			if i < adjLen-1 {
				output += `|`
			}
		}
		output += `)</ansi>`
	}

	if f.QuestAlert {
		output = `<ansi fg="questflag">★</ansi>` + output
	}

	return output
}

func GetFormattedAdjective(adjName string) string {
	if newAdj, ok := adjectiveSwaps[adjName]; ok {
		return newAdj
	}
	return adjName
}
