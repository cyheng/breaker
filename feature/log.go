package feature

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"runtime"
)

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

func (l *LoggerConfig) InitLog() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:03:04",
		ForceColors:     true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			//处理文件名
			fileName := path.Base(frame.File)
			return frame.Function, fileName
		},
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
