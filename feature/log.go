package feature

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
)

type LoggerConfig struct {
	// LogFile specifies a file where logs will be written to. This value will
	// only be used if LogWay is set appropriately. By default, this value is
	// "console".
	LogFile string `ini:"log_file" json:"log_file"`
	// LogWay specifies the way logging is managed. Valid values are "console"
	// or "file". If "console" is used, logs will be printed to stdout. If
	// "file" is used, logs will be printed to LogFile. By default, this value
	// is "console".
	LogWay string `ini:"log_way" json:"log_way"`
	// LogLevel specifies the minimum log level. Valid values are "trace",
	// "debug", "info", "warn", and "error". By default, this value is "info".
	LogLevel string `ini:"log_level" json:"log_level"`
	// LogMaxDays specifies the maximum number of days to store log information
	// before deletion. This is only used if LogWay == "file". By default, this
	// value is 0.
	LogMaxDays int64 `ini:"log_max_days" json:"log_max_days"`
}

func (l *LoggerConfig) OnInit() {
	if l.LogFile == "" {
		l.LogFile = "breaker.log"
	}
	if l.LogWay == "" {
		l.LogWay = "console"
	}
	if l.LogLevel == "" {
		l.LogLevel = "info"
	}
	InitLog(l.LogWay, l.LogFile, l.LogLevel)
}

func (l *LoggerConfig) NewFeature() (Feature, error) {
	return nil, errors.New("not support feature")
}

func InitLog(logWay string, logFile string, logLevel string) {
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:03:04",
	})
	SetLogFile(logWay, logFile)
	SetLogLevel(logLevel)
}

// logWay: such as file or console
func SetLogFile(logWay string, logFile string) {
	if logWay == "console" {
		log.SetOutput(os.Stdout)
	} else {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		log.SetOutput(file)
	}
}

// value: error, warning, info, debug
func SetLogLevel(logLevel string) {
	_, err := log.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}
}
