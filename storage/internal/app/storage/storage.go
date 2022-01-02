package storage

import (
	"context"
	in "storage_service/internal/app/interfaces"
	"storage_service/internal/app/logic"
	"storage_service/internal/pkg/broker"
	"storage_service/internal/pkg/conf"
	"storage_service/internal/pkg/db"

	"github.com/sirupsen/logrus"
)

type App struct {
	StorageItemsDAO        in.StorageItemsDAO
	StorageTransactionsDAO in.StorageTransactionsDAO
	BrokerClient           in.BrokerClient

	StorageService *logic.StorageService

	Logger *logrus.Entry
	Config *conf.Config
}

func NewStorageApp(ctx context.Context) *App {
	config := conf.New()
	logger := logrus.New()
	// logger.SetFormatter(&logrus.JSONFormatter{})
	logEntry := logrus.NewEntry(logger)
	logrus.SetLevel(
		parseLogLevel(config.Logger.LogLevel),
	)

	brokerClient, err := broker.NewKafkaClient(
		config.Kafka.Brokers,
		config.Kafka.NewOrdersTopic,
		config.Kafka.RejectedOrdersTopic,
		config.Kafka.GroupID,
	)
	if err != nil {
		panic(err)
	}

	storageItemsDAO := db.NewPostgresStorageItemsDAO(ctx, config)
	storageTransactionsDAO := db.NewPostgresStorageTransDAO(ctx, config)

	storageService := logic.NewStorageService(
		storageItemsDAO,
		storageTransactionsDAO,
		brokerClient,
		logEntry,
		config,
	)

	app := App{
		Logger:                 logEntry,
		Config:                 config,
		BrokerClient:           brokerClient,
		StorageItemsDAO:        storageItemsDAO,
		StorageTransactionsDAO: storageTransactionsDAO,
		StorageService:         storageService,
	}

	return &app
}
