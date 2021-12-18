package feature

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
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
	LogMaxDays int `ini:"log_max_days" json:"log_max_days"`
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
	l.InitLog()
}

func (l *LoggerConfig) NewFeature() (Feature, error) {
	return nil, errors.New("not support feature")
}

func (l *LoggerConfig) InitLog() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:03:04",
	})
	if l.LogWay == "console" {
		log.SetOutput(os.Stdout)
	} else {
		ljack := &lumberjack.Logger{
			Filename:   l.LogFile,
			MaxSize:    100, // megabytes
			MaxBackups: 52,
			MaxAge:     l.LogMaxDays, //days
			Compress:   false,        // disabled by default
		}
		io.MultiWriter(ljack, os.Stdout)
	}
	_, err := log.ParseLevel(l.LogLevel)
	if err != nil {
		panic(err)
	}
}
