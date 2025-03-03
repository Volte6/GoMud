package discord

// Reference: https://birdie0.github.io/discord-webhooks-guide/discord_webhook.html

// Discord message payload for non-rich content
// This object has many more fields, see reference
type simpleMessage struct {
	Content string `json:"content"`
}

// Discord message payload for rich content
type richMessage struct {
	Embeds []embed `json:"embeds"`
}

// These objects have many more fields, see reference
type embed struct {
	Description string `json:"description"`
	Color       int32    `json:"color"`
}
