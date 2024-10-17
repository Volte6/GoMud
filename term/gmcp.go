package term

import (
	"strconv"
	"strings"
)

const (
	GMCP IACByte = 201 // https://tintin.mudhalla.net/protocols/gmcp/
)

/*
Handshake:

When a client connects to a GMCP enabled server the server should send IAC WILL GMCP.
The client should respond with either IAC DO GMCP or IAC DONT GMCP.
Once the server receives IAC DO GMCP both the client and the server can send GMCP sub-negotiations.


////

Example MSDP over GMCP handshake

server - IAC WILL GMCP
client - IAC   DO GMCP
client - IAC   SB GMCP 'MSDP {"LIST" : "COMMANDS"}' IAC SE
server - IAC   SB GMCP 'MSDP {"COMMANDS" : ["LIST", "REPORT", "RESET", "SEND", "UNREPORT"]}' IAC SE

The single quote characters mean that the encased text is a string, the single quotes themselves should not be send.
*/

var (
	///////////////////////////
	// GMCP COMMANDS
	///////////////////////////
	GmcpEnable  = TerminalCommand{[]byte{TELNET_IAC, TELNET_WILL, GMCP}, []byte{}} // Indicates the server wants to enable GMCP.
	GmcpDisable = TerminalCommand{[]byte{TELNET_IAC, TELNET_WONT, GMCP}, []byte{}} // Indicates the server wants to disable GMCP.

	GmcpAccept = TerminalCommand{[]byte{TELNET_IAC, TELNET_DO, GMCP}, []byte{}}   // Indicates the client accepts GMCP sub-negotiations.
	GmcpRefuse = TerminalCommand{[]byte{TELNET_IAC, TELNET_DONT, GMCP}, []byte{}} // Indicates the client refuses GMCP sub-negotiations.

	GmcpPayload = TerminalCommand{[]byte{TELNET_IAC, TELNET_SB, GMCP}, []byte{TELNET_IAC, TELNET_SE}} // Wrapper for sending GMCP payloads

	// If not found here, will ignore incoming message
	SupportedGMCP = map[string]struct{}{
		`External.Discord.Hello`: {},
		`Core.Hello`:             {},
		`Core.Supports.Set`:      {},
		`Core.Supports.Remove`:   {},
		`Char.Login`:             {},
	}
)

func IsGMCPCommand(b []byte) bool {
	return len(b) > 2 && b[0] == TELNET_IAC && b[2] == GMCP
}

type GMCPHello struct {
	Client  string
	Version string
}

type GMCPDiscord struct {
	User    string
	Private bool
}

type GMCPSupportsSet []string

// Returns a map of module name to version number
func (s GMCPSupportsSet) GetSupportedModules() map[string]int {

	ret := map[string]int{}

	for _, entry := range s {

		parts := strings.Split(entry, ` `)
		if len(parts) == 2 {
			ret[parts[0]], _ = strconv.Atoi(parts[1])
		}

	}

	return ret
}

type GMCPSupportsRemove = []string

type GMCPLogin struct {
	Name     string
	Password string
}
