package term

const (
	MSP IACByte = 90 // https://www.zuggsoft.com/zmud/msp.htm
)

/*
Handshake
When a client connects to an MSDP enabled server the server should send IAC WILL MSDP.
The client should respond with either IAC DO MSDP or IAC DONT MSDP.
Once the server receives IAC DO MSDP both the client and the server can send MSDP sub-negotiations.
*/

var (
	MspEnable  = TerminalCommand{[]byte{TELNET_IAC, TELNET_WILL, MSP}, []byte{}} // Indicates the server wants to enable MSP.
	MspDisable = TerminalCommand{[]byte{TELNET_IAC, TELNET_WONT, MSP}, []byte{}} // Indicates the server wants to disable MSP.

	MspAccept = TerminalCommand{[]byte{TELNET_IAC, TELNET_DO, MSP}, []byte{}}   // Indicates the client accepts MSP
	MspRefuse = TerminalCommand{[]byte{TELNET_IAC, TELNET_DONT, MSP}, []byte{}} // Indicates the client refuses MSP

	MspCommand = TerminalCommand{[]byte{TELNET_IAC, TELNET_SB, MSP}, []byte{TELNET_IAC, TELNET_SE}} // Send via TELNET MSP Command
)

func IsMSPCommand(b []byte) bool {
	return len(b) > 2 && b[0] == TELNET_IAC && b[2] == MSP
}
