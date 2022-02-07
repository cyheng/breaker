package feature

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

//Config is base config
type Config struct {
	LoggerConfig
}
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

func (l *LoggerConfig) InitLog() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:03:04",
		ForceColors:     true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			//处理文件名
			fileName := path.Base(frame.File)
			return fmt.Sprintf("%s:%d", frame.Function, frame.Line), fileName
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

type BridgeConfig struct {
	Config
	ServerAddr        string `ini:"server_addr"`
	LocalPort         int    `ini:"local_port"`
	RemotePort        int    `ini:"remote_port"`
	ProxyName         string `ini:"proxy_name"`
	HeartbeatInterval int64  `ini:"heartbeat_interval" `
}

func (b *BridgeConfig) OnInit() {
	b.Config.OnInit()
	if b.ServerAddr == "" {
		panic("breaker address can not be empty")
	}
	if b.LocalPort < 0 || b.LocalPort > 65535 {
		panic("invalid local port[0-65535]")
	}
	if b.RemotePort < 0 || b.RemotePort > 65535 {
		panic("invalid remote port[0-65535]")
	}
	if b.HeartbeatInterval < 0 {
		panic("invalid HeartbeatInterval, can't less than 0")
	}
	if b.HeartbeatInterval == 0 {
		b.HeartbeatInterval = 5
	}
	if b.ProxyName == "" {
		b.ProxyName = b.ServerAddr + "_to_" + strconv.Itoa(b.LocalPort)
	}
}

type PortalConfig struct {
	Config
	ServerAddr string `ini:"server_addr"`
}

func (c *PortalConfig) OnInit() {
	c.Config.OnInit()
	if c.ServerAddr == "" {
		c.ServerAddr = "0.0.0.0:80"
	}
}
