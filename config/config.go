package config

import (
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type AppConf struct {
	Port                string `mapstructure:"PORT"`
	JwtSecret           string `mapstructure:"JWT_SECRET"`
	ServiceQrCodeSecret string `mapstructure:"SERVICE_QRCODE_SECRET"`
	AppTimeZone         string `mapstructure:"APP_TIME_ZONE"`

	PGHost     string `mapstructure:"PG_HOST"`
	PGPort     uint   `mapstructure:"PG_PORT"`
	PGUser     string `mapstructure:"PG_USER"`
	PGPassword string `mapstructure:"PG_PASSWORD"`
	PGDbname   string `mapstructure:"PG_DBNAME"`
	PGTimeZone string `mapstructure:"PG_TIMEZONE"`

	TestPGHost     string `mapstructure:"TEST_PG_HOST"`
	TestPGPort     uint   `mapstructure:"TEST_PG_PORT"`
	TestPGUser     string `mapstructure:"TEST_PG_USER"`
	TestPGPassword string `mapstructure:"TEST_PG_PASSWORD"`
	TestPGDbname   string `mapstructure:"TEST_PG_DBNAME"`
	TestPGTimeZone string `mapstructure:"TEST_PG_TIMEZONE"`

	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDb       uint   `mapstructure:"REDIS_DB"`

	TwilioAccountID    string `mapstructure:"TWILIO_ACCOUNT_ID"`
	TwilioAuthToken    string `mapstructure:"TWILIO_AUTH_TOKEN"`
	TwilioFrom         string `mapstructure:"TWILIO_FROM"`
	TwilioDevAccountID string `mapstructure:"TWILIO_DEV_ACCOUNT_ID"`
	TwilioDevAuthToken string `mapstructure:"TWILIO_DEV_AUTH_TOKEN"`
	TwilioDevFrom      string `mapstructure:"TWILIO_DEV_FROM"`

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

	ErrorLogPath string `mapstructure:"ERROR_LOG_PATH"`
	InfoLogPath  string `mapstructure:"INFO_LOG_PATH"`

	AppcenterAppSecret                 string `mapstructure:"APPCENTER_APP_SECRET"`
	AppcenterPublicDistributionGroupId string `mapstructure:"APPCENTER_PUBLIC_DISTRIBUTION_GROUP_ID"`

	// Note: We are hardcoding app currency here. Since this app only operates in Taiwan for now.
	Currency string `mapstructure:"CURRENCY"`

	// DEV usernames, login via DEV usernames receive 1234 for otp code.
	DevUsernames []string
}

func (ac *AppConf) IsDevUser(username string) bool {
	for _, devUser := range ac.DevUsernames {
		if username == devUser {
			return true
		}
	}

	return false
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

	// Setup DEV usernames
	appConf.DevUsernames = []string{
		"tester",
		"arthur",
	}
}

func GetAppConf() *AppConf {
	return &appConf
}
