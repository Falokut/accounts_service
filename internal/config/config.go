package config

import (
	"sync"
	"time"

	"github.com/Falokut/accounts_service/internal/repository"
	"github.com/Falokut/accounts_service/pkg/jaeger"

	"github.com/Falokut/accounts_service/pkg/logging"
	"github.com/Falokut/accounts_service/pkg/metrics"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel        string `yaml:"log_level" env:"LOG_LEVEL"`
	HealthcheckPort string `yaml:"healthcheck_port" env:"HEALTHCHECK_PORT"`
	Listen          struct {
		Host                   string            `yaml:"host" env:"HOST"`
		Port                   string            `yaml:"port" env:"PORT"`
		Mode                   string            `yaml:"server_mode" env:"SERVER_MODE"` // support GRPC, REST, BOTH
		AllowedHeaders         []string          `yaml:"allowed_headers"`               // Need for REST API gateway, list of metadata headers
		AllowedOutgoingHeaders map[string]string `yaml:"allowed_outgoing_header"`       // Key - pretty header name, value - header name
	} `yaml:"listen"`

	PrometheusConfig struct {
		Name         string                      `yaml:"service_name" ENV:"PROMETHEUS_SERVICE_NAME"`
		ServerConfig metrics.MetricsServerConfig `yaml:"server_config"`
	} `yaml:"prometheus"`

	NonActivatedAccountTTL time.Duration       `yaml:"nonactivated_account_ttl"`
	SessionsTTL            time.Duration       `yaml:"sessions_ttl"` // The lifetime of an inactive session in the cache
	DBConfig               repository.DBConfig `yaml:"db_config"`
	JaegerConfig           jaeger.Config       `yaml:"jaeger"`

	RegistrationRepositoryConfig struct {
		Network  string `yaml:"network" env:"REGISTRATION_REPOSITORY_NETWORK"`
		Addr     string `yaml:"addr" env:"REGISTRATION_REPOSITORY_ADDRESS"`
		Password string `yaml:"password" env:"REGISTRATION_REPOSITORY_PASSWORD"`
		DB       int    `yaml:"db" env:"REGISTRATION_REPOSITORY_DATABASE"`
	} `yaml:"registration_repository"`
	SessionsCacheOptions struct {
		Network  string `yaml:"network" env:"SESSIONS_REPOSITORY_NETWORK"`
		Addr     string `yaml:"addr" env:"SESSIONS_REPOSITORY_ADDRESS"`
		Password string `yaml:"password" env:"SESSIONS_REPOSITORY_PASSWORD"`
		DB       int    `yaml:"db" env:"SESSIONS_REPOSITORY_DATABASE"`
	} `yaml:"sessions_repository"`

	NumRetriesForTerminateSessions     uint32        `yaml:"num_retries_for_terminate_sessions"`
	RetrySleepTimeForTerminateSessions time.Duration `yaml:"retry_sleep_time_for_terminate_sessions"`
	Crypto                             struct {
		BcryptCost int `yaml:"bcrypt_cost" enb:"BCRYPT_COST"`
	} `yaml:"crypto"`
	JWT struct {
		ChangePasswordToken struct {
			TTL    time.Duration `yaml:"ttl"`
			Secret string        `yaml:"secret" env:"CHANGE_PASSWORD_TOKEN_SECRET"`
		} `yaml:"change_password_token"`

		VerifyAccountToken struct {
			TTL    time.Duration `yaml:"ttl"`
			Secret string        `yaml:"secret" env:"VERIFY_ACCOUNT_TOKEN_SECRET"`
		} `yaml:"verify_account_token"`
	} `yaml:"JWT"`

	AccountEventsConfig struct {
		Brokers []string `yaml:"brokers"`
	} `yaml:"account_events"`
	TokensDeliveryConfig struct {
		Brokers []string `yaml:"brokers"`
	} `yaml:"tokens_delivery"`
}

var instance *Config
var once sync.Once

const configsPath = "configs/"

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		instance = &Config{}

		if err := cleanenv.ReadConfig(configsPath+"config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Fatal(help, " ", err)
		}
		if instance.NumRetriesForTerminateSessions <= 0 {
			instance.NumRetriesForTerminateSessions = 1
		}
	})

	return instance
}
