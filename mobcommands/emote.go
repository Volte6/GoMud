package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"
)

var (
	emoteAliases map[string]string = map[string]string{
		"armcross": "crosses their arms.",
		"backflip": "does a backflip.",
		"beam":     "beams with pride.",
		"blink":    "blinks in surprise.",
		"blush":    "blushes slightly.",
		"bounce":   "bounces up and down.",
		"bow":      "bows gracefully.",
		"brood":    "broods in the corner.",
		"chew":     "chews thoughtfully.",
		"cheer":    "cheers loudly.",
		"chuckle":  "chuckles softly.",
		"clap":     "claps enthusiastically.",
		"cringe":   "cringes in embarrassment.",
		"cry":      "cries softly.",
		"dance":    "starts dancing.",
		"daydream": "daydreams wistfully.",
		"doze":     "dozes off for a moment.",
		"drum":     "drums their fingers.",
		"duck":     "ducks to avoid something.",
		"eyeroll":  "rolls their eyes.",
		"eyebrow":  "raises an eyebrow.",
		"facepalm": "facepalms in disbelief.",
		"flail":    "flails their arms.",
		"flex":     "flexes their muscles.",
		"flinch":   "flinches unexpectedly.",
		"flirt":    "is feeling flirty.",
		"flutter":  "flutters their eyelashes.",
		"frown":    "frowns deeply.",
		"giggle":   "giggles softly.",
		"glare":    "glares menacingly.",
		"grin":     "grins cheekily.",
		"groan":    "groans in frustration.",
		"headache": "rubs their temples. They seem to be getting a headache.",
		"hum":      "hums a familiar tune.",
		"jump":     "jumps in excitement.",
		"juggle":   "juggles a few items skillfully.",
		"laugh":    "laughs heartily.",
		"listen":   "listens intently.",
		"meditate": "meditates peacefully.",
		"murmur":   "murmurs something under their breath.",
		"nod":      "nods in agreement.",
		"pace":     "paces back and forth.",
		"point":    "points at something.",
		"ponder":   "is pondering something.",
		"pout":     "pouts adorably.",
		"prance":   "prances around.",
		"roar":     "roars mightily.",
		"salute":   "salutes respectfully.",
		"scratch":  "scratches their head.",
		"shake":    "shakes their head.",
		"shiver":   "shivers from the cold... or perhaps something else.",
		"shudder":  "shudders in fear.",
		"shrug":    "shrugs nonchalantly.",
		"shush":    "shushes everyone.",
		"sigh":     "sighs deeply.",
		"sing":     "sings a tune.",
		"sit":      "sits down for a think.",
		"skip":     "skips joyfully.",
		"slap":     "slaps their forehead.",
		"slouch":   "slouches lazily.",
		"smile":    "smiles warmly.",
		"snicker":  "snickers quietly.",
		"sniff":    "sniffs the air.",
		"snore":    "snores loudly.",
		"spin":     "spins around dizzyingly.",
		"stand":    "stands up straight.",
		"stomp":    "stomps their foot.",
		"stretch":  "stretches their limbs.",
		"stumble":  "stumbles a bit.",
		"swim":     "swims around.",
		"tap":      "taps their foot impatiently.",
		"think":    "thinks hard.",
		"tilt":     "tilts their head curiously.",
		"tremble":  "trembles in anticipation.",
		"trip":     "trips over their own feet.",
		"twirl":    "twirls around with a flourish.",
		"wave":     "waves.",
		"whine":    "whines pitifully.",
		"whistle":  "whistles a catchy melody.",
		"yawn":     "yawns sleepily.",
	}
)

func Emote(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	// Don't bother if no players are present
	if room.PlayerCt() < 1 {
		response.Handled = true
		return response, nil
	}

	if len(rest) == 0 {
		response.SendRoomMessage(mob.Character.RoomId,
			fmt.Sprintf(`<ansi fg="mobname">%s</ansi> emotes.`, mob.Character.Name),
			true)
		response.Handled = true
		return response, nil
	}

	response.SendRoomMessage(mob.Character.RoomId,
		fmt.Sprintf(`<ansi fg="mobname">%s</ansi> <ansi fg="blue">%s</ansi>`, mob.Character.Name, rest),
		true)

	response.Handled = true
	return response, nil
}
