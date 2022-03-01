package shared

import (
	"fmt"
	"log"
)

const level_info = "Info"
const level_warning = "Warning"
const level_error = "error"

type Logger interface {
	Info(msg []byte)
	Warning(msg []byte)
	Error(msg []byte)
}

type cliLogger struct {
	prefix string
}

type nullLogger struct{}

func (l cliLogger) Info(msg []byte) {
	l.print(fmt.Sprintf("\033[32m%s\033[0m", string(msg)))
}

func (l cliLogger) Warning(msg []byte) {
	l.print(fmt.Sprintf("%s\033[0m", string(msg)))
}

func (l cliLogger) Error(msg []byte) {
	l.print(fmt.Sprintf("%s\033[0m", string(msg)))
}

func (l cliLogger) print(msg string) {
	log.Println(fmt.Sprintf("%s | %s", l.prefix, msg))
}

func MakeCliLogger(app string, prefix string) Logger {
	log.SetPrefix(app)
	log.SetFlags(log.Llongfile | log.Lmsgprefix | log.Ltime)
	return cliLogger{prefix: prefix}
}

func (l nullLogger) Info(msg []byte)    {}
func (l nullLogger) Warning(msg []byte) {}
func (l nullLogger) Error(msg []byte)   {}

func MakeNullLogger() Logger {
	return nullLogger{}
}
