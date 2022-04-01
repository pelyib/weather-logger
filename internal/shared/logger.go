package shared

import (
	"fmt"
	"log"
	"os"
)

const App_Logger string = "logger"
const App_Http string = "http"
const App_Commander string = "commander"

type Logger interface {
	Info(msg string)
	Warning(msg string)
	Error(msg string)
}

type cliLogger struct {
	l log.Logger
}

type nullLogger struct{}

func (l cliLogger) Info(msg string) {
	l.print(fmt.Sprintf("\033[32m[INFO]\033[0m %s", msg))
}

func (l cliLogger) Warning(msg string) {
	l.print(fmt.Sprintf("\033[33m[WARNING]\033[0m %s", msg))
}

func (l cliLogger) Error(msg string) {
	l.print(fmt.Sprintf("\033[31m[ERROR]\033[0m %s", msg))
}

func (l cliLogger) print(msg string) {
	l.l.Println(msg)
}

func MakeCliLogger(app string, prefix string) Logger {
	l := log.New(os.Stdout, fmt.Sprintf("[%s][%s]", app, prefix), log.Lmsgprefix|log.Ltime)

	return cliLogger{l: *l}
}

func (l nullLogger) Info(msg string)    {}
func (l nullLogger) Warning(msg string) {}
func (l nullLogger) Error(msg string)   {}

func MakeNullLogger() Logger {
	return nullLogger{}
}
