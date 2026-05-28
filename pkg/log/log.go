package log

import (
	"os"
	"time"

	ll "github.com/moorada/log"
)

var pathDir = "logs"
var pathFile = ""

func init() {
	ll.LevelNames[ll.DEBUG] = "DEBUG"
	ll.LevelNames[ll.INFO] = "INFO"
	ll.LevelNames[ll.IMPORTANT] = "INFO"
	ll.LevelNames[ll.WARNING] = "WARN"
	ll.LevelNames[ll.ERROR] = "ERROR"
	ll.LevelNames[ll.FATAL] = "FATAL"
}

// ActiveConsoleLog configures stdout logging. When plain is true, ANSI colors are omitted.
func ActiveConsoleLog(plain bool) (err error) {
	config := ll.FormatConfigBasic
	if plain {
		config.Format = "{time} {level:name} {message}"
		err = ll.AddOutput("", ll.INFO, config, true)
	} else {
		config.Format = "{time} {level:color}{level:name}{reset} {message}"
		err = ll.AddOutput("", ll.INFO, config, false)
	}
	return err
}

func RemoveConsoleLog() {
	ll.RemoveOutput("")
}

// ReconfigureConsole replaces stdout logging (e.g. after config is loaded).
func ReconfigureConsole(plain bool) error {
	RemoveConsoleLog()
	return ActiveConsoleLog(plain)
}

func RemoveDebugLog() {
	ll.RemoveOutput(pathFile)
}

func ActiveDebugLog() (err error) {

	pathFile = pathDir + "/" + time.Now().Format("2January2006-15:04:05") + ".log"
	if _, err := os.Stat(pathDir); os.IsNotExist(err) {
		_ = os.MkdirAll(pathDir, os.ModePerm)
	}

	config := ll.FormatConfigBasic
	config.Format = "{datetime} {level:name} {message}"
	err = ll.AddOutput(pathFile, ll.DEBUG, config, true)
	return err
}

func Close() {
	ll.CloseOutputs()
}

func Debug(format string, args ...interface{}) {
	ll.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	ll.Info(format, args...)
	syncStdout()
}

func Important(format string, args ...interface{}) {
	ll.Important(format, args...)
}

func Warning(format string, args ...interface{}) {
	ll.Warning(format, args...)
	syncStdout()
}

func Error(format string, args ...interface{}) {
	ll.Error(format, args...)
	syncStdout()
}

func syncStdout() {
	_ = os.Stdout.Sync()
}

func Fatal(format string, args ...interface{}) {
	ll.Fatal(format, args...)
}
