package term

import (
	"bytes"
	"testing"
)

// TestGenerateSimpleVariable tests generating an MSDP message with a simple variable.
func TestGenerateSimpleVariable(t *testing.T) {
	variables := map[string]interface{}{
		"HEALTH": "100",
	}

	expectedOutput := `TELNET_IAC TELNET_SB MSDP MSDP_VAR "HEALTH" MSDP_VAL "100" TELNET_IAC TELNET_SE`

	data, err := GenerateMSDP(variables)
	if err != nil {
		t.Fatalf("GenerateMSDP failed: %v", err)
	}

	formatted, err := FormatMSDPPacket(data)
	if err != nil {
		t.Fatalf("FormatMSDPPacket failed: %v", err)
	}

	if formatted != expectedOutput {
		t.Errorf("Expected: %s\nGot:      %s", expectedOutput, formatted)
	}
}

// TestGenerateArrayVariable tests generating an MSDP message with an array.
func TestGenerateArrayVariable(t *testing.T) {
	variables := map[string]interface{}{
		"REPORTABLE_VARIABLES": []interface{}{"HEALTH", "MANA", "MOVEMENT"},
	}

	expectedOutput := `TELNET_IAC TELNET_SB MSDP MSDP_VAR "REPORTABLE_VARIABLES" MSDP_VAL MSDP_ARRAY_OPEN MSDP_VAL "HEALTH" MSDP_VAL "MANA" MSDP_VAL "MOVEMENT" MSDP_ARRAY_CLOSE TELNET_IAC TELNET_SE`

	data, err := GenerateMSDP(variables)
	if err != nil {
		t.Fatalf("GenerateMSDP failed: %v", err)
	}

	formatted, err := FormatMSDPPacket(data)
	if err != nil {
		t.Fatalf("FormatMSDPPacket failed: %v", err)
	}

	if formatted != expectedOutput {
		t.Errorf("Expected: %s\nGot:      %s", expectedOutput, formatted)
	}
}

// Helper function to remove extra spaces for comparison
func removeExtraSpaces(s string) string {
	var buffer bytes.Buffer
	inSpace := false
	for _, r := range s {
		if r == ' ' {
			if !inSpace {
				buffer.WriteRune(r)
				inSpace = true
			}
		} else {
			buffer.WriteRune(r)
			inSpace = false
		}
	}
	return buffer.String()
}
