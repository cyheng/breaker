package config

import (
	"breaker/feature"
	"errors"
	"github.com/go-ini/ini"
	"os"
	"path/filepath"
	"strings"
)

type Loader func([]byte) ([]feature.Feature, error)

var Loaders = make(map[string]Loader)

func init() {
	RegisterLoader("ini", func(b []byte) (result []feature.Feature, err error) {
		f, err := ini.LoadSources(ini.LoadOptions{
			Insensitive:         false,
			InsensitiveSections: false,
			InsensitiveKeys:     false,
			IgnoreInlineComment: true,
			AllowBooleanKeys:    true,
		}, b)
		if err != nil {
			return nil, err
		}
		for _, item := range f.Sections() {
			secName := item.Name()
			config, e := feature.GetConfig(secName)
			if e != nil {
				continue
			}
			item.MapTo(config)
			config.OnInit()
			feature, err := config.NewFeature()
			if err != nil {
				continue
			}
			result = append(result, feature)

		}
		return
	})
}

func RegisterLoader(ext string, c Loader) {
	Loaders[ext] = c
}

func LoadFromFile(cfgFile string) ([]feature.Feature, error) {
	ext := strings.ToLower(filepath.Ext(cfgFile))
	ext = strings.TrimLeft(ext, ".")
	b, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}
	// only support ini format
	loader, ok := Loaders[ext]
	if !ok {
		return nil, errors.New("config extent not define")
	}
	return loader(b)
}
