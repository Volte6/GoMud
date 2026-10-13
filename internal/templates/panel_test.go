package templates

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/ansitags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makePanel is a test helper that builds a Panel directly without loading YAML.
// width is the border-inclusive total panel width (0 = size to content).
func makePanel(id, title string, width int, b borderStyle, rows []PanelRow) *Panel {
	return &Panel{
		id:        id,
		title:     title,
		width:     width,
		border:    b,
		chars:     charsetSingle,
		columns:   1,
		columnGap: defaultColumnGap,
		rows:      rows,
	}
}

// makeMultiColPanel is a test helper for panels with columns > 1.
// width is the border-inclusive total panel width (0 = size to content).
func makeMultiColPanel(id, title string, width, columns, columnGap int, b borderStyle, rows []PanelRow) *Panel {
	return &Panel{
		id:        id,
		title:     title,
		width:     width,
		border:    b,
		chars:     charsetSingle,
		columns:   columns,
		columnGap: columnGap,
		rows:      rows,
	}
}

// splitLines splits a rendered string into lines for easier assertion.
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

// makeLayout is a test helper that builds a PanelLayout from a slice of slots,
// where each slot is a slice of panel rows ([][]*Panel).
func makeLayout(gap int, slotRows ...[][]*Panel) *PanelLayout {
	l := &PanelLayout{
		border: borderFull,
		gap:    gap,
		byID:   make(map[string]*Panel),
	}
	for _, rows := range slotRows {
		slot := &layoutSlot{rows: rows}
		for _, row := range rows {
			for _, p := range row {
				l.byID[p.id] = p
			}
		}
		l.slots = append(l.slots, slot)
	}
	return l
}

// ---------------------------------------------------------------------------
// panelVisualWidth
// ---------------------------------------------------------------------------

func TestPanelVisualWidth_PlainText(t *testing.T) {
	assert.Equal(t, 5, panelVisualWidth("hello"))
}

func TestPanelVisualWidth_AnsiTagsIgnored(t *testing.T) {
	assert.Equal(t, 5, panelVisualWidth(`<ansi fg="yellow">hello</ansi>`))
}

func TestPanelVisualWidth_EmptyString(t *testing.T) {
	assert.Equal(t, 0, panelVisualWidth(""))
}

// ---------------------------------------------------------------------------
// panelInnerWidth
// ---------------------------------------------------------------------------

func TestPanelInnerWidth_UsesMinWidth(t *testing.T) {
	// width=24 → innerWidth()=24-4=20; content "A: v"=4 < 20, so result=20.
	p := makePanel("x", "X", 24, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "v"},
	})
	assert.Equal(t, 20, panelInnerWidth(p))
}

func TestPanelInnerWidth_ExpandsBeyondMinWidth(t *testing.T) {
	// width=9 → innerWidth()=5; content "Label:"(6)+1+"a long value here"(17)=24 > 5, so expands to 24.
	p := makePanel("x", "X", 9, borderFull, []PanelRow{
		{FullLabel: "Label:", ShortLabel: "L:", Value: "a long value here"},
	})
	// "Label:" (6) + 1 + "a long value here" (17) = 24
	assert.Equal(t, 24, panelInnerWidth(p))
}

func TestPanelInnerWidth_MultiColumn_EvenRows(t *testing.T) {
	// width=34 → innerWidth()=30; colTarget=(30-2)/2=14
	// widestCell = "Str:"(4)+1+"42(+3)"(6) = 11 < 14, so colWidth=14, total=14*2+2=30
	p := makeMultiColPanel("x", "X", 34, 2, 2, borderFull, []PanelRow{
		{FullLabel: "Str:", ShortLabel: "S:", Value: "42(+3)"},
		{FullLabel: "Vit:", ShortLabel: "V:", Value: "38(+0)"},
	})
	assert.Equal(t, 30, panelInnerWidth(p))
}

func TestPanelInnerWidth_MultiColumn_ExpandsForWideCell(t *testing.T) {
	// width=14 → innerWidth()=10; colTarget=(10-2)/2=4
	// "Mysticism:"(10)+1+"42(+3)"(6) = 17 > 4, so colWidth=17, total=17*2+2=36
	p := makeMultiColPanel("x", "X", 14, 2, 2, borderFull, []PanelRow{
		{FullLabel: "Mysticism:", ShortLabel: "Mys:", Value: "42(+3)"},
		{FullLabel: "Mysticism:", ShortLabel: "Mys:", Value: "38(+0)"},
	})
	assert.Equal(t, 36, panelInnerWidth(p))
}

// ---------------------------------------------------------------------------
// chooseLabel
// ---------------------------------------------------------------------------

func TestChooseLabel_PrefersFullLabel(t *testing.T) {
	row := PanelRow{FullLabel: "Full:", ShortLabel: "F:", Value: "val"}
	assert.Equal(t, "Full:", chooseLabel(row, 20))
}

func TestChooseLabel_FallsBackToShortLabel(t *testing.T) {
	row := PanelRow{FullLabel: "Very Long Label:", ShortLabel: "VLL:", Value: "val"}
	assert.Equal(t, "VLL:", chooseLabel(row, 10))
}

// ---------------------------------------------------------------------------
// renderPanel – single column
// ---------------------------------------------------------------------------

func TestRenderPanel_FullBorder_Structure(t *testing.T) {
	p := makePanel("info", "Info", 14, borderFull, []PanelRow{
		{FullLabel: "Name:", ShortLabel: "N:", Value: "Alice"},
		{FullLabel: "Age:", ShortLabel: "A:", Value: "30"},
	})
	got := renderPanel(p)

	require.Equal(t, 4, len(got), "top + 2 content + bottom")
	assert.True(t, strings.HasPrefix(got[0], "┌"))
	assert.True(t, strings.HasSuffix(got[0], "┐"))
	assert.True(t, strings.HasPrefix(got[1], "│"))
	assert.True(t, strings.HasSuffix(got[1], "│"))
	assert.True(t, strings.HasPrefix(got[2], "│"))
	assert.True(t, strings.HasSuffix(got[2], "│"))
	assert.True(t, strings.HasPrefix(got[3], "└"))
	assert.True(t, strings.HasSuffix(got[3], "┘"))
}

func TestRenderPanel_OpenBorder_InteriorRowsOpen(t *testing.T) {
	p := makePanel("info", "Info", 14, borderOpen, []PanelRow{
		{FullLabel: "First:", ShortLabel: "F:", Value: "a"},
		{FullLabel: "Middle:", ShortLabel: "M:", Value: "b"},
		{FullLabel: "Last:", ShortLabel: "L:", Value: "c"},
	})
	got := renderPanel(p)

	require.Equal(t, 5, len(got))
	assert.True(t, strings.HasPrefix(got[1], "│"), "first row has left border")
	assert.True(t, strings.HasSuffix(got[1], "│"), "first row has right border")
	assert.True(t, strings.HasPrefix(got[2], " "), "middle row is open")
	assert.True(t, strings.HasSuffix(got[2], " "), "middle row is open")
	assert.True(t, strings.HasPrefix(got[3], "│"), "last row has left border")
	assert.True(t, strings.HasSuffix(got[3], "│"), "last row has right border")
}

func TestRenderPanel_BlankRow(t *testing.T) {
	p := makePanel("x", "X", 14, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "1"},
		{Blank: true},
		{FullLabel: "B:", ShortLabel: "B:", Value: "2"},
	})
	got := renderPanel(p)
	require.Equal(t, 5, len(got))
	inner := panelInnerWidth(p)
	assert.Equal(t, charsetSingle.VerticalLeft+strings.Repeat(" ", inner+2*panelPad)+charsetSingle.VerticalRight, got[2])
}

func TestRenderPanel_BlockRow(t *testing.T) {
	p := makePanel("x", "X", 40, borderFull, []PanelRow{
		{
			ShortLabel: "A:",
			FullLabel:  "Alpha:",
			Value:      "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since 1966, when designers at Letraset and James Mosley, the librarian at St Bride Printing Library in London, took a 1914 Cicero translation and scrambled it to make dummy text for Letraset's Body Type sheets.",
			WrapWidth:  80,
			Block:      true,
		},
	})
	got := renderPanel(p)
	require.Equal(t, 13, len(got))
	for i, line := range got {
		assert.LessOrEqual(t, panelVisualWidth(line), p.width, "line %d longer than panel width", i)
	}
}

func TestRenderPanel_AllLinesEqualVisualWidth(t *testing.T) {
	p := makePanel("x", "Title", 19, borderFull, []PanelRow{
		{FullLabel: "Short:", ShortLabel: "S:", Value: "v"},
		{FullLabel: "A much longer label:", ShortLabel: "AML:", Value: "value"},
	})
	got := renderPanel(p)
	w0 := panelVisualWidth(got[0])
	for i, line := range got {
		assert.Equal(t, w0, panelVisualWidth(line), "line %d visual width differs", i)
	}
}

func TestRenderPanel_AnsiTagsDoNotAffectWidth(t *testing.T) {
	p := makePanel("x", "X", 19, borderFull, []PanelRow{
		{FullLabel: `<ansi fg="yellow">Name:</ansi>`, ShortLabel: "N:", Value: `<ansi fg="green">Alice</ansi>`},
	})
	got := renderPanel(p)
	w0 := panelVisualWidth(got[0])
	for i, line := range got {
		assert.Equal(t, w0, panelVisualWidth(line), "line %d visual width differs", i)
	}
}

func TestRenderPanel_SetLabelWidth_AlignsValues(t *testing.T) {
	// Labels of different lengths — without SetLabelWidth values start at different columns.
	// With SetLabelWidth(12), every value starts at column 12+1=13.
	p := makePanel("x", "X", 34, borderFull, []PanelRow{
		{FullLabel: "Short:", ShortLabel: "S:", Value: "val1"},
		{FullLabel: "Much Longer:", ShortLabel: "ML:", Value: "val2"},
	})
	p.labelWidth = 12
	got := renderPanel(p)

	// Strip ANSI and borders to get raw content lines.
	line1 := ansitags.Parse(got[1], ansitags.StripTags)
	line2 := ansitags.Parse(got[2], ansitags.StripTags)

	// Find the column where the value starts (after '│ ' + label + ' ').
	// Both lines should have their value at the same column.
	valCol1 := strings.Index(line1, "val1")
	valCol2 := strings.Index(line2, "val2")
	assert.Equal(t, valCol1, valCol2, "values should start at the same column")
	assert.True(t, valCol1 > 0, "value column should be positive")
}

func TestRenderPanel_SetLabelWidth_ExpandsInnerWidth(t *testing.T) {
	// labelWidth wider than any label should expand the panel's inner width accordingly.
	p := makePanel("x", "X", 0, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "v"},
	})
	p.labelWidth = 20
	// inner = max(innerWidth()=0, labelWidth=20 + 1 + visualWidth("v")=1) = 22
	assert.Equal(t, 22, panelInnerWidth(p))

	// All lines must have equal visual width.
	got := renderPanel(p)
	w0 := panelVisualWidth(got[0])
	for i, line := range got {
		assert.Equal(t, w0, panelVisualWidth(line), "line %d visual width differs", i)
	}
}

func TestRenderPanel_ContentContainsLabelAndValue(t *testing.T) {
	p := makePanel("x", "X", 19, borderFull, []PanelRow{
		{FullLabel: "Name:", ShortLabel: "N:", Value: "Bob"},
	})
	got := renderPanel(p)
	assert.Contains(t, got[1], "Name:")
	assert.Contains(t, got[1], "Bob")
}

// ---------------------------------------------------------------------------
// renderPanel – multi-column
// ---------------------------------------------------------------------------

func TestRenderPanel_MultiColumn_EvenRows_LineCount(t *testing.T) {
	// 4 rows paired into 2 content lines -> top + 2 + bottom = 4
	p := makeMultiColPanel("x", "X", 34, 2, 2, borderFull, []PanelRow{
		{FullLabel: "Str:", ShortLabel: "S:", Value: "10"},
		{FullLabel: "Vit:", ShortLabel: "V:", Value: "11"},
		{FullLabel: "Spd:", ShortLabel: "Sp:", Value: "12"},
		{FullLabel: "Mys:", ShortLabel: "M:", Value: "13"},
	})
	got := renderPanel(p)
	assert.Equal(t, 4, len(got), "top + 2 paired content lines + bottom")
}

func TestRenderPanel_MultiColumn_OddRows_LineCount(t *testing.T) {
	// 3 rows: pair(0,1) + lone(2) -> top + 2 + bottom = 4
	p := makeMultiColPanel("x", "X", 34, 2, 2, borderFull, []PanelRow{
		{FullLabel: "Str:", ShortLabel: "S:", Value: "10"},
		{FullLabel: "Vit:", ShortLabel: "V:", Value: "11"},
		{FullLabel: "Spd:", ShortLabel: "Sp:", Value: "12"},
	})
	got := renderPanel(p)
	assert.Equal(t, 4, len(got), "top + pair + lone + bottom")
}

func TestRenderPanel_MultiColumn_BothCellsPresent(t *testing.T) {
	p := makeMultiColPanel("x", "X", 34, 2, 2, borderFull, []PanelRow{
		{FullLabel: "Str:", ShortLabel: "S:", Value: "10"},
		{FullLabel: "Vit:", ShortLabel: "V:", Value: "20"},
	})
	got := renderPanel(p)
	assert.Contains(t, got[1], "Str:")
	assert.Contains(t, got[1], "10")
	assert.Contains(t, got[1], "Vit:")
	assert.Contains(t, got[1], "20")
}

func TestRenderPanel_MultiColumn_AllLinesEqualVisualWidth(t *testing.T) {
	p := makeMultiColPanel("x", "Attrs", 34, 2, 2, borderFull, []PanelRow{
		{FullLabel: "Strength:", ShortLabel: "Str:", Value: "42(+3)"},
		{FullLabel: "Vitality:", ShortLabel: "Vit:", Value: "38(+0)"},
		{FullLabel: "Speed:", ShortLabel: "Spd:", Value: "55(+5)"},
		{FullLabel: "Mysticism:", ShortLabel: "Mys:", Value: "20(+0)"},
		{FullLabel: "Smarts:", ShortLabel: "Smt:", Value: "30(+1)"},
	})
	got := renderPanel(p)
	w0 := panelVisualWidth(got[0])
	for i, line := range got {
		assert.Equal(t, w0, panelVisualWidth(line), "line %d visual width differs", i)
	}
}

func TestRenderPanel_MultiColumn_AnsiTagsDoNotAffectWidth(t *testing.T) {
	p := makeMultiColPanel("x", "X", 34, 2, 2, borderFull, []PanelRow{
		{FullLabel: `<ansi fg="yellow">Str:</ansi>`, ShortLabel: "S:", Value: `<ansi fg="stat">42</ansi><ansi fg="statmod">(+3)</ansi>`},
		{FullLabel: `<ansi fg="yellow">Vit:</ansi>`, ShortLabel: "V:", Value: `<ansi fg="stat">38</ansi><ansi fg="statmod">(+0)</ansi>`},
		{FullLabel: `<ansi fg="yellow">Spd:</ansi>`, ShortLabel: "Sp:", Value: `<ansi fg="stat">55</ansi><ansi fg="statmod">(+5)</ansi>`},
		{FullLabel: `<ansi fg="yellow">Mys:</ansi>`, ShortLabel: "M:", Value: `<ansi fg="stat">20</ansi><ansi fg="statmod">(+0)</ansi>`},
	})
	got := renderPanel(p)
	w0 := panelVisualWidth(got[0])
	for i, line := range got {
		assert.Equal(t, w0, panelVisualWidth(line), "line %d visual width differs", i)
	}
}

func TestRenderPanel_MultiColumn_ShortLabelFallback(t *testing.T) {
	// renderCellContent with a colWidth smaller than the full label cell triggers fallback.
	row := PanelRow{FullLabel: "Strength:", ShortLabel: "Str:", Value: "10"}
	result := renderCellContent(row, 6)
	assert.True(t, strings.HasPrefix(result, "Str:"),
		"expected short label, got: %q", result)
	assert.NotContains(t, result, "Strength:")
}

// ---------------------------------------------------------------------------
// composePanelGroup
// ---------------------------------------------------------------------------

func TestComposePanelGroup_TwoPanelsEqualHeight(t *testing.T) {
	p1 := makePanel("a", "A", 14, borderFull, []PanelRow{
		{FullLabel: "X:", ShortLabel: "X:", Value: "1"},
	})
	p2 := makePanel("b", "B", 14, borderFull, []PanelRow{
		{FullLabel: "Y:", ShortLabel: "Y:", Value: "2"},
	})
	got := composePanelGroup([]*Panel{p1, p2}, 1)
	assert.Equal(t, 3, len(got))
	for _, line := range got {
		assert.True(t, len(line) > 0)
	}
}

func TestComposePanelGroup_TwoPanelsUnequalHeight(t *testing.T) {
	p1 := makePanel("a", "A", 14, borderFull, []PanelRow{
		{FullLabel: "X:", ShortLabel: "X:", Value: "1"},
		{FullLabel: "Y:", ShortLabel: "Y:", Value: "2"},
		{FullLabel: "Z:", ShortLabel: "Z:", Value: "3"},
	})
	p2 := makePanel("b", "B", 14, borderFull, []PanelRow{
		{FullLabel: "W:", ShortLabel: "W:", Value: "9"},
	})
	got := composePanelGroup([]*Panel{p1, p2}, 1)
	assert.Equal(t, 5, len(got))
	w0 := panelVisualWidth(got[0])
	for i, line := range got {
		assert.Equal(t, w0, panelVisualWidth(line), "composed line %d visual width differs", i)
	}
	assert.True(t, strings.HasSuffix(got[len(got)-1], "┘"))
}

func TestComposePanelGroup_GapIsRespected(t *testing.T) {
	p1 := makePanel("a", "A", 9, borderFull, []PanelRow{
		{FullLabel: "X:", ShortLabel: "X:", Value: "1"},
	})
	p2 := makePanel("b", "B", 9, borderFull, []PanelRow{
		{FullLabel: "Y:", ShortLabel: "Y:", Value: "2"},
	})
	gap0 := composePanelGroup([]*Panel{p1, p2}, 0)
	gap3 := composePanelGroup([]*Panel{p1, p2}, 3)
	for i := range gap0 {
		assert.Equal(t, panelVisualWidth(gap0[i])+3, panelVisualWidth(gap3[i]),
			"line %d: gap3 should be 3 wider than gap0", i)
	}
}

// ---------------------------------------------------------------------------
// PanelLayout – lookup and render
// ---------------------------------------------------------------------------

func TestPanelLayout_PanelLookup(t *testing.T) {
	p := makePanel("test", "Test", 14, borderFull, nil)
	layout := makeLayout(1, [][]*Panel{{p}})
	assert.Equal(t, p, layout.Panel("test"))
}

func TestPanelLayout_PanelLookup_PanicsOnMissing(t *testing.T) {
	layout := &PanelLayout{byID: make(map[string]*Panel)}
	assert.Panics(t, func() { layout.Panel("nonexistent") })
}

func TestPanelLayout_Render_SingleSlot_TwoRows(t *testing.T) {
	// One slot, two stacked rows -> lines from row1 then row2, no gap between slots.
	p1 := makePanel("a", "A", 12, borderFull, []PanelRow{{FullLabel: "X:", ShortLabel: "X:", Value: "1"}})
	p2 := makePanel("b", "B", 12, borderFull, []PanelRow{{FullLabel: "Y:", ShortLabel: "Y:", Value: "2"}})
	layout := makeLayout(1, [][]*Panel{{p1}, {p2}})
	output := layout.Render()
	rendered := splitLines(output)
	// Each panel: top+1content+bottom = 3 lines. Two panels stacked = 6 lines.
	assert.Equal(t, 6, len(rendered))
	assert.True(t, strings.HasPrefix(rendered[0], "┌"), "first line is top border of p1")
	assert.True(t, strings.HasPrefix(rendered[3], "┌"), "fourth line is top border of p2")
}

func TestPanelLayout_Render_TwoSlots_SameHeight(t *testing.T) {
	// Two slots, each with one row of one panel, same height.
	p1 := makePanel("a", "A", 12, borderFull, []PanelRow{{FullLabel: "X:", ShortLabel: "X:", Value: "1"}})
	p2 := makePanel("b", "B", 12, borderFull, []PanelRow{{FullLabel: "Y:", ShortLabel: "Y:", Value: "2"}})
	layout := makeLayout(1, [][]*Panel{{p1}}, [][]*Panel{{p2}})
	output := layout.Render()
	rendered := splitLines(output)
	// Both slots: 3 lines each, same height -> 3 lines total (side by side).
	assert.Equal(t, 3, len(rendered))
	// Each line contains content from both panels joined by the gap.
	assert.Contains(t, rendered[1], "X:")
	assert.Contains(t, rendered[1], "Y:")
}

func TestPanelLayout_Render_TwoSlots_DifferentHeight(t *testing.T) {
	// Left slot: 1 row with 3-content-row panel (5 lines).
	// Right slot: 2 stacked rows, each with 1-content-row panel (3+3=6 lines).
	// Final height = max(5, 6) = 6.
	pLeft := makePanel("left", "Left", 14, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "1"},
		{FullLabel: "B:", ShortLabel: "B:", Value: "2"},
		{FullLabel: "C:", ShortLabel: "C:", Value: "3"},
	})
	pTop := makePanel("top", "Top", 14, borderFull, []PanelRow{
		{FullLabel: "X:", ShortLabel: "X:", Value: "9"},
	})
	pBot := makePanel("bot", "Bot", 14, borderFull, []PanelRow{
		{FullLabel: "Y:", ShortLabel: "Y:", Value: "8"},
	})
	layout := makeLayout(1,
		[][]*Panel{{pLeft}},
		[][]*Panel{{pTop}, {pBot}},
	)
	output := layout.Render()
	rendered := splitLines(output)
	assert.Equal(t, 6, len(rendered), "height = max(left=5, right=6)")

	// All lines must have the same visual width.
	w0 := panelVisualWidth(rendered[0])
	for i, line := range rendered {
		assert.Equal(t, w0, panelVisualWidth(line), "line %d visual width differs", i)
	}
}

func TestCharsetForName(t *testing.T) {
	assert.Equal(t, charsetSingle, charsetForName(""))
	assert.Equal(t, charsetSingle, charsetForName("single"))
	assert.Equal(t, charsetDouble, charsetForName("double"))
	assert.Equal(t, charsetDouble, charsetForName("DOUBLE"))
	assert.Equal(t, charsetRounded, charsetForName("rounded"))
}

func TestCharsetForName_LiteralString(t *testing.T) {
	// 8-rune literal: TL HTop TR VL VR BL HBottom BR
	got := charsetForName("╔═╗║│╚─┘")
	assert.Equal(t, "╔", got.TopLeft)
	assert.Equal(t, "═", got.Horizontal)
	assert.Equal(t, "╗", got.TopRight)
	assert.Equal(t, "║", got.VerticalLeft)
	assert.Equal(t, "│", got.VerticalRight)
	assert.Equal(t, "╚", got.BottomLeft)
	assert.Equal(t, "─", got.HorizontalBottom)
	assert.Equal(t, "┘", got.BottomRight)
}

func TestCharsetForName_LiteralString_WrongLength_FallsBackToSingle(t *testing.T) {
	// 7 runes – not 8 – should fall back to single
	assert.Equal(t, charsetSingle, charsetForName("╔═╗║│╚┘"))
	// 9 runes – not 8 – should fall back to single
	assert.Equal(t, charsetSingle, charsetForName("╔═╗║║╚╝═─"))
}

func TestRenderPanel_Charset_Double(t *testing.T) {
	p := makePanel("x", "X", 12, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "1"},
	})
	p.chars = charsetDouble
	got := renderPanel(p)
	assert.True(t, strings.HasPrefix(got[0], "\u2554"), "double top-left corner")
	assert.True(t, strings.HasSuffix(got[0], "\u2557"), "double top-right corner")
	assert.True(t, strings.HasPrefix(got[1], "\u2551"), "double vertical-left")
	assert.True(t, strings.HasSuffix(got[1], "\u2551"), "double vertical-right")
	assert.True(t, strings.HasPrefix(got[len(got)-1], "\u255a"), "double bottom-left corner")
	assert.True(t, strings.HasSuffix(got[len(got)-1], "\u255d"), "double bottom-right corner")
}

func TestRenderPanel_Charset_Rounded(t *testing.T) {
	p := makePanel("x", "X", 12, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "1"},
	})
	p.chars = charsetRounded
	got := renderPanel(p)
	assert.True(t, strings.HasPrefix(got[0], "\u256d"), "rounded top-left corner")
	assert.True(t, strings.HasSuffix(got[0], "\u256e"), "rounded top-right corner")
	assert.True(t, strings.HasPrefix(got[len(got)-1], "\u2570"), "rounded bottom-left corner")
	assert.True(t, strings.HasSuffix(got[len(got)-1], "\u256f"), "rounded bottom-right corner")
}

func TestRenderPanel_Charset_Literal_DistinctLeftRight(t *testing.T) {
	// Use a charset where VerticalLeft (║) != VerticalRight (│) and
	// HorizontalTop (═) != HorizontalBottom (─) to verify each side uses the
	// correct character.
	p := makePanel("x", "X", 12, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "1"},
	})
	p.chars = charsetForName("╔═╗║│╚─┘")
	got := renderPanel(p)
	assert.True(t, strings.HasPrefix(got[1], "\u2551"), "content row left border is VerticalLeft")
	assert.True(t, strings.HasSuffix(got[1], "\u2502"), "content row right border is VerticalRight")
	assert.True(t, strings.HasPrefix(got[0], "\u2554"), "top border uses TopLeft")
	assert.True(t, strings.Contains(got[0], "\u2550"), "top border uses HorizontalTop")
	assert.True(t, strings.HasPrefix(got[len(got)-1], "\u255a"), "bottom border uses BottomLeft")
	assert.True(t, strings.Contains(got[len(got)-1], "\u2500"), "bottom border uses HorizontalBottom")
	assert.True(t, strings.HasSuffix(got[len(got)-1], "\u2518"), "bottom border uses BottomRight")
}

func TestRenderPanel_PerPanelCharset_OverridesLayout(t *testing.T) {
	pSingle := makePanel("a", "A", 12, borderFull, []PanelRow{
		{FullLabel: "X:", ShortLabel: "X:", Value: "1"},
	})
	pDouble := makePanel("b", "B", 12, borderFull, []PanelRow{
		{FullLabel: "Y:", ShortLabel: "Y:", Value: "2"},
	})
	pDouble.chars = charsetDouble

	gotSingle := renderPanel(pSingle)
	gotDouble := renderPanel(pDouble)

	// Single panel uses ┌/┘
	assert.True(t, strings.HasPrefix(gotSingle[0], "\u250c"), "single top-left")
	assert.True(t, strings.HasSuffix(gotSingle[len(gotSingle)-1], "\u2518"), "single bottom-right")

	// Double panel uses ╔/╝
	assert.True(t, strings.HasPrefix(gotDouble[0], "\u2554"), "double top-left")
	assert.True(t, strings.HasSuffix(gotDouble[len(gotDouble)-1], "\u255d"), "double bottom-right")

	// All lines of each panel still have equal visual width
	w0 := panelVisualWidth(gotSingle[0])
	for i, line := range gotSingle {
		assert.Equal(t, w0, panelVisualWidth(line), "single panel line %d width differs", i)
	}
	w1 := panelVisualWidth(gotDouble[0])
	for i, line := range gotDouble {
		assert.Equal(t, w1, panelVisualWidth(line), "double panel line %d width differs", i)
	}
}

func TestRenderPanel_MaxWidth_WrapsValue(t *testing.T) {
	// Value "hello world" (11 chars) with WrapWidth=5:
	// SplitStringOnSpaces produces ["hello", "world"] = 2 non-empty chunks.
	p := makePanel("x", "X", 24, borderFull, []PanelRow{
		{FullLabel: "Desc:", ShortLabel: "D:", Value: "hello world", WrapWidth: 5},
	})
	got := renderPanel(p)
	// top border + 2 content lines + bottom border = 4
	require.Equal(t, 4, len(got), "should have a continuation line")
	assert.Contains(t, got[1], "Desc:")
	assert.Contains(t, got[1], "hello")
	assert.Contains(t, got[2], "world")
	// Continuation line must NOT contain the label.
	assert.NotContains(t, got[2], "Desc:")
}

func TestRenderPanel_MaxWidth_ContinuationAlignsWithValue(t *testing.T) {
	// Continuation lines must be indented by panelPad + labelWidth + 1 spaces,
	// which is the same offset as the value column on the first line.
	// SplitStringOnSpaces("aaa bbb", 3) produces ["aaa", "bbb"] (2 non-empty chunks).
	p := makePanel("x", "X", 34, borderFull, []PanelRow{
		{FullLabel: "Label:", ShortLabel: "L:", Value: "aaa bbb", WrapWidth: 3},
	})
	got := renderPanel(p)
	require.Equal(t, 4, len(got), "should have one continuation line")

	// Expected indent: panelPad + visualWidth("Label:") + 1
	expectedIndent := panelPad + panelVisualWidth("Label:") + 1

	// Strip ANSI and border from the continuation line, then count leading spaces.
	line2 := ansitags.Parse(got[2], ansitags.StripTags)
	line2Runes := []rune(line2)
	line2Body := string(line2Runes[1:]) // remove leading border rune (│)
	actualIndent := len([]rune(line2Body)) - len([]rune(strings.TrimLeft(line2Body, " ")))
	assert.Equal(t, expectedIndent, actualIndent, "continuation line should be indented to value column")
}

func TestRenderPanel_MaxWidth_AllLinesEqualVisualWidth(t *testing.T) {
	p := makePanel("x", "X", 34, borderFull, []PanelRow{
		{FullLabel: "Note:", ShortLabel: "N:", Value: "one two three four five", WrapWidth: 9},
	})
	got := renderPanel(p)
	w0 := panelVisualWidth(got[0])
	for i, line := range got {
		assert.Equal(t, w0, panelVisualWidth(line), "line %d visual width differs", i)
	}
}

func TestRenderPanel_MaxWidth_Zero_NoWrap(t *testing.T) {
	// WrapWidth=0 means no explicit wrapping; behaves like a plain Add row.
	p := makePanel("x", "X", 24, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "hello world", WrapWidth: 0},
	})
	got := renderPanel(p)
	// top + 1 content + bottom = 3
	require.Equal(t, 3, len(got))
	assert.Contains(t, got[1], "hello world")
}

func TestRenderPanel_MaxWidth_AnsiTagsHandled(t *testing.T) {
	// ANSI tags must not be mangled across split boundaries.
	value := `<ansi fg="yellow">hello</ansi> <ansi fg="green">world</ansi>`
	p := makePanel("x", "X", 34, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: value, WrapWidth: 5},
	})
	got := renderPanel(p)
	// Should produce at least 2 content lines.
	require.GreaterOrEqual(t, len(got), 4)
	// All lines must have equal visual width.
	w0 := panelVisualWidth(got[0])
	for i, line := range got {
		assert.Equal(t, w0, panelVisualWidth(line), "line %d visual width differs", i)
	}
}

func TestRenderPanel_TitleUsedVerbatim(t *testing.T) {
	p := makePanel("x", `<ansi fg="20">MyTitle</ansi>`, 19, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "1"},
	})
	got := renderPanel(p)
	assert.Contains(t, got[0], `<ansi fg="20">MyTitle</ansi>`, "title appears verbatim in top border")
	assert.NotContains(t, got[0], ".:")
}

func TestPanelLayout_Render_Margin(t *testing.T) {
	p := makePanel("a", "A", 12, borderFull, []PanelRow{
		{FullLabel: "X:", ShortLabel: "X:", Value: "1"},
	})
	layout := makeLayout(1, [][]*Panel{{p}})
	layout.margin = 3
	output := layout.Render()
	for _, line := range splitLines(output) {
		assert.True(t, strings.HasPrefix(line, "   "), "every line starts with 3-space margin, got: %q", line)
	}
}

func TestPanelLayout_Render_NestedLayout_StatusShape(t *testing.T) {
	// Mirrors the status page shape:
	// Slot 0: [info] (tall)
	// Slot 1: [attributes] stacked above [wealth, training] (same total height)
	info := makePanel("info", "Info", 14, borderFull, []PanelRow{
		{FullLabel: "A:", ShortLabel: "A:", Value: "1"},
		{FullLabel: "B:", ShortLabel: "B:", Value: "2"},
		{FullLabel: "C:", ShortLabel: "C:", Value: "3"},
		{FullLabel: "D:", ShortLabel: "D:", Value: "4"},
	})
	attrs := makePanel("attrs", "Attrs", 24, borderFull, []PanelRow{
		{FullLabel: "Str:", ShortLabel: "S:", Value: "10"},
		{FullLabel: "Vit:", ShortLabel: "V:", Value: "11"},
	})
	wealth := makePanel("wealth", "Wealth", 14, borderFull, []PanelRow{
		{FullLabel: "Gold:", ShortLabel: "G:", Value: "100"},
		{FullLabel: "Bank:", ShortLabel: "B:", Value: "500"},
	})
	training := makePanel("training", "Training", 14, borderFull, []PanelRow{
		{FullLabel: "Trn:", ShortLabel: "T:", Value: "3"},
		{FullLabel: "Sta:", ShortLabel: "S:", Value: "1"},
	})

	layout := makeLayout(1,
		[][]*Panel{{info}},
		[][]*Panel{{attrs}, {wealth, training}},
	)

	output := layout.Render()
	rendered := splitLines(output)

	// All lines must have the same visual width (the key correctness property).
	require.True(t, len(rendered) > 0)
	w0 := panelVisualWidth(rendered[0])
	for i, line := range rendered {
		assert.Equal(t, w0, panelVisualWidth(line), "line %d visual width differs", i)
	}

	// Content from all four panels must appear somewhere.
	full := strings.Join(rendered, "\n")
	assert.Contains(t, full, "Info")
	assert.Contains(t, full, "Attrs")
	assert.Contains(t, full, "Wealth")
	assert.Contains(t, full, "Training")
}
