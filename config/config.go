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

type TestDBConf struct {
	Host     string
	Port     uint
	User     string
	Password string
	Dbname   string
}

type RedisConf struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type TwilioConf struct {
	AccountId string `mapstructure:"account_id"`
	AuthToken string `mapstructure:"auth_token"`
	From      string `mapstructure:"from"`
}

// GCSCredentials for manipulating google cloud storage.
type GCSCredentials struct {
	GoogleServiceAccountName string `mapstructure:"google_service_account_name"`
	BucketName               string `mapstructure:"bucket_name"`
}

// @deprecated
type PubnubCredentials struct {
	PublishKey   string `mapstructure:"publish_key"`
	SubscribeKey string `mapstructure:"subscribe_key"`
	SecretKey    string `mapstructure:"secret_key"`
}

type FirestoreCredentials struct {
	CredentialFile string `mapstructure:"credential_file"`
}

type AppConf struct {
	Port           string          `mapstructure:"port"`
	JwtSecret      string          `mapstructure:"jwt_secret"`
	DBConf         *DBConf         `mapstructure:"db"`
	TestDBConf     *TestDBConf     `mapstructure:"test_db"`
	RedisConf      *RedisConf      `mapstructure:"redis"`
	TwilioConf     *TwilioConf     `mapstructure:"twilio"`
	GCSCredentials *GCSCredentials `mapstructure:"gcs"`

	// @deprecated
	PubnubCredentials *PubnubCredentials `mapstructure:"pubnub"`

	Firestore *FirestoreCredentials `mapstructure:"firestore"`
}

var appConf AppConf

// GetProjRootPath gets project root directory relative to `config/config.go`
func GetProjRootPath() string {
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
	viper.AddConfigPath(GetProjRootPath())
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

func GetDBConf() *DBConf {
	return GetAppConf().DBConf
}

func GetTestDBConf() *TestDBConf {
	return GetAppConf().TestDBConf
}

func GetAppConf() *AppConf {
	return &appConf
}
