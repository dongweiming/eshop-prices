package config

import (
	"log"
	"runtime"
	"path"
	"sync"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	DB      database `toml:"db"`
}

var once sync.Once

type config struct {
	instance tomlConfig
}
var conf *config

type database struct {
	Host   string
	Port   int
	User   string
	Passwd string
	Database string
}

func ReadConfig() tomlConfig {
	once.Do(func() {
		conf = new(config)
		_, filename, _, _ := runtime.Caller(0)
		f := filepath.Join(path.Dir(filename), "..", "config.toml")
		_, err := toml.DecodeFile(f, &(conf.instance))
		if err != nil {
			log.Fatal(err)
		}
	})
	return conf.instance
}
