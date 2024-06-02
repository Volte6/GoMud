package progressbar

import (
	"fmt"
	"math"
	"strings"
)

type BarDisplay uint8

const (
	PromptReplace          BarDisplay = iota // Show instead of the normal prompt
	PromptPrefix                             // Show before the prompt
	PromptSuffix                             // Show after the prompt
	PromptNoBar                              // Don't show the bar at all
	PromptDisableInput                       // Disable input while the bar is showing
	PromptTypingInterrupts                   // Typing interrupts/cancels the bar
)

type ProgressBar struct {
	name           string
	turnStart      uint64
	turnTotal      int
	turnNow        uint64
	size           int
	renderStyle    BarDisplay
	cancelled      bool
	onCompleteFunc func()
}

func (p *ProgressBar) Update(turnNow uint64) {
	if p.turnStart == 0 {
		p.turnStart = turnNow
	}
	p.turnNow = turnNow
}

func (p *ProgressBar) String() string {
	output := strings.Builder{}

	turnsPassed := int(p.turnNow - uint64(p.turnStart))

	pctDone := float64(turnsPassed) / float64(p.turnTotal)
	if pctDone > 1 {
		pctDone = 1
	}

	if p.size > 0 {
		fullQty := int(math.Floor(float64(p.size) * pctDone))
		emptyQty := p.size - fullQty

		colors := []string{
			`22`, `28`, `34`, `40`, `46`, `83`, `120`, `157`, `194`, `231`,
		}

		chunkSize := float64(p.size) / float64(len(colors))

		output.WriteString(`<ansi fg="22">`)
		lastColor := colors[0]
		for i := 0; i < fullQty; i++ {
			nextColor := colors[int(math.Floor(float64(i)/chunkSize))]
			if nextColor != lastColor {
				output.WriteString(`</ansi><ansi fg="` + nextColor + `">`)
			}
			output.WriteString(`█`)
		}
		output.WriteString(`</ansi>`)

		output.WriteString(`<ansi fg="22">`)
		output.WriteString(strings.Repeat(`░`, emptyQty))
		output.WriteString(`</ansi>`)
	}

	output.WriteString(fmt.Sprintf(`<ansi fg="57">%d%% %s </ansi>`, int(pctDone*100), p.name))

	return output.String()
}

func (p *ProgressBar) Done() bool {
	return p.cancelled || p.turnNow >= p.turnStart+uint64(p.turnTotal)
}

func (p *ProgressBar) RenderStyle() BarDisplay {
	return p.renderStyle
}

func (p *ProgressBar) Cancel() {
	p.cancelled = true
	p.onCompleteFunc = nil
}

func (p *ProgressBar) OnComplete() {
	if p.onCompleteFunc != nil {
		p.onCompleteFunc()
	}
}

func New(name string, turns int, onComplete func(), renderFlags ...BarDisplay) *ProgressBar {

	renderFlag := PromptReplace
	size := 30

	for _, flag := range renderFlags {
		switch flag {
		case PromptReplace:
			renderFlag = PromptReplace
		case PromptPrefix:
			renderFlag = PromptPrefix
		case PromptSuffix:
			renderFlag = PromptSuffix
		case PromptNoBar:
			size = 0
		}
	}

	return &ProgressBar{
		name:           name,
		turnStart:      0,
		turnTotal:      turns,
		size:           size,
		renderStyle:    renderFlag,
		onCompleteFunc: onComplete,
	}

}
