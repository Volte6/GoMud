package term

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const (
	MSDP IACByte = 69 // https://tintin.mudhalla.net/protocols/msdp/

	MSDP_VAR IACByte = 1
	MSDP_VAL IACByte = 2

	MSDP_TABLE_OPEN  IACByte = 3
	MSDP_TABLE_CLOSE IACByte = 4

	MSDP_ARRAY_OPEN  IACByte = 5
	MSDP_ARRAY_CLOSE IACByte = 6
)

/*
Handshake
When a client connects to an MSDP enabled server the server should send IAC WILL MSDP.
The client should respond with either IAC DO MSDP or IAC DONT MSDP.
Once the server receives IAC DO MSDP both the client and the server can send MSDP sub-negotiations.
*/

var (
	///////////////////////////
	// GMCP COMMANDS
	///////////////////////////
	MsdpEnable  = TerminalCommand{[]byte{TELNET_IAC, TELNET_WILL, MSDP}, []byte{}} // Indicates the server wants to enable MSDP.
	MsdpDisable = TerminalCommand{[]byte{TELNET_IAC, TELNET_WONT, MSDP}, []byte{}} // Indicates the server wants to disable MSDP.

	MsdpAccept = TerminalCommand{[]byte{TELNET_IAC, TELNET_DO, MSDP}, []byte{}}   // Indicates the client accepts MSDP sub-negotiations.
	MsdpRefuse = TerminalCommand{[]byte{TELNET_IAC, TELNET_DONT, MSDP}, []byte{}} // Indicates the client refuses MSDP sub-negotiations.

	// Send variable data?
	MsdpVar = TerminalCommand{[]byte{TELNET_IAC, TELNET_SB, MSDP}, []byte{MSDP_VAL, TELNET_IAC, TELNET_SE}} // Indicates the client refuses MSDP sub-negotiations.
	// Payload would be: MSDP_VAR, "VARNAME", MSDP_VAL, "VARVALUE"
)

// IAC SB MSDP MSDP_VAR "SEND" MSDP_VAL "HEALTH" IAC SE

// GenerateMSDP generates an MSDP byte stream from a map[string]interface{}.
func GenerateMSDP(variables map[string]interface{}) ([]byte, error) {
	var buffer bytes.Buffer

	buffer.Write([]byte{TELNET_IAC, TELNET_SB, MSDP})

	for varName, val := range variables {
		buffer.WriteByte(MSDP_VAR)
		writeString(&buffer, varName)
		buffer.WriteByte(MSDP_VAL)
		if err := writeValue(&buffer, val); err != nil {
			return nil, err
		}
	}

	buffer.Write([]byte{TELNET_IAC, TELNET_SE})

	return buffer.Bytes(), nil
}

// writeString writes a string to the buffer.
func writeString(buffer *bytes.Buffer, s string) {
	buffer.WriteString(s)
}

// writeValue writes a value to the buffer.
func writeValue(buffer *bytes.Buffer, val interface{}) error {
	switch v := val.(type) {
	case string:
		writeString(buffer, v)
	case map[string]interface{}:
		buffer.WriteByte(MSDP_TABLE_OPEN)
		for key, value := range v {
			buffer.WriteByte(MSDP_VAR)
			writeString(buffer, key)
			buffer.WriteByte(MSDP_VAL)
			if err := writeValue(buffer, value); err != nil {
				return err
			}
		}
		buffer.WriteByte(MSDP_TABLE_CLOSE)
	case []interface{}:
		buffer.WriteByte(MSDP_ARRAY_OPEN)
		for _, item := range v {
			buffer.WriteByte(MSDP_VAL)
			if err := writeValue(buffer, item); err != nil {
				return err
			}
		}
		buffer.WriteByte(MSDP_ARRAY_CLOSE)
	default:
		return errors.New("unsupported MSDP value type")
	}
	return nil
}

// FormatMSDPPacket formats an MSDP packet into a single-line string as per the specification.
func FormatMSDPPacket(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	var parts []string

	// Mapping of control bytes to their names
	controlBytes := map[byte]string{
		TELNET_IAC:       "TELNET_IAC",
		TELNET_SB:        "TELNET_SB",
		TELNET_SE:        "TELNET_SE",
		MSDP_VAR:         "MSDP_VAR",
		MSDP_VAL:         "MSDP_VAL",
		MSDP_TABLE_OPEN:  "MSDP_TABLE_OPEN",
		MSDP_TABLE_CLOSE: "MSDP_TABLE_CLOSE",
		MSDP_ARRAY_OPEN:  "MSDP_ARRAY_OPEN",
		MSDP_ARRAY_CLOSE: "MSDP_ARRAY_CLOSE",
		MSDP:             "MSDP",
	}

	endStringControlBytes := map[byte]string{
		TELNET_IAC:       "TELNET_IAC",
		TELNET_SB:        "TELNET_SB",
		TELNET_SE:        "TELNET_SE",
		MSDP_VAR:         "MSDP_VAR",
		MSDP_VAL:         "MSDP_VAL",
		MSDP_TABLE_OPEN:  "MSDP_TABLE_OPEN",
		MSDP_TABLE_CLOSE: "MSDP_TABLE_CLOSE",
		MSDP_ARRAY_OPEN:  "MSDP_ARRAY_OPEN",
		MSDP_ARRAY_CLOSE: "MSDP_ARRAY_CLOSE",
	}
	// Read and process each byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break // End of data
		}

		if name, ok := controlBytes[b]; ok {
			fmt.Printf(`[%d]`, b)
			parts = append(parts, name)
		} else {

			reader.UnreadByte()
			var buf bytes.Buffer
			for {
				b, err := reader.ReadByte()
				if err != nil {
					return "", err
				}

				fmt.Printf(`[%d?]`, b)
				if _, ok := endStringControlBytes[b]; ok {
					reader.UnreadByte()
					break
				}
				fmt.Printf(`[%d-%s]`, b, string(b))
				buf.WriteByte(b)
			}

			// Enclose the string in quotes
			parts = append(parts, fmt.Sprintf("\"%s\"", buf.String()))
		}
	}

	// Join all parts into a single line
	return strings.Join(parts, " "), nil
}
