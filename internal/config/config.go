package config

import (
	"sync"
	"time"

	"github.com/Falokut/accounts_service/internal/repository"
	"github.com/Falokut/accounts_service/pkg/jaeger"

	"github.com/Falokut/accounts_service/pkg/metrics"
	logging "github.com/Falokut/online_cinema_ticket_office.loggerwrapper"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/redis/go-redis/v9"
)

type token struct {
	TTL    time.Duration `yaml:"TTL"`
	Secret string        `yaml:"secret"`
}

type redisOptions struct {
	Network  string `yaml:"network"`
	Addr     string `yaml:"addr"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

func (r redisOptions) ConvertToRedisOptions() *redis.Options {
	return &redis.Options{
		Network:  r.Network,
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	}
}

type Config struct {
	LogLevel            string `yaml:"log_level" env:"LOG_LEVEL"`
	ProfilesServiceAddr string `yaml:"profiles_service_addr" env:"PROFILES_SERVICE_ADDR"`
	HealthcheckPort     string `yaml:"healthcheck_port" env:"HEALTHCHECK_PORT"`
	Listen              struct {
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
	EmailKafka             KafkaConfig         `yaml:"email_kafka_config"`
	JaegerConfig           jaeger.Config       `yaml:"jaeger"`

	RegistrationCacheOptions           redisOptions  `yaml:"redis_registration_options"`
	SessionCacheOptions                redisOptions  `yaml:"session_cache_options"`
	AccountSessionsCacheOptions        redisOptions  `yaml:"account_sessions_cache_options"`
	NumRetriesForTerminateSessions     int32         `yaml:"num_retries_for_terminate_sessions"`
	RetrySleepTimeForTerminateSessions time.Duration `yaml:"retry_sleep_time_for_terminate_sessions"`
	Crypto                             struct {
		BcryptCost int `yaml:"bcrypt_cost" enb:"BCRYPT_COST"`
	} `yaml:"crypto"`
	JWT struct {
		ChangePasswordToken token `yaml:"change_password_token"`
		VerifyAccountToken  token `yaml:"verify_account_token"`
	} `yaml:"JWT"`
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

		if err := cleanenv.ReadConfig(configsPath+"secrets.env.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Fatal(help, " ", err)
		}
		if instance.NumRetriesForTerminateSessions <= 0 {
			instance.NumRetriesForTerminateSessions = 1
		}
	})

	return instance
}
