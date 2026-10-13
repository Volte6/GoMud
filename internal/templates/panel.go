package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/ansitags"
	"github.com/mattn/go-runewidth"
	"gopkg.in/yaml.v2"
)

type borderStyle string

const (
	borderFull borderStyle = "full"
	borderOpen borderStyle = "open"

	panelPad         = 1 // spaces inside border on each side of content
	defaultColumnGap = 2 // spaces between columns when columns > 1
)

// borderChars holds the characters used to draw a panel border.
type borderChars struct {
	TopLeft          string // e.g. ┌
	TopRight         string // e.g. ┐
	BottomLeft       string // e.g. └
	BottomRight      string // e.g. ┘
	Horizontal       string // e.g. ─  (top border horizontal fill)
	HorizontalBottom string // e.g. ─  (bottom border horizontal fill)
	VerticalLeft     string // e.g. │  (left side of content rows)
	VerticalRight    string // e.g. │  (right side of content rows)
}

var (
	charsetSingle  = borderChars{"┌", "┐", "└", "┘", "─", "─", "│", "│"}
	charsetDouble  = borderChars{"╔", "╗", "╚", "╝", "═", "═", "║", "║"}
	charsetRounded = borderChars{"╭", "╮", "╰", "╯", "─", "─", "│", "│"}
)

// charsetForName returns the borderChars for a named preset or a literal
// string.
//
// Named presets: "single" (default), "double", "rounded".
//
// Literal format - exactly 8 Unicode code points in order:
//
//	[TopLeft][HorizontalTop][TopRight][VerticalLeft][VerticalRight][BottomLeft][HorizontalBottom][BottomRight]
//
// Example: "\u2554\u2550\u2557\u2551\u2502\u255a\u2500\u2518"  (double top/left, single bottom/right)
func charsetForName(name string) borderChars {
	switch strings.ToLower(name) {
	case "double":
		return charsetDouble
	case "rounded":
		return charsetRounded
	case "", "single":
		return charsetSingle
	}
	// Try to parse as an 8-rune literal:
	// TopLeft, HorizontalTop, TopRight, VerticalLeft, VerticalRight,
	// BottomLeft, HorizontalBottom, BottomRight.
	runes := []rune(name)
	if len(runes) == 8 {
		return borderChars{
			TopLeft:          string(runes[0]),
			Horizontal:       string(runes[1]),
			TopRight:         string(runes[2]),
			VerticalLeft:     string(runes[3]),
			VerticalRight:    string(runes[4]),
			BottomLeft:       string(runes[5]),
			HorizontalBottom: string(runes[6]),
			BottomRight:      string(runes[7]),
		}
	}
	return charsetSingle
}

// PanelRow is one label+value line inside a panel.
// The renderer uses FullLabel when it fits the panel width, ShortLabel otherwise.
// Set Blank to true to insert an empty spacer line; label and value are ignored.
// When WrapWidth > 0, the value is word-wrapped at that visual width; continuation
// lines are indented to align with the value column of the first line.
type PanelRow struct {
	FullLabel  string
	ShortLabel string
	Value      string
	Blank      bool
	WrapWidth  int
}

// Panel holds the rows for one titled box. Obtain via PanelLayout.Panel(id).
// width is the total border-inclusive panel width (border chars + padding + content).
// When width > 0, the panel targets that exact outer width and wraps values to fit;
// it will expand beyond width only when a label+value pair cannot fit otherwise.
type Panel struct {
	id         string
	title      string // raw title string, may contain ANSI tags; used verbatim in the top border
	width      int    // border-inclusive target width; 0 means size to content
	border     borderStyle
	chars      borderChars
	columns    int // 1 (default) or 2: how many label+value pairs share a line
	columnGap  int // spaces between columns when columns > 1
	labelWidth int // if > 0, all labels are right-padded to this visual width
	rows       []PanelRow
}

// innerWidth returns the target inner content width derived from p.width.
// Returns 0 when p.width is 0 (size-to-content mode).
func (p *Panel) innerWidth() int {
	if p.width <= 0 {
		return 0
	}
	w := p.width - 2 - 2*panelPad // subtract 2 border chars and 2 padding chars
	if w < 0 {
		w = 0
	}
	return w
}

// Add appends a label+value row and returns the panel for chaining.
// When the panel has a non-zero width, values are wrapped to fit within the panel.
func (p *Panel) Add(fullLabel, shortLabel, value string) *Panel {
	p.rows = append(p.rows, PanelRow{
		FullLabel:  fullLabel,
		ShortLabel: shortLabel,
		Value:      value,
	})
	return p
}

// AddWithWrapWidth appends a label+value row with an explicit value wrap width.
// When the value's visual width exceeds wrapWidth, it is wrapped onto continuation
// lines indented to align with the value column of the first line.
// Pass -1 to disable wrapping entirely for this row, even when the panel has a
// fixed width that would otherwise trigger automatic wrapping.
func (p *Panel) AddWithWrapWidth(fullLabel, shortLabel, value string, wrapWidth int) *Panel {
	p.rows = append(p.rows, PanelRow{
		FullLabel:  fullLabel,
		ShortLabel: shortLabel,
		Value:      value,
		WrapWidth:  wrapWidth,
	})
	return p
}

// AddBlank appends an empty spacer row and returns the panel for chaining.
func (p *Panel) AddBlank() *Panel {
	p.rows = append(p.rows, PanelRow{Blank: true})
	return p
}

// layoutSlot is one vertical column in the layout.
// It contains one or more horizontal rows of panels stacked top-to-bottom.
type layoutSlot struct {
	rows [][]*Panel
}

// renderSlot renders a slot's stacked rows into a single []string column.
// All lines are padded to the same visual width (the widest row in the slot)
// so that the slot can be composed side-by-side with other slots cleanly.
func renderSlot(slot *layoutSlot, gap int) []string {
	var result []string
	for _, row := range slot.rows {
		result = append(result, composePanelGroup(row, gap)...)
	}
	if len(result) == 0 {
		return result
	}
	// Find the widest line in this slot.
	maxW := 0
	for _, line := range result {
		if w := panelVisualWidth(line); w > maxW {
			maxW = w
		}
	}
	// Right-pad any shorter lines so all lines have the same visual width.
	for i, line := range result {
		if w := panelVisualWidth(line); w < maxW {
			result[i] = line + strings.Repeat(" ", maxW-w)
		}
	}
	return result
}

// PanelLayout is a loaded layout skeleton populated with data and ready to render.
type PanelLayout struct {
	border       borderStyle
	chars        borderChars
	gap          int
	margin       int               // spaces prepended to every output line
	defaultColor string            // optional ANSI fg color/alias wrapping entire output
	slots        []*layoutSlot     // top-level horizontal slots (columns)
	byID         map[string]*Panel // fast lookup by panel id
}

// Panel returns the named panel for data population.
// If the id is not defined in the layout, a no-op panel is returned and an
// error is logged so the caller continues without panicking.
func (l *PanelLayout) Panel(id string) *Panel {
	p, ok := l.byID[id]
	if !ok {
		mudlog.Error("panel layout", "error", fmt.Sprintf("no panel with id %q", id))
		return &Panel{border: l.border, chars: l.chars, columns: 1, columnGap: defaultColumnGap}
	}
	return p
}

// Render synthesizes all slots into a single terminal string.
func (l *PanelLayout) Render() string {
	if len(l.slots) == 0 {
		return ""
	}

	// Render each slot to its own []string column.
	rendered := make([][]string, len(l.slots))
	for i, slot := range l.slots {
		rendered[i] = renderSlot(slot, l.gap)
	}

	// Find the maximum line count across all slots.
	maxLines := 0
	for _, col := range rendered {
		if len(col) > maxLines {
			maxLines = len(col)
		}
	}

	// Pad shorter slots with blank lines of the correct visual width.
	for i, col := range rendered {
		if len(col) >= maxLines {
			continue
		}
		// Measure the visual width of this slot from its first line.
		w := 0
		if len(col) > 0 {
			w = panelVisualWidth(col[0])
		}
		blank := strings.Repeat(" ", w)
		for len(rendered[i]) < maxLines {
			rendered[i] = append(rendered[i], blank)
		}
	}

	gapStr := strings.Repeat(" ", l.gap)
	out := make([]string, maxLines)
	for row := 0; row < maxLines; row++ {
		parts := make([]string, len(rendered))
		for i, col := range rendered {
			parts[i] = col[row]
		}
		out[row] = strings.Join(parts, gapStr)
	}
	if l.margin > 0 {
		prefix := strings.Repeat(" ", l.margin)
		for i, line := range out {
			out[i] = prefix + line
		}
	}
	result := strings.Join(out, "\n")
	if l.defaultColor != "" {
		result = fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, l.defaultColor, result)
	}
	return result
}

// ---------------------------------------------------------------------------
// YAML definition structs
// ---------------------------------------------------------------------------

// panelDef is the YAML structure for a single panel entry.
type panelDef struct {
	ID        string `yaml:"id"`
	Title     string `yaml:"title"`
	Width     int    `yaml:"width"`      // border-inclusive total panel width; 0 means size to content
	Columns   int    `yaml:"columns"`    // optional, default 1
	ColumnGap int    `yaml:"column_gap"` // optional, default 2
	Charset   string `yaml:"charset"`    // optional, overrides layout-level charset
}

// panelRowDef is a horizontal group of panels in the YAML.
type panelRowDef struct {
	Panels []panelDef `yaml:"panels"`
}

// slotDef is one vertical slot (column) in the YAML.
// It contains one or more rows of panels stacked top-to-bottom.
type slotDef struct {
	Rows []panelRowDef `yaml:"rows"`
}

// panelLayoutDef is the top-level YAML structure.
type panelLayoutDef struct {
	Border       string    `yaml:"border"`
	Gap          int       `yaml:"gap"`
	Margin       int       `yaml:"margin"`        // optional left margin applied to every output line
	Charset      string    `yaml:"charset"`       // optional: "single" (default), "double", "rounded", or 8-rune literal
	DefaultColor string    `yaml:"default_color"` // optional: ANSI color code or alias to wrap entire output
	Slots        []slotDef `yaml:"slots"`
}

// LayoutSlot is the exported handle for a slot, used by the scripting layer.
type LayoutSlot = layoutSlot

// NewPanelLayout creates a PanelLayout programmatically without loading a YAML file.
func NewPanelLayout(border, charset string, gap, margin int) *PanelLayout {
	b := borderFull
	if borderStyle(border) == borderOpen {
		b = borderOpen
	}
	if gap < 0 {
		gap = 0
	}
	return &PanelLayout{
		border: b,
		gap:    gap,
		margin: margin,
		chars:  charsetForName(charset),
		byID:   make(map[string]*Panel),
	}
}

// AddSlot appends a new empty slot to the layout and returns it.
func (l *PanelLayout) AddSlot() *LayoutSlot {
	slot := &layoutSlot{}
	l.slots = append(l.slots, slot)
	return slot
}

// AddPanelsToSlot appends a new row of panels (one per id) to the given slot.
// Panels are created with default settings. Use Panel(id) to configure them further.
func (l *PanelLayout) AddPanelsToSlot(slot *LayoutSlot, ids ...string) {
	var row []*Panel
	for _, id := range ids {
		p := &Panel{
			id:        id,
			border:    l.border,
			chars:     l.chars,
			columns:   1,
			columnGap: defaultColumnGap,
		}
		l.byID[id] = p
		row = append(row, p)
	}
	slot.rows = append(slot.rows, row)
}

// SetCharset sets the border character set for this panel, overriding the
// layout-level charset. Accepts a named preset ("single", "double", "rounded")
// or a 7-rune literal string in the order:
// TopLeft, Horizontal, TopRight, VerticalLeft, VerticalRight, BottomLeft, BottomRight.
// Example literal: "╔═╗║║╚╝". An unrecognised value falls back to "single".
func (p *Panel) SetCharset(name string) *Panel { p.chars = charsetForName(name); return p }

// SetTitle sets the panel's title string verbatim.
func (p *Panel) SetTitle(title string) *Panel { p.title = title; return p }

// SetWidth sets the total border-inclusive panel width.
// The panel will target this exact outer width and wrap values to fit within it.
// The panel will expand beyond this width only when a label+value pair cannot fit.
func (p *Panel) SetWidth(w int) *Panel { p.width = w; return p }

// SetLabelWidth sets a fixed visual width that all labels are padded to.
// When non-zero, every label is right-padded with spaces to this width before
// rendering, so values always start at the same column regardless of label length.
// ANSI tags in labels are accounted for correctly.
func (p *Panel) SetLabelWidth(w int) *Panel { p.labelWidth = w; return p }

// SetColumns sets the number of label+value pairs per rendered line (1 or 2).
func (p *Panel) SetColumns(n int) *Panel {
	if n < 1 {
		n = 1
	}
	p.columns = n
	return p
}

// SetColumnGap sets the spaces between columns when columns > 1.
func (p *Panel) SetColumnGap(n int) *Panel {
	if n < 0 {
		n = 0
	}
	p.columnGap = n
	return p
}

// PanelLayoutInfo holds metadata about a discovered panel layout file.
type PanelLayoutInfo struct {
	// Name is the relative path under panel-layouts/ without the .yaml extension.
	// Example: "character/status"
	Name string
	// YAML is the raw file content.
	YAML string
}

// ListPanelLayouts returns all panel layout files found under the panel-layouts/
// subdirectory of the data files directory. Each entry's Name is the relative
// path without the .yaml extension (e.g. "character/status").
func ListPanelLayouts() ([]PanelLayoutInfo, error) {
	dataFiles := string(configs.GetFilePathsConfig().DataFiles)
	root := dataFiles + "/panel-layouts"

	var result []PanelLayoutInfo
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		name := strings.TrimSuffix(rel, ".yaml")
		// Normalise Windows separators.
		name = strings.ReplaceAll(name, "\\", "/")

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		result = append(result, PanelLayoutInfo{Name: name, YAML: string(data)})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list panel layouts: %w", err)
	}
	return result, nil
}

// PanelValidationIssue describes a single problem found during layout validation.
type PanelValidationIssue struct {
	Message string `json:"message"`
}

// ValidatePanelLayout parses the given YAML and returns a list of structural
// issues. An empty slice means the layout is valid. A non-nil error means the
// YAML itself could not be parsed.
func ValidatePanelLayout(yamlData string) ([]PanelValidationIssue, error) {
	var def panelLayoutDef
	if err := yaml.Unmarshal([]byte(yamlData), &def); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	var issues []PanelValidationIssue
	add := func(msg string) { issues = append(issues, PanelValidationIssue{Message: msg}) }

	if def.Border != "" && def.Border != string(borderFull) && def.Border != string(borderOpen) {
		add(fmt.Sprintf("unknown border style %q (valid: \"full\", \"open\")", def.Border))
	}

	validCharsets := map[string]bool{"single": true, "double": true, "rounded": true, "": true}
	if !validCharsets[def.Charset] && len([]rune(def.Charset)) != 8 {
		add(fmt.Sprintf("unknown charset %q (valid: \"single\", \"double\", \"rounded\", or 8-rune literal)", def.Charset))
	}

	if def.DefaultColor != "" && strings.TrimSpace(def.DefaultColor) == "" {
		add("default_color must not be blank whitespace; omit the field to disable it")
	}

	if len(def.Slots) == 0 {
		add("layout has no slots defined")
	}

	seenIDs := make(map[string]bool)
	for si, sd := range def.Slots {
		if len(sd.Rows) == 0 {
			add(fmt.Sprintf("slot %d has no rows", si+1))
		}
		for ri, rd := range sd.Rows {
			if len(rd.Panels) == 0 {
				add(fmt.Sprintf("slot %d, row %d has no panels", si+1, ri+1))
			}
			for pi, pd := range rd.Panels {
				loc := fmt.Sprintf("slot %d, row %d, panel %d", si+1, ri+1, pi+1)
				if pd.ID == "" {
					add(fmt.Sprintf("%s: missing id", loc))
				} else if seenIDs[pd.ID] {
					add(fmt.Sprintf("%s: duplicate panel id %q", loc, pd.ID))
				} else {
					seenIDs[pd.ID] = true
				}
				if pd.Columns < 0 {
					add(fmt.Sprintf("%s (id=%q): columns must be >= 0", loc, pd.ID))
				}
				if pd.Width < 0 {
					add(fmt.Sprintf("%s (id=%q): width must be >= 0", loc, pd.ID))
				}
				if pd.Charset != "" && !validCharsets[pd.Charset] && len([]rune(pd.Charset)) != 8 {
					add(fmt.Sprintf("%s (id=%q): unknown charset %q", loc, pd.ID, pd.Charset))
				}
			}
		}
	}

	return issues, nil
}

// PreviewPanelLayout parses the given YAML definition and renders a text
// preview by filling every panel with at least two dummy rows.
// When stripAnsi is true the output is plain ASCII (ANSI tags removed from
// titles). When false, titles are kept verbatim so the caller can render the
// ANSI tags with a client-side library.
func PreviewPanelLayout(yamlData string, stripAnsi bool) (string, error) {
	var def panelLayoutDef
	if err := yaml.Unmarshal([]byte(yamlData), &def); err != nil {
		return "", fmt.Errorf("preview panel layout: %w", err)
	}

	border := borderFull
	if borderStyle(def.Border) == borderOpen {
		border = borderOpen
	}
	chars := charsetForName(def.Charset)
	gap := def.Gap
	if gap < 0 {
		gap = 0
	}

	layout := &PanelLayout{
		border:       border,
		chars:        chars,
		gap:          gap,
		margin:       def.Margin,
		defaultColor: def.DefaultColor,
		byID:         make(map[string]*Panel),
	}
	if stripAnsi {
		layout.defaultColor = ""
	}

	for _, sd := range def.Slots {
		slot := &layoutSlot{}
		for _, rd := range sd.Rows {
			panelCount := len(rd.Panels)
			var rowPanels []*Panel
			for _, pd := range rd.Panels {
				cols := pd.Columns
				if cols < 1 {
					cols = 1
				}
				colGap := pd.ColumnGap
				if colGap < 1 {
					colGap = defaultColumnGap
				}
				panelChars := chars
				if pd.Charset != "" {
					panelChars = charsetForName(pd.Charset)
				}
				title := pd.Title
				if stripAnsi {
					title = panelVisualTitle(pd.Title)
				}
				p := &Panel{
					id:        pd.ID,
					title:     title,
					width:     pd.Width,
					border:    border,
					chars:     panelChars,
					columns:   cols,
					columnGap: colGap,
				}
				// Sole panel in its row defines its own height; give it more rows
				// so the preview is representative. Panels with siblings get 2.
				p.rows = append(p.rows, previewDummyRows(cols, panelCount == 1, !stripAnsi)...)
				layout.byID[pd.ID] = p
				rowPanels = append(rowPanels, p)
			}
			slot.rows = append(slot.rows, rowPanels)
		}
		layout.slots = append(layout.slots, slot)
	}

	return layout.Render(), nil
}

// panelVisualTitle strips ANSI tags from a title string and returns the
// printable text only, suitable for plain-text preview output.
func panelVisualTitle(title string) string {
	stripped := ansitags.Parse(title, ansitags.StripTags)
	return strings.TrimSpace(stripped)
}

// previewDummyRows returns dummy PanelRows for preview purposes.
// alone=true means the panel is the sole panel in its row and will define the
// row height; it gets 4 rows. Panels with horizontal siblings get 2 rows.
// Multi-column panels double the count so both columns appear populated.
// When withAnsi is true, labels and values are wrapped in ANSI colour tags.
func previewDummyRows(columns int, alone bool, withAnsi bool) []PanelRow {
	base := 2
	if alone {
		base = 4
	}
	count := base * columns
	rows := make([]PanelRow, count)
	for i := range rows {
		label := fmt.Sprintf("label-%d", i+1)
		value := fmt.Sprintf("value-%d", i+1)
		if withAnsi {
			label = fmt.Sprintf(`<ansi fg="yellow">%s</ansi>`, label)
			value = fmt.Sprintf(`<ansi fg="green-bold">%s</ansi>`, value)
		}
		rows[i] = PanelRow{
			FullLabel:  label,
			ShortLabel: label,
			Value:      value,
		}
	}
	return rows
}

// SavePanelLayout writes yamlData to the panel layout file for the given name.
// name is relative to panel-layouts/ without the .yaml extension.
// The file must already exist; this function does not create new layout files.
func SavePanelLayout(name, yamlData string) error {
	dataFiles := string(configs.GetFilePathsConfig().DataFiles)
	path := dataFiles + "/panel-layouts/" + name + ".yaml"

	// Verify the file exists before overwriting.
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("panel layout %q: %w", name, err)
	}

	// Validate the YAML parses correctly before writing.
	var def panelLayoutDef
	if err := yaml.Unmarshal([]byte(yamlData), &def); err != nil {
		return fmt.Errorf("panel layout %q: invalid YAML: %w", name, err)
	}

	if err := os.WriteFile(path, []byte(yamlData), 0644); err != nil {
		return fmt.Errorf("panel layout %q: %w", name, err)
	}
	return nil
}

// LoadPanelLayout loads a panel layout definition from the datafiles directory.
// name is relative to the panel-layouts/ subdirectory and has no extension.
// Example: LoadPanelLayout("character/status")
func LoadPanelLayout(name string) (*PanelLayout, error) {
	dataFiles := string(configs.GetFilePathsConfig().DataFiles)
	path := dataFiles + "/panel-layouts/" + name + ".yaml"

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("panel layout %q: %w", name, err)
	}

	var def panelLayoutDef
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("panel layout %q: %w", name, err)
	}

	border := borderFull
	if borderStyle(def.Border) == borderOpen {
		border = borderOpen
	}

	chars := charsetForName(def.Charset)

	gap := def.Gap
	if gap < 0 {
		gap = 0
	}

	layout := &PanelLayout{
		border:       border,
		chars:        chars,
		gap:          gap,
		margin:       def.Margin,
		defaultColor: def.DefaultColor,
		byID:         make(map[string]*Panel),
	}

	for _, sd := range def.Slots {
		slot := &layoutSlot{}
		for _, rd := range sd.Rows {
			var rowPanels []*Panel
			for _, pd := range rd.Panels {
				cols := pd.Columns
				if cols < 1 {
					cols = 1
				}
				colGap := pd.ColumnGap
				if colGap < 1 {
					colGap = defaultColumnGap
				}
				panelChars := chars
				if pd.Charset != "" {
					panelChars = charsetForName(pd.Charset)
				}
				p := &Panel{
					id:        pd.ID,
					title:     pd.Title,
					width:     pd.Width,
					border:    border,
					chars:     panelChars,
					columns:   cols,
					columnGap: colGap,
				}
				layout.byID[pd.ID] = p
				rowPanels = append(rowPanels, p)
			}
			slot.rows = append(slot.rows, rowPanels)
		}
		layout.slots = append(layout.slots, slot)
	}

	return layout, nil
}

// ---------------------------------------------------------------------------
// Width helpers
// ---------------------------------------------------------------------------

// panelVisualWidth returns the visible terminal width of s, stripping ANSI tags.
func panelVisualWidth(s string) int {
	return runewidth.StringWidth(ansitags.Parse(s, ansitags.StripTags))
}

// panelInnerWidth calculates the inner content width for a panel.
// When p.width > 0, the target inner width is p.innerWidth(); the panel will
// use that target but expand if content cannot fit.
// For single-column panels the result is the max of the target and the widest row,
// where wrapped rows contribute only their wrap width (not the full value width).
// For multi-column panels the result is the max of the target and
// twice the widest half-column (each half = label+1+value), plus the column gap.
func panelInnerWidth(p *Panel) int {
	target := p.innerWidth()

	if p.columns < 2 {
		width := target
		for _, row := range p.rows {
			if row.Blank {
				continue
			}
			lw := panelVisualWidth(row.FullLabel)
			if p.labelWidth > lw {
				lw = p.labelWidth
			}
			vw := panelVisualWidth(row.Value)
			// When the panel has a target width, compute the available value
			// width and treat that as the effective wrap width for sizing.
			// Only apply wrapping if there is actually space for the label.
			// WrapWidth < 0 disables wrapping for this row entirely.
			effectiveWrap := row.WrapWidth
			if effectiveWrap >= 0 && target > 0 {
				availForValue := target - lw - 1
				if availForValue >= 1 {
					// There is room for at least one char of value after the label.
					if effectiveWrap <= 0 || availForValue < effectiveWrap {
						effectiveWrap = availForValue
					}
				}
				// If availForValue < 1, the label alone is wider than the target;
				// don't apply any wrap — let the content drive expansion.
			}
			if effectiveWrap > 0 && vw > effectiveWrap {
				vw = effectiveWrap
			}
			needed := lw + 1 + vw
			if needed > width {
				width = needed
			}
		}
		return width
	}
	// Multi-column: find the widest single cell, then total = 2*colWidth + gap.
	widestCell := 0
	for _, row := range p.rows {
		if row.Blank {
			continue
		}
		lw := panelVisualWidth(row.FullLabel)
		vw := panelVisualWidth(row.Value)
		cell := lw + 1 + vw
		if cell > widestCell {
			widestCell = cell
		}
	}
	colTarget := (target - p.columnGap) / 2
	if colTarget < 0 {
		colTarget = 0
	}
	colWidth := colTarget
	if widestCell > colWidth {
		colWidth = widestCell
	}
	total := colWidth*2 + p.columnGap
	if total > target && target > 0 {
		return total
	}
	if target > total {
		return target
	}
	return total
}

// chooseLabel returns the label to use for a row given the available inner width.
// It prefers FullLabel; falls back to ShortLabel if FullLabel doesn't fit alongside the value.
func chooseLabel(row PanelRow, innerWidth int) string {
	flw := panelVisualWidth(row.FullLabel)
	vw := panelVisualWidth(row.Value)
	if flw+1+vw <= innerWidth {
		return row.FullLabel
	}
	return row.ShortLabel
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

// renderCellContent renders one label+value cell right-padded to colWidth.
func renderCellContent(row PanelRow, colWidth int) string {
	label := chooseLabel(row, colWidth)
	lw := panelVisualWidth(label)
	vw := panelVisualWidth(row.Value)
	rightPad := colWidth - lw - 1 - vw
	if rightPad < 0 {
		rightPad = 0
	}
	return label + " " + row.Value + strings.Repeat(" ", rightPad)
}

// renderPanel renders a single panel into a slice of terminal lines.
// Each line is a complete row including border characters and padding.
// Lines do not end with a newline.
func renderPanel(p *Panel) []string {
	inner := panelInnerWidth(p)
	c := p.chars

	var lines []string

	// Top border.
	// With title:    TopLeft + H + " " + title + " " + H*n + TopRight
	// Without title: TopLeft + H*(inner + 2*panelPad) + TopRight
	var topBorder string
	if p.title == "" {
		topBorder = c.TopLeft + strings.Repeat(c.Horizontal, inner+2*panelPad) + c.TopRight
	} else {
		titleVW := panelVisualWidth(p.title)
		// visible structure: TL + H + " " + title + " " + H*n + TR
		// that is 1 + 1 + 1 + titleVW + 1 + n + 1 = inner + 2*panelPad + 2
		dashCount := inner + 2*panelPad + 2 - 1 - 1 - 1 - titleVW - 1 - 1
		if dashCount < 0 {
			dashCount = 0
		}
		topBorder = c.TopLeft + c.Horizontal + " " + p.title + " " + strings.Repeat(c.Horizontal, dashCount) + c.TopRight
	}
	lines = append(lines, topBorder)

	nRows := len(p.rows)

	if p.columns < 2 {
		// Single-column layout.
		for i, row := range p.rows {
			isFirst := i == 0
			isLast := i == nRows-1
			lines = append(lines, renderSingleColumnLines(p, row, inner, isFirst, isLast)...)
		}
	} else {
		// Multi-column layout: pair up rows.
		colWidth := (inner - p.columnGap) / 2
		gapStr := strings.Repeat(" ", p.columnGap)

		for i := 0; i < nRows; i += p.columns {
			isFirst := i == 0
			isLast := i+p.columns >= nRows

			rowA := p.rows[i]
			hasB := i+1 < nRows

			var content string
			if rowA.Blank {
				content = strings.Repeat(" ", inner+2*panelPad)
			} else if !hasB || p.rows[i+1].Blank {
				// Odd row out or right cell is blank: span full width.
				content = strings.Repeat(" ", panelPad) +
					renderCellContent(rowA, inner) +
					strings.Repeat(" ", panelPad)
			} else {
				rowB := p.rows[i+1]
				cellA := renderCellContent(rowA, colWidth)
				cellB := renderCellContent(rowB, colWidth)
				content = strings.Repeat(" ", panelPad) +
					cellA + gapStr + cellB +
					strings.Repeat(" ", panelPad)
			}

			if p.border == borderFull || isFirst || isLast {
				lines = append(lines, c.VerticalLeft+content+c.VerticalRight)
			} else {
				lines = append(lines, " "+content+" ")
			}
		}
	}

	// Bottom border
	lines = append(lines, c.BottomLeft+strings.Repeat(c.HorizontalBottom, inner+2*panelPad)+c.BottomRight)

	return lines
}

// renderSingleColumnLines renders one content row for a single-column panel,
// returning one or more lines. When a wrap width is determined (from the row's
// WrapWidth or the panel's target width), the value is split and continuation
// lines are indented to align with the value column of the first line.
func renderSingleColumnLines(p *Panel, row PanelRow, inner int, isFirst, isLast bool) []string {
	c := p.chars
	borderLine := func(content string) string {
		if p.border == borderFull || isFirst || isLast {
			return c.VerticalLeft + content + c.VerticalRight
		}
		return " " + content + " "
	}

	if row.Blank {
		return []string{borderLine(strings.Repeat(" ", inner+2*panelPad))}
	}

	label := chooseLabel(row, inner)
	lw := panelVisualWidth(label)
	if p.labelWidth > lw {
		label = label + strings.Repeat(" ", p.labelWidth-lw)
		lw = p.labelWidth
	}

	// valueIndent is the number of spaces to prepend on continuation lines so
	// that they align with the value column of the first line.
	// Layout: panelPad + label + " " + value
	valueIndent := panelPad + lw + 1

	// Determine the effective wrap width for this row.
	// WrapWidth < 0 disables wrapping entirely for this row.
	// WrapWidth > 0 uses that as the explicit wrap width.
	// WrapWidth == 0 falls back to the panel's target width.
	// If your wrap is greater than visible width it is adjusted.
	effectiveWrap := row.WrapWidth
	visibleWidth := inner - lw - 1
	if effectiveWrap == 0 && p.width > 0 {
		if visibleWidth > 0 {
			effectiveWrap = visibleWidth
		}
	} else if effectiveWrap > visibleWidth {
		effectiveWrap = visibleWidth
	}

	var chunks []string
	if effectiveWrap > 0 {
		for _, chunk := range ansitags.SplitStringOnSpaces(row.Value, effectiveWrap, true) {
			if panelVisualWidth(chunk) > 0 {
				chunks = append(chunks, chunk)
			}
		}
	}
	if len(chunks) == 0 {
		chunks = []string{row.Value}
	}

	var result []string
	for ci, chunk := range chunks {
		vw := panelVisualWidth(chunk)
		var content string
		if ci == 0 {
			rightPad := inner - lw - 1 - vw
			if rightPad < 0 {
				rightPad = 0
			}
			content = strings.Repeat(" ", panelPad) +
				label + " " + chunk +
				strings.Repeat(" ", rightPad) +
				strings.Repeat(" ", panelPad)
		} else {
			rightPad := inner + 2*panelPad - valueIndent - vw
			if rightPad < 0 {
				rightPad = 0
			}
			content = strings.Repeat(" ", valueIndent) +
				chunk +
				strings.Repeat(" ", rightPad)
		}
		// Only the first and last logical rows of the panel get border treatment
		// from isFirst/isLast; continuation lines are never first or last.
		if ci == 0 {
			result = append(result, borderLine(content))
		} else {
			// Continuation lines: open border unless the panel uses full borders.
			if p.border == borderFull {
				result = append(result, c.VerticalLeft+content+c.VerticalRight)
			} else {
				result = append(result, " "+content+" ")
			}
		}
	}
	return result
}

// composePanelGroup renders a group of panels side by side into a slice of lines.
// Panels of different heights are padded so all have the same number of lines.
// The bottom border is always the last line; blank padding is inserted above it.
func composePanelGroup(panels []*Panel, gap int) []string {
	if len(panels) == 0 {
		return nil
	}

	rendered := make([][]string, len(panels))
	for i, p := range panels {
		rendered[i] = renderPanel(p)
	}

	// Find the maximum line count.
	maxLines := 0
	for _, lines := range rendered {
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	// Pad each panel's lines to maxLines.
	// The last line is always the bottom border; insert blank filler lines before it.
	for i, lines := range rendered {
		if len(lines) >= maxLines {
			continue
		}
		p := panels[i]
		inner := panelInnerWidth(p)
		blankContent := strings.Repeat(" ", inner+2*panelPad)

		var blankLine string
		if p.border == borderFull {
			blankLine = p.chars.VerticalLeft + blankContent + p.chars.VerticalRight
		} else {
			blankLine = " " + blankContent + " "
		}

		needed := maxLines - len(lines)
		bottom := lines[len(lines)-1]
		middle := lines[:len(lines)-1]
		padding := make([]string, needed)
		for j := range padding {
			padding[j] = blankLine
		}
		rendered[i] = append(middle, append(padding, bottom)...)
	}

	gapStr := strings.Repeat(" ", gap)
	result := make([]string, maxLines)
	for row := 0; row < maxLines; row++ {
		var parts []string
		for _, lines := range rendered {
			parts = append(parts, lines[row])
		}
		result[row] = strings.Join(parts, gapStr)
	}
	return result
}
