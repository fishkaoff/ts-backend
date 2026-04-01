package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	excel_adapter "github.com/fishkaoff/ts-backend/internal/adapters/excel"
	"github.com/fishkaoff/ts-backend/internal/config"
	jwtclient "github.com/fishkaoff/ts-backend/internal/domain/lib/jwt"
	"github.com/fishkaoff/ts-backend/internal/services/auth"
	"github.com/fishkaoff/ts-backend/internal/services/carts"
	"github.com/fishkaoff/ts-backend/internal/services/products"
	mongostorage "github.com/fishkaoff/ts-backend/internal/storage/mongo"
	httpserver "github.com/fishkaoff/ts-backend/internal/transport/http"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mongo := mongostorage.New(cfg.MongoConfig)

	log.Info("connect to database")
	client, err := mongo.Connect()
	if err != nil {
		panic(err)
	}

	log.Info("init services")
	authSvc := auth.New(log, cfg.JWTConfig, mongo)
	excelAdapter := excel_adapter.New()
	productsSvc := products.New(log, mongo, excelAdapter)
	jwtService := jwtclient.New([]byte(cfg.JWTConfig.SecretKey))
	cartsService := carts.New(log, mongo)

	log.Info("start http server")
	httpServer := httpserver.New(
		cfg.RESTConfig,
		log, jwtService,
		authSvc,
		productsSvc,
		cartsService,
	)
	go func() {
		err := httpServer.Start()
		if err != nil {
			panic(err)
		}
	}()

	// gracefully shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Info("stop app")

	log.Info("disconnect database")
	if err := client.Disconnect(ctx); err != nil {
		panic(err)
	}

}
