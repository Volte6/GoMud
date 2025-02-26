package events

type EventLogger struct{}

func (e EventLogger) Println(v ...any) {
	AddToQueue(Log{Data: v})
}

func GetLogger() EventLogger {
	return EventLogger{}
}
