package config

import (
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type AppConf struct {
	Port      string `mapstructure:"PORT"`
	JwtSecret string `mapstructure:"JWT_SECRET"`

	PGHost     string `mapstructure:"PG_HOST"`
	PGPort     uint   `mapstructure:"PG_PORT"`
	PGUser     string `mapstructure:"PG_USER"`
	PGPassword string `mapstructure:"PG_PASSWORD"`
	PGDbname   string `mapstructure:"PG_DBNAME"`

	TestPGHost     string `mapstructure:"TEST_PG_HOST"`
	TestPGPort     uint   `mapstructure:"TEST_PG_PORT"`
	TestPGUser     string `mapstructure:"TEST_PG_USER"`
	TestPGPassword string `mapstructure:"TEST_PG_PASSWORD"`
	TestPGDbname   string `mapstructure:"TEST_PG_DBNAME"`

	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDb       uint   `mapstructure:"REDIS_DB"`

	TwilioAccountID string `mapstructure:"TWILIO_ACCOUNT_ID"`
	TwilioAuthToken string `mapstructure:"TWILIO_AUTH_TOKEN"`
	TwilioFrom      string `mapstructure:"TWILIO_FROM"`

	GcpProjectID string `mapstructure:"GCP_PROJECT_ID"`

	GcsGoogleServiceAccountName string `mapstructure:"GCS_GOOGLE_SERVICE_ACCOUNT_NAME"`
	GcsBucketName               string `mapstructure:"GCS_BUCKET_NAME"`

	PubnubPublishKey   string `mapstructure:"PUBNUB_PUBLISH_KEY"`
	PubnubSubscribeKey string `mapstructure:"PUBNUB_SUBSCRIBE_KEY"`
	PubnubSecretKey    string `mapstructure:"PUBNUB_SECRET_KEY"`

	FirestoreCredentialFile string `mapstructure:"FIRESTORE_CREDENTIAL_FILE"`

	TappayEndpoint   string `mapstructure:"TAPPAY_ENDPOINT"`
	TappayPartnerKey string `mapstructure:"TAPPAY_PARTNER_KEY"`
	TappayMerchantId string `mapstructure:"TAPPAY_MERCHANT_ID"`
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

// List of environments.
// Note: not in use at the moment
type Env string

var (
	Production  Env = "production"
	Staging     Env = "staging"
	Development Env = "development"
	Test        Env = "test"
)

// reads config from .env
func InitConfig() {
	viper.SetConfigType("env")
	viper.SetConfigName(".app")

	// Config path can be at project root directory
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

func GetAppConf() *AppConf {
	return &appConf
}
