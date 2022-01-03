package registry

import (
	"context"
	in "registry_service/internal/app/interfaces"
	"registry_service/internal/app/logic"
	"registry_service/internal/pkg/broker"
	"registry_service/internal/pkg/conf"
	"registry_service/internal/pkg/db"

	"github.com/sirupsen/logrus"
)

type App struct {
	OrdersDAO        in.OrdersDAO
	OrderItemsDAO    in.OrderItemsDAO
	ProductPricesDAO in.ProductPricesDAO
	BrokerClient     in.BrokerClient

	OrdersService *logic.OrdersService

	Logger *logrus.Entry
	Config *conf.Config
}

func NewRegistryApp(ctx context.Context) *App {
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
		config.Kafka.SuccessTopic,
		config.Kafka.GroupID,
	)
	if err != nil {
		panic(err)
	}

	// brokerClient := broker.NewInMemoryBrokerClient()
	// ordersDAO := db.NewInMemoryOrdersDAO()
	// orderItemsDAO := db.NewInMemoryOrderItemsDAO()
	// productPricesDAO := db.NewInMemoryProductPricesDAO()

	ordersDAO := db.NewPostgresOrdersDAO(ctx, config)
	orderItemsDAO := db.NewPostgresOrderItemsDAO(ctx, config)
	productPricesDAO := db.NewPostgresProductPricesDAO(ctx, config)

	ordersService := logic.NewOrdersService(
		ordersDAO,
		orderItemsDAO,
		productPricesDAO,
		brokerClient,
		logEntry,
		config,
	)

	app := App{
		Logger:           logEntry,
		Config:           config,
		BrokerClient:     brokerClient,
		OrdersDAO:        ordersDAO,
		OrderItemsDAO:    orderItemsDAO,
		ProductPricesDAO: productPricesDAO,
		OrdersService:    ordersService,
	}

	return &app
}
