package logger

import "fmt"

const (
	MessageTypeSuccess = "Success"
	MessageTypeFail    = "Fail"
	MessageTypeWarning = "Warning"
	MessageTypeInfo    = "Info"
)

type ConsoleLogger struct {
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

func (logger *ConsoleLogger) Info(data string) {
	fmt.Printf("[%s] %v\n", MessageTypeInfo, data)
}

func (logger *ConsoleLogger) Fail(data string) {
	fmt.Printf("[%s] %v\n", MessageTypeFail, data)
}

func (logger *ConsoleLogger) Warning(data string) {
	fmt.Printf("[%s] %v\n", MessageTypeWarning, data)
}

func (logger *ConsoleLogger) Success(data string) {
	fmt.Printf("[%s] %v\n", MessageTypeSuccess, data)
}

func (logger *ConsoleLogger) Close() bool {
	return true
}

type NullLogger struct {
}

func NewNullLogger() *NullLogger {
	return &NullLogger{}
}

func (logger *NullLogger) Info(data string) {

}

func (logger *NullLogger) Fail(data string) {

}

func (logger *NullLogger) Warning(data string) {

}

func (logger *NullLogger) Success(data string) {

}
