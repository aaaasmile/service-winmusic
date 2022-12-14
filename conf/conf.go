package conf

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ServiceURL     string
	RootURLPattern string
	DebugVerbose   bool
	VuetifyLibName string
	VueLibName     string
	DBPath         string
	MusicDir       string
	Player         Player
}

type Player struct {
	Path   string
	Params string
}

var Current = &Config{}

func ReadConfig(configfile string) *Config {
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := toml.DecodeFile(configfile, &Current); err != nil {
		log.Fatal(err)
	}

	return Current
}
