package events

type EventLogger struct{}

func (e EventLogger) Println(level string, v ...any) {
	AddToQueue(Log{Level: level, Data: v})
}

func GetLogger() EventLogger {
	return EventLogger{}
}
