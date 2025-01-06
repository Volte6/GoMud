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
