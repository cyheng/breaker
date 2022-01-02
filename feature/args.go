package feature

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

type BridgeConfig struct {
	Config
	ServerAddr string `ini:"server_addr"`
	LocalPort  int    `ini:"local_port"`
	RemotePort int    `ini:"remote_port"`
	ProxyName  string `ini:"proxy_name"`
}

type PortalConfig struct {
	Config
	ServerAddr string `ini:"server_addr"`
}
