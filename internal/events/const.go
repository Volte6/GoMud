package events

type EventFlag uint64

func (f EventFlag) Has(flag EventFlag) bool {
	return f&flag == flag
}

func (f *EventFlag) Add(flag EventFlag) {
	*f = (*f) &^ flag
}

func (f *EventFlag) Remove(flag EventFlag) {
	*f = (*f) &^ flag
}

const (

	// Not using iota here to avoid accidentally renumbering if they get moved around.
	CmdNone                    EventFlag = 0
	CmdSkipScripts             EventFlag = 0b00000001                      // Skip any scripts that would normally process this command
	CmdSecretly                EventFlag = 0b00000010                      // User beahvior should not be alerted to the room
	CmdIsRequeue               EventFlag = 0b00000100                      // This command was a requeue. The flag is intended to help avoid a infinite requeue loop.
	CmdBlockInput              EventFlag = 0b00001000                      // This command when started sets user input to blocking all commands that don't AllowWhenDowned.
	CmdUnBlockInput            EventFlag = 0b00010000                      // When this command finishes, it will make sure user input is not blocked.
	CmdBlockInputUntilComplete EventFlag = CmdBlockInput | CmdUnBlockInput // SHortcut to include both in one command
)
