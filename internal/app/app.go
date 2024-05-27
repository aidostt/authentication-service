package app

import (
	"authentication-service/internal/config"
	"authentication-service/internal/delivery"
	"authentication-service/internal/repository"
	"authentication-service/internal/server"
	"authentication-service/internal/service"
	"authentication-service/pkg/database/mongodb"
	"authentication-service/pkg/hash"
	"authentication-service/pkg/logger"
	authManager "authentication-service/pkg/manager"
	"context"
	"errors"
	"fmt"
	"net"
	netHttp "net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(configPath, envPath string) {
	cfg, err := config.Init(configPath, envPath)
	if err != nil {
		logger.Error(err)

		return
	}

	// Dependencies
	mongoClient, err := mongodb.NewClient(cfg.Mongo.URI, cfg.Mongo.User, cfg.Mongo.Password)
	if err != nil {
		logger.Error(err)

		return
	}
	db := mongoClient.Database(cfg.Mongo.Name)

	hasher := hash.NewHasher(cfg.Auth.PasswordCost)

	tokenManager, err := authManager.NewManager(cfg.Auth.JWT.SigningKey)
	if err != nil {
		logger.Error(err)

		return
	}

	repos := repository.NewModels(db)
	services := service.NewServices(service.Dependencies{
		Repos:              repos,
		Hasher:             hasher,
		TokenManager:       tokenManager,
		AccessTokenTTL:     cfg.Auth.JWT.AccessTokenTTL,
		RefreshTokenTTL:    cfg.Auth.RefreshTokenTTL,
		ActivationTokenTTL: cfg.Auth.ActivationTokenTTL,
		ActivationCodeTTL:  cfg.Auth.ActivationCodeTTL,
		Environment:        cfg.Environment,
		Domain:             cfg.GRPC.Host,
	})
	handlers := delivery.NewHandler(services)

	// gRPC Server
	srv := server.NewServer()
	srv.RegisterServers(handlers)

	l, err := net.Listen("tcp", fmt.Sprintf("%v:%v", cfg.GRPC.Host, cfg.GRPC.Port))
	if err != nil {
		logger.Errorf("error occurred while getting listener for the server: %s\n", err.Error())
		return
	}
	go func() {
		if err := srv.Run(l); err != nil && !errors.Is(err, netHttp.ErrServerClosed) {
			logger.Errorf("error occurred while running grpc server: %s\n", err.Error())
		}
	}()

	logger.Info("Server started at: " + cfg.GRPC.Host + ":" + cfg.GRPC.Port)

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	srv.Stop()
	logger.Info("Stopping server at: " + cfg.GRPC.Host + ":" + cfg.GRPC.Port)
	if err := mongoClient.Disconnect(context.Background()); err != nil {
		logger.Error(err.Error())
	}

}
