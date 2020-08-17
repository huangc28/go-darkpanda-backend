package config

import (
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"

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

type TwilioConf struct {
	AccountId string `mapstructure:"account_id"`
	AuthToken string `mapstructure:"auth_token"`
	From      string `mapstructure:"from"`
}

type AppConf struct {
	Port       string      `mapstructure:"port"`
	JwtSecret  string      `mapstructure:"jwt_secret"`
	DBConf     *DBConf     `mapstructure:"db"`
	TwilioConf *TwilioConf `mapstructure:"twilio"`
}

var appConf AppConf

// getProjRootPath gets project root directory relative to `config/config.go`
func getProjRootPath() string {
	var (
		_, b, _, _ = runtime.Caller(0)
		basepath   = filepath.Dir(b)
	)

	log.Printf("project root dir %v", basepath)

	return filepath.Join(basepath, "..")
}

// reads config from .env.toml
func InitConfig() {
	viper.SetConfigName(".env")
	viper.SetConfigType("toml")

	// config path can be at project root directory
	cwd, err := os.Getwd()
	// retrieve executable path

	if err != nil {
		log.Fatalf("failed to get current working directory: %s", err.Error())
	}

	log.Infof("search .env in path... %s", cwd)

	viper.AddConfigPath(cwd)
	viper.AddConfigPath(".")
	viper.AddConfigPath(getProjRootPath())
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
