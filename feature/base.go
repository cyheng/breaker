package feature

import "errors"

func (b *Config) OnInit() {
	b.LoggerConfig.InitLog()
}

type FeatureConfig interface {
	OnInit()
	ServiceName() string
}

var configFactory = make(map[string]FeatureConfig)

func RegisterConfig(featureName string, cfg FeatureConfig) {
	configFactory[featureName] = cfg
}

func GetConfig(featureName string) (FeatureConfig, error) {
	cfg, ok := configFactory[featureName]
	if !ok {
		return nil, errors.New("can not find config:" + featureName)
	}
	return cfg, nil
}
