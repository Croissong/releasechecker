package config

import (
	"github.com/croissong/releasechecker/pkg/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"os"
)

var CfgFile string
var Config configuration

func InitConfig() {
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
		viper.SetConfigType("yaml")
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Logger.Error(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".releasechecker")
	}

	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err != nil {
		log.Logger.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&Config)
	log.Logger.Infof("Using config: %+v", Config)
	if err != nil {
		log.Logger.Fatalf("unable to decode into struct, %v", err)
	}
}

type configuration struct {
	Debug       bool
	InitSources bool
	Entries     []entry
}

type entry struct {
	Name     string
	Source   map[string]interface{}
	Provider map[string]interface{}
	Hooks    []map[string]interface{}
}
