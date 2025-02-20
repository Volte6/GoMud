package rooms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volte6/gomud/internal/characters"
)

func TestRoom_AddCorpse(t *testing.T) {
	r := &Room{}
	assert.Empty(t, r.Corpses, "Expected no corpses initially")

	corpse := Corpse{
		MobId:        0,
		UserId:       123,
		Character:    characters.Character{Name: "TestPlayer"},
		RoundCreated: 10,
	}
	r.AddCorpse(corpse)
	assert.Len(t, r.Corpses, 1, "Expected exactly one corpse after adding")
	assert.Equal(t, corpse, r.Corpses[0], "Expected the added corpse to match")
}

func TestRoom_RemoveCorpse(t *testing.T) {
	r := &Room{}
	corpse1 := Corpse{
		MobId:        0,
		UserId:       123,
		Character:    characters.Character{Name: "PlayerOne"},
		RoundCreated: 10,
	}
	corpse2 := Corpse{
		MobId:        456,
		UserId:       0,
		Character:    characters.Character{Name: "MobOne"},
		RoundCreated: 11,
	}

	r.AddCorpse(corpse1)
	r.AddCorpse(corpse2)

	// Removing existing corpse
	removed := r.RemoveCorpse(corpse1)
	assert.True(t, removed, "Expected to remove an existing corpse successfully")
	assert.Len(t, r.Corpses, 1, "Expected exactly one corpse remaining")

	// Try removing a corpse that does not exist
	nonExistent := Corpse{
		MobId:        999,
		UserId:       999,
		Character:    characters.Character{Name: "Ghost"},
		RoundCreated: 99,
	}
	removed = r.RemoveCorpse(nonExistent)
	assert.False(t, removed, "Expected removal to fail for non-existent corpse")
	assert.Len(t, r.Corpses, 1, "Expected no change in corpses")
}

func TestRoom_FindCorpse(t *testing.T) {
	r := &Room{}

	playerCorpse := Corpse{
		UserId:       123,
		Character:    characters.Character{Name: "PlayerOne"},
		RoundCreated: 5,
		Prunable:     false,
	}
	mobCorpse := Corpse{
		MobId:        456,
		Character:    characters.Character{Name: "MobOne"},
		RoundCreated: 6,
		Prunable:     false,
	}
	r.AddCorpse(playerCorpse)
	r.AddCorpse(mobCorpse)

	// Exact search
	found, ok := r.FindCorpse("PlayerOne corpse")
	assert.True(t, ok, "Expected to find player corpse by exact name")
	assert.Equal(t, "PlayerOne", found.Character.Name, "Expected found corpse to match the correct character")

	// Searching for mob
	found, ok = r.FindCorpse("MobOne corpse")
	assert.True(t, ok, "Expected to find mob corpse by exact name")
	assert.Equal(t, "MobOne", found.Character.Name, "Expected found corpse to match the correct character")

	// Searching partial name (depends on your util.FindMatchIn logic)
	found, ok = r.FindCorpse("player")
	assert.True(t, ok, "Expected to find a close match for player corpse")
	assert.Equal(t, "PlayerOne", found.Character.Name, "Expected found corpse to be the player's")

	// Non-existent
	found, ok = r.FindCorpse("NonExistent")
	assert.False(t, ok, "Expected not to find a missing corpse")
}

func TestFindNoun(t *testing.T) {
	// Create a room with various noun mappings (including aliases).
	r := &Room{
		Nouns: map[string]string{
			"torch":            "a fiery torch",
			"torchAlias":       ":torch", // alias -> :torch means "torch"
			"lamp":             "an illuminating lamp",
			"lampAlias":        ":lamp", // alias -> :lamp means "lamp"
			"candles":          "some wax candles",
			"pony":             "a small horse",
			"mystery":          "just a riddle",
			"secret":           ":mystery",                   // chain alias -> :mystery
			"projector screen": "a wide, matte-white screen", // multi-word matching
		},
	}

	// Table-driven tests.
	tests := []struct {
		name          string
		inputNoun     string
		wantFoundNoun string
		wantDesc      string
	}{
		{
			name:          "Direct match (torch)",
			inputNoun:     "torch",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},
		{
			name:          "Direct match (candles)",
			inputNoun:     "candles",
			wantFoundNoun: "candles",
			wantDesc:      "some wax candles",
		},
		{
			name:          "Direct match (pony)",
			inputNoun:     "pony",
			wantFoundNoun: "pony",
			wantDesc:      "a small horse",
		},
		{
			name:          "Alias match (torchAlias -> :torch)",
			inputNoun:     "torchAlias",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},
		{
			name:          "Alias match (lampAlias -> :lamp)",
			inputNoun:     "lampAlias",
			wantFoundNoun: "lamp",
			wantDesc:      "an illuminating lamp",
		},
		{
			name:          "Chain alias (secret -> :mystery)",
			inputNoun:     "secret",
			wantFoundNoun: "mystery",
			wantDesc:      "just a riddle",
		},
		{
			name:          "Plural form by adding s (pony -> ponies)",
			inputNoun:     "ponies",
			wantFoundNoun: "pony",
			wantDesc:      "a small horse",
		},
		{
			name:          "Plural form by adding es (torches)",
			inputNoun:     "torches",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},

		{
			name:          "Multi-word input, second word valid (red torches)",
			inputNoun:     "red torches",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},
		{
			name:          "Multi-word input, multi-word match",
			inputNoun:     "projector screen",
			wantFoundNoun: "projector screen",
			wantDesc:      "a wide, matte-white screen",
		},
		{
			name:          "Single-word input, multi-word match 1",
			inputNoun:     "projector",
			wantFoundNoun: "projector screen",
			wantDesc:      "a wide, matte-white screen",
		},
		{
			name:          "Single-word input, multi-word match 2",
			inputNoun:     "screen",
			wantFoundNoun: "projector screen",
			wantDesc:      "a wide, matte-white screen",
		},
		{
			name:          "Multi-word input, first word valid (torch something)",
			inputNoun:     "torch something",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},
		{
			name:          "No match",
			inputNoun:     "gibberish",
			wantFoundNoun: "",
			wantDesc:      "",
		},
		{
			name:          "Multi-word no match (foo bar)",
			inputNoun:     "foo bar",
			wantFoundNoun: "",
			wantDesc:      "",
		},
	}

	// Run each sub-test.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFound, gotDesc := r.FindNoun(tt.inputNoun)
			if gotFound != tt.wantFoundNoun || gotDesc != tt.wantDesc {
				t.Errorf("FindNoun(%q) = (%q, %q), want (%q, %q)",
					tt.inputNoun, gotFound, gotDesc, tt.wantFoundNoun, tt.wantDesc)
			}
		})
	}
}
