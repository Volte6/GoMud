package characters

import (
	"fmt"
)

var (
	// -short suffix should also be defined in case shorthand symbols are preferred
	adjectiveSwaps = map[string]string{
		// Have they disconnected and are zombie status?
		`zombie`:       `<ansi fg="77">z</ansi><ansi fg="77">O</ansi><ansi fg="113">m</ansi><ansi fg="72">B</ansi><ansi fg="65">i</ansi><ansi fg="78">e</ansi>`,
		`zombie-short`: `<ansi fg="77">z</ansi>`,
		// Are they charmed/friendly?
		//`charmed`: `<ansi fg="225">♥</ansi><ansi fg="219">c</ansi><ansi fg="213">h</ansi><ansi fg="207">a</ansi><ansi fg="201">r</ansi><ansi fg="164">m</ansi><ansi fg="127">e</ansi><ansi fg="90">d</ansi>`,
		`charmed`:       `<ansi fg="225">♥</ansi><ansi fg="219">f</ansi><ansi fg="213">r</ansi><ansi fg="207">i</ansi><ansi fg="201">e</ansi><ansi fg="164">n</ansi><ansi fg="127">d</ansi>`,
		`charmed-short`: `<ansi fg="127">♥</ansi>`,
		// Are they downed?
		`downed`:       `<ansi fg="7">☠</ansi><ansi fg="red">downed</ansi>`,
		`downed-short`: `<ansi fg="red">☠</ansi>`,
		// Does light come from this character?
		`lit`:       `<ansi fg="187">⚙</ansi><ansi fg="229">L</ansi><ansi fg="228">i</ansi><ansi fg="227">t</ansi>`,
		`lit-short`: `<ansi fg="187">⚙</ansi>`,
	}
)

type FormattedName struct {
	Name               string
	Type               string // mob/user
	Suffix             string // What ansi alias suffix to use (if any)
	Adjectives         []string
	UseShortAdjectives bool // Whether to failover to short adjectives
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

	return output
}
