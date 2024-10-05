package term

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
)
