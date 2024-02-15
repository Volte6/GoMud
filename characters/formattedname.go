package characters

import (
	"fmt"
	"strings"
)

type FormattedName struct {
	Name          string
	Type          string   // mob/user
	LightSource   bool     // Are they a source of light?
	Suffix        string   // What ansi alias suffix to use (if any)
	Flags         []string // Single charavter flags
	HealthDisplay string   // What health to append to the end of the name (if any)
}

func (c FormattedName) String() string {

	ansiAlias := c.Type
	if c.Suffix != `` {
		ansiAlias = fmt.Sprintf(`%s-%s`, ansiAlias, c.Suffix)
	}

	output := ``
	if c.LightSource {
		output = fmt.Sprintf(`<ansi fg="%s">%s</ansi><ansi fg="187">⚙</ansi>`, ansiAlias, c.Name)
	} else {
		output = fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, ansiAlias, c.Name)
	}

	if len(c.Flags) > 0 {
		output += fmt.Sprintf(` <ansi fg="name-flags-wrapper">(<ansi fg="name-flag">%s</ansi>)</ansi>`, strings.Join(c.Flags, `</ansi>, <ansi fg="name-flag">`))
	}

	if c.HealthDisplay != `` {
		if c.HealthDisplay == `downed` {
			output = fmt.Sprintf(`%s <ansi fg="red">(downed)</ansi>`, output)
		} else {
			output = fmt.Sprintf(`%s <ansi fg="black-bold">(%s)</ansi>`, output, c.HealthDisplay)
		}
	}
	return output
}
