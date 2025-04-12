package scripting

import (
	"errors"
	"time"

	"github.com/dop251/goja"
	"github.com/volte6/gomud/internal/colorpatterns"
)

var (
	ErrEventNotFound = errors.New(`event not found`)
)

type TextWrapperStyle struct {
	cache             string
	Fg                string // ansi class name for foreground
	Bg                string // ansi class name for background
	ColorPattern      string // optional color pattern
	colorPatternStyle colorpatterns.ColorizeStyle
}

func (t *TextWrapperStyle) Set(fg string, bg string, colorpattern string, colorStyle ...colorpatterns.ColorizeStyle) {
	t.Fg = fg
	t.Bg = bg
	t.ColorPattern = colorpattern
	if len(colorStyle) > 0 {
		t.colorPatternStyle = colorStyle[0]
	} else {
		t.colorPatternStyle = colorpatterns.Default
	}
}

func (t *TextWrapperStyle) Reset() {
	t.cache = ``
	t.Fg = ``
	t.Bg = ``
	t.ColorPattern = ``
}

func (t *TextWrapperStyle) Empty() bool {
	return t.Fg == `` && t.Bg == `` && t.ColorPattern == ``
}

func (t *TextWrapperStyle) AnsiClass() string {
	if t.cache != `` {
		return t.cache
	}

	if t.Fg != `` {
		t.cache += ` fg="` + t.Fg + `"`
	}

	if t.Bg != `` {
		t.cache += ` bg="` + t.Bg + `"`
	}

	return t.cache
}

func (t *TextWrapperStyle) Wrap(str string) string {
	if !t.Empty() {

		if t.ColorPattern != `` {
			str = colorpatterns.ApplyColorPattern(str, t.ColorPattern, t.colorPatternStyle)
		}

		if ac := t.AnsiClass(); ac != `` {
			str = `<ansi ` + t.AnsiClass() + `>` + str + `</ansi>`
		}
	}
	return str
}

var (
	errNoScript = errors.New("no script")
	errTimeout  = errors.New("script timeout")

	// If non empty, will wrap output to users or rooms in this style
	userTextWrap = TextWrapperStyle{}
	roomTextWrap = TextWrapperStyle{}
)

func Setup(scriptLoadTimeoutMs int, scriptRoomTimeoutMs int) {

	scriptLoadTimeout = time.Duration(scriptLoadTimeoutMs) * time.Millisecond

	t := time.Duration(scriptRoomTimeoutMs) * time.Millisecond
	scriptRoomTimeout = t
	scriptBuffTimeout = t
	scriptItemTimeout = t
	scriptMobTimeout = t
	scriptSpellTimeout = t
}

func setAllScriptingFunctions(vm *goja.Runtime) {
	setMessagingFunctions(vm)
	setRoomFunctions(vm)
	setActorFunctions(vm)
	setSpellFunctions(vm)
	setItemFunctions(vm)
	setUtilFunctions(vm)
}

func PruneVMs(forceClear ...bool) {

	if len(forceClear) > 0 && forceClear[0] {
		ClearRoomVMs()
		ClearMobVMs()
		ClearBuffVMs()
		ClearItemVMs()
		ClearSpellVMs()
	} else {
		PruneRoomVMs()
		PruneMobVMs()
		PruneBuffVMs()
		PruneItemVMs()
		PruneSpellVMs()
	}

}
