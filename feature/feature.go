package feature

import "errors"

type Feature interface {
}

type FeatureConfig interface {
	OnInit()
	NewFeature() (Feature, error)
}

var configFactory = make(map[string]FeatureConfig)

//Loader config to feature
//type Loader func(config interface{}) (Feature, error)

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
