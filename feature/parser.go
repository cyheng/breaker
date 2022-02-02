package feature

import (
	"errors"
	"github.com/go-ini/ini"
	"os"
	"path/filepath"
	"strings"
)

type Loader func([]byte, interface{}) error

var Loaders = make(map[string]Loader)

func init() {
	RegisterLoader("ini", func(b []byte, conf interface{}) error {
		f, err := ini.LoadSources(ini.LoadOptions{
			Insensitive:         false,
			InsensitiveSections: false,
			InsensitiveKeys:     false,
			IgnoreInlineComment: true,
			AllowBooleanKeys:    true,
		}, b)
		if err != nil {
			return err
		}
		return f.MapTo(conf)
	})
}

func RegisterLoader(ext string, c Loader) {
	Loaders[ext] = c
}

func LoadFromFile(cfgFile string, conf interface{}) error {
	ext := strings.ToLower(filepath.Ext(cfgFile))
	ext = strings.TrimLeft(ext, ".")
	b, err := os.ReadFile(cfgFile)
	if err != nil {
		return err
	}
	// only support ini format
	loader, ok := Loaders[ext]
	if !ok {
		return errors.New("config extent not define")
	}

	return loader(b, conf)
}
