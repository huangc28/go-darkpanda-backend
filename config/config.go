package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

const (
	ConfigName = ".env"
	ConfigType = "toml"
)

type DBConf struct {
	Host     string
	Port     uint
	User     string
	Password string
	Dbname   string
}

type AppConf struct {
	Port   string
	DBConf *DBConf `mapstructure:"db"`
}

var appConf AppConf

// reads config from .env.toml
func InitConfig() {
	viper.SetConfigName(".env")
	viper.SetConfigType("toml")

	// config path can be at project root directory
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("failed to get current working directory: %s", err.Error())
	}

	log.Printf("cwd %s", cwd)
	viper.AddConfigPath(cwd)
	viper.AddConfigPath(".")
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("configuration file not found: %s", err.Error())
		} else {
			log.Fatalf("error occurred when read in config file: %s", err.Error())
		}
	}

	if err = viper.Unmarshal(&appConf); err != nil {
		log.Fatalf("failed to unmarshal app config to struct %s", err.Error())
	}
}

func GetAppConf() *AppConf {
	return &appConf
}
