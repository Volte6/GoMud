package discord

// Reference: https://birdie0.github.io/discord-webhooks-guide/discord_webhook.html

// Discord message payload for non-rich content
// This object has many more fields, see reference

type Color int

const (
	Default           Color = 0        // #000000
	Aqua              Color = 1752220  // #1ABC9C
	DarkAqua          Color = 1146986  // #11806A
	Green             Color = 5763719  // #57F287
	DarkGreen         Color = 2067276  // #1F8B4C
	Blue              Color = 3447003  // #3498DB
	DarkBlue          Color = 2123412  // #206694
	Purple            Color = 10181046 // #9B59B6
	DarkPurple        Color = 7419530  // #71368A
	LuminousVividPink Color = 15277667 // #E91E63
	DarkVividPink     Color = 11342935 // #AD1457
	Gold              Color = 15844367 // #F1C40F
	DarkGold          Color = 12745742 // #C27C0E
	Orange            Color = 15105570 // #E67E22
	DarkOrange        Color = 11027200 // #A84300
	Red               Color = 15548997 // #ED4245
	DarkRed           Color = 10038562 // #992D22
	Grey              Color = 9807270  // #95A5A6
	DarkGrey          Color = 9936031  // #979C9F
	DarkerGrey        Color = 8359053  // #7F8C8D
	LightGrey         Color = 12370112 // #BCC0C0
	Navy              Color = 3426654  // #34495E
	DarkNavy          Color = 2899536  // #2C3E50
	Yellow            Color = 16776960 // #FFFF00
)

// Discord message payload for rich content
type webHookPayload struct {
	Username  string  `json:"username,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Content   string  `json:"content,omitempty"`
	Embeds    []embed `json:"embeds,omitempty"`
}

// These objects have many more fields, see reference
type embed struct {
	Author embedAuthor `json:"author,omitempty"`

	Title       string       `json:"title,omitempty"`
	URL         string       `json:"url,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       Color        `json:"color,omitempty"`
	Fields      []embedField `json:"fields,omitempty"`

	Image     embedURL    `json:"image,omitempty"`
	Thumbnail embedURL    `json:"thumbnail,omitempty"`
	Footer    embedFooter `json:"footer,omitempty"`
}

type embedAuthor struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}
type embedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type embedURL struct {
	URL string `json:"url,omitempty"`
}

type embedField struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}
