package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/Falokut/accounts_service/internal/config"
	"github.com/Falokut/accounts_service/internal/events"
	"github.com/Falokut/accounts_service/internal/handler"
	"github.com/Falokut/accounts_service/internal/repository/postgresrepository"
	"github.com/Falokut/accounts_service/internal/repository/redisrepository"
	"github.com/Falokut/accounts_service/internal/service"
	accounts_service "github.com/Falokut/accounts_service/pkg/accounts_service/v1/protos"
	jaegerTracer "github.com/Falokut/accounts_service/pkg/jaeger"
	"github.com/Falokut/accounts_service/pkg/logging"
	"github.com/Falokut/accounts_service/pkg/metrics"
	server "github.com/Falokut/grpc_rest_server"
	"github.com/Falokut/healthcheck"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/opentracing/opentracing-go"
)

func main() {
	logging.NewEntry(logging.ConsoleOutput)
	logger := logging.GetLogger()
	cfg := config.GetConfig()

	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Logger.SetLevel(logLevel)

	tracer, closer, err := jaegerTracer.InitJaeger(cfg.JaegerConfig)
	if err != nil {
		logger.Errorf("Shutting down, error while creating tracer %v", err)
		return
	}
	logger.Info("Jaeger connected")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	logger.Info("Metrics initializing")
	metric, err := metrics.CreateMetrics(cfg.PrometheusConfig.Name)
	if err != nil {
		logger.Errorf("Shutting down, error while creating metrics %v", err)
		return
	}

	shutdown := make(chan error, 1)
	go func() {
		logger.Info("Metrics server running")
		if err = metrics.RunMetricServer(cfg.PrometheusConfig.ServerConfig); err != nil {
			logger.Errorf("Shutting down, error while running metrics server %v", err)
			shutdown <- err
			return
		}
	}()

	logger.Info("Registration cache initializing")
	registrationRepository, err := redisrepository.NewRedisRegistrationRepository(
		&redis.Options{
			Network:  cfg.RegistrationRepositoryConfig.Network,
			Addr:     cfg.RegistrationRepositoryConfig.Addr,
			Password: cfg.RegistrationRepositoryConfig.Password,
			DB:       cfg.RegistrationRepositoryConfig.DB,
		}, logger.Logger, metric)

	if err != nil {
		logger.Errorf("Shutting down, connection to the redis registration repository is not established: %s",
			err.Error())
		return
	}
	defer registrationRepository.Shutdown()

	logger.Info("Sessions cache initializing")
	sessionsRepository, err := redisrepository.NewSessionsRepository(
		&redis.Options{
			Network:  cfg.SessionsCacheOptions.Network,
			Addr:     cfg.SessionsCacheOptions.Addr,
			Password: cfg.SessionsCacheOptions.Password,
			DB:       cfg.SessionsCacheOptions.DB,
		},
		logger.Logger, metric)
	if err != nil {
		logger.Errorf("Shutting down, connection to the redis sessions repository is not established: %s",
			err.Error())
		return
	}
	defer sessionsRepository.Shutdown()

	logger.Info("Database initializing")
	database, err := postgresrepository.NewPostgreDB(&cfg.DBConfig)
	if err != nil {
		logger.Errorf("Shutting down, connection to the database is not established: %s", err.Error())
		return
	}

	logger.Info("Repository initializing")
	repo := postgresrepository.NewAccountsRepository(database, logger.Logger)
	defer repo.Shutdown()

	accountsEventsMQ := events.NewAccountsEvents(events.KafkaConfig{
		Brokers: cfg.AccountEventsConfig.Brokers,
	}, logger.Logger)
	defer accountsEventsMQ.Shutdown()

	tokenDeliveryMQ := events.NewTokensDeliveryMQ(events.KafkaConfig{
		Brokers: cfg.TokensDeliveryConfig.Brokers,
	}, logger.Logger)
	defer accountsEventsMQ.Shutdown()

	go func() {
		logger.Info("Healthcheck initializing")
		healthcheckManager := healthcheck.NewHealthManager(logger.Logger,
			[]healthcheck.HealthcheckResource{repo}, cfg.HealthcheckPort, nil)
		if err := healthcheckManager.RunHealthcheckEndpoint(); err != nil {
			logger.Errorf("Shutting down, error while running healthcheck endpoint %s", err.Error())
			shutdown <- err
			return
		}
	}()

	logger.Info("Service initializing")
	s := service.NewAccountsService(repo,
		logger.Logger, registrationRepository, sessionsRepository, accountsEventsMQ, tokenDeliveryMQ,
		getAccountServiceConfig(cfg))

	h := handler.NewAccountsServiceHandler(logger.Logger, s)

	logger.Info("Server initializing")
	serv := server.NewServer(logger.Logger, h)
	go func() {
		if err := serv.Run(getListenServerConfig(cfg), metric, nil, nil); err != nil {
			logger.Errorf("Shutting down, error while running server %s", err.Error())
			shutdown <- err
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGTERM)

	select {
	case <-quit:
		break
	case <-shutdown:
		break
	}

	serv.Shutdown()
}

func getListenServerConfig(cfg *config.Config) server.Config {
	return server.Config{
		Mode:                   cfg.Listen.Mode,
		Host:                   cfg.Listen.Host,
		Port:                   cfg.Listen.Port,
		AllowedHeaders:         cfg.Listen.AllowedHeaders,
		AllowedOutgoingHeaders: cfg.Listen.AllowedOutgoingHeaders,
		ServiceDesc:            &accounts_service.AccountsServiceV1_ServiceDesc,
		RegisterRestHandlerServer: func(_ context.Context, mux *runtime.ServeMux, service any) error {
			serv, ok := service.(accounts_service.AccountsServiceV1Server)
			if !ok {
				return errors.New("can't convert")
			}
			return accounts_service.RegisterAccountsServiceV1HandlerServer(context.Background(),
				mux, serv)
		},
	}
}

func getAccountServiceConfig(cfg *config.Config) *service.AccountsServiceConfig {
	return &service.AccountsServiceConfig{
		ChangePasswordTokenTTL:             cfg.JWT.ChangePasswordToken.TTL,
		ChangePasswordTokenSecret:          cfg.JWT.ChangePasswordToken.Secret,
		VerifyAccountTokenTTL:              cfg.NonActivatedAccountTTL,
		VerifyAccountTokenSecret:           cfg.JWT.VerifyAccountToken.Secret,
		NumRetriesForTerminateSessions:     cfg.NumRetriesForTerminateSessions,
		RetrySleepTimeForTerminateSessions: cfg.RetrySleepTimeForTerminateSessions,
		NonActivatedAccountTTL:             cfg.NonActivatedAccountTTL,
		BcryptCost:                         cfg.Crypto.BcryptCost,
		SessionTTL:                         cfg.SessionsTTL,
	}
}
