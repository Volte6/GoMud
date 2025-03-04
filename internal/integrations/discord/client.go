// Package discord
//
// This package provides programatic access to discord messages integrated with GoMud.
//
// References:
// https://leovoel.github.io/embed-visualizer/
// https://birdie0.github.io/discord-webhooks-guide/discord_webhook.html
package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/events"
)

var (
	WebhookUrl  string
	initialized bool
)

// Initializes and sets the webhook so we can send messages to discord
// and registers listeners to listen for events
func Init(webhookUrl string) {
	if initialized {
		return
	}

	WebhookUrl = webhookUrl
	registerListeners()
	initialized = true
}

func registerListeners() {
	events.RegisterListener(events.PlayerSpawn{}, HandlePlayerSpawn)
	events.RegisterListener(events.PlayerDespawn{}, HandlePlayerDespawn)
	events.RegisterListener(events.Log{}, HandleLogs)
}

// Sends an embed message to discord which includes a colored bar to the left
// hexColor should be specified as a string in this format "#000000"
func SendRichMessage(message string, hexColor string) error {
	if !initialized {
		return errors.New("Discord client was not initialized.")
	}

	if strings.HasPrefix(hexColor, "#") {
		hexColor = hexColor[1:]
	}

	color, err := strconv.ParseInt(hexColor, 16, 32)
	if err != nil {
		message := fmt.Sprintf("Invalid color specified, expected format #000000")
		return errors.New(message)
	}

	payload := richMessage{
		Embeds: []embed{
			{
				Description: message,
				Color:       int32(color),
			},
		},
	}

	marshalled, err := json.Marshal(payload)
	if err != nil {
		message := fmt.Sprintf("Couldn't marshal discord message")
		return errors.New(message)
	}

	return send(marshalled)
}

// Sends a simple message to discord
func SendMessage(message string) error {
	if !initialized {
		return errors.New("Discord client was not initialized.")
	}

	payload := simpleMessage{
		Content: message,
	}

	marshalled, err := json.Marshal(payload)
	if err != nil {
		message := fmt.Sprintf("Couldn't marshal discord message")
		return errors.New(message)
	}

	return send(marshalled)
}

func send(marshalled []byte) error {
	request, err := http.NewRequest("POST", WebhookUrl, bytes.NewReader(marshalled))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		message := fmt.Sprintf("Couldn't send POST request to discord.")
		return errors.New(message)
	}

	// Expect 204 No Content reply
	if response.StatusCode != 204 {
		message := fmt.Sprintf("Expected discord to send status code 204, got %v.", response.StatusCode)
		return errors.New(message)
	}

	return nil
}
