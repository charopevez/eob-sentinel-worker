package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/charopevez/eob-wayfinder-worker/internal/client/sentinel_service"
	"github.com/charopevez/eob-wayfinder-worker/internal/config"
	"github.com/charopevez/eob-wayfinder-worker/internal/handlers/sentinel"
	"github.com/charopevez/eob-wayfinder-worker/pkg/cache/freecache"
	"github.com/charopevez/eob-wayfinder-worker/pkg/handlers/metric"
	"github.com/charopevez/eob-wayfinder-worker/pkg/jwt"
	"github.com/charopevez/eob-wayfinder-worker/pkg/logging"
	"github.com/charopevez/eob-wayfinder-worker/pkg/shutdown"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Println("logger initialized")

	logger.Println("config initializing")
	cfg := config.GetConfig()

	logger.Println("router initializing")
	router := httprouter.New()

	logger.Println("cache initializing")
	refreshTokenCache := freecache.NewCacheRepo(104857600) // 100MB

	logger.Println("helpers initializing")
	jwtHelper := jwt.NewHelper(refreshTokenCache, logger)

	logger.Println("create and register handlers")

	metricHandler := metric.Handler{Logger: logger}
	metricHandler.Register(router)

	sentinelService := sentinel_service.NewService(cfg.SentinelService.URL, "/users", logger)
	authHandler := sentinel.Handler{JWTHelper: jwtHelper, SentinelService: sentinelService, Logger: logger}
	authHandler.Register(router)

	logger.Println("start application")
	start(router, logger, cfg)
}

func start(router *httprouter.Router, logger logging.Logger, cfg *config.Config) {
	var server *http.Server
	var listener net.Listener

	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		socketPath := path.Join(appDir, "app.sock")
		logger.Infof("socket path: %s", socketPath)

		logger.Info("create and listen unix socket")
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		logger.Infof("bind application to host: %s and port: %s", cfg.Listen.BindIP, cfg.Listen.Port)

		var err error

		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
		if err != nil {
			logger.Fatal(err)
		}
	}

	server = &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go shutdown.Graceful([]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM},
		server)

	logger.Println("application initialized and started")

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("server shutdown")
		default:
			logger.Fatal(err)
		}
	}
}
