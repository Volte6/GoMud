package events

type eventLogger struct{}

func (e eventLogger) Println(level string, v ...any) {
	AddToQueue(Log{Level: level, Data: v})
}

func GetLogger() eventLogger {
	return eventLogger{}
}
