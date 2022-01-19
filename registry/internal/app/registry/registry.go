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
	// Я бы передавал логгер и конфигурацию через аргументы.
	// Что ты думаешь об этом? Есть ли разница в этих подходах?
	// Proc and cons этих подходов.
	config := conf.New()
	logger := logrus.New()
	// logger.SetFormatter(&logrus.JSONFormatter{})
	logEntry := logrus.NewEntry(logger)
	logrus.SetLevel(
		parseLogLevel(config.Logger.LogLevel),
	)

	// Плохая идея передавать так конфиг в клиент кафки и другие клиенты ниже.
	// Как ты думаешь, какие аргументы у меня могут быть, чтобы это говорить?
	brokerClient, err := broker.NewKafkaClient(config)
	if err != nil {
		panic(err)
	}

	// Закоментированный код. Ай-ай-ай)
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

func (app *App) Close() {
	app.OrdersService.Close()
	// Не логируется ошибка. Но метод всегда возвращает nil в ошибке.
	// Ай-ай-ай)
	app.BrokerClient.CloseReader()
	app.BrokerClient.CloseWriter()
	app.OrdersDAO.Close()
	app.OrderItemsDAO.Close()
	app.ProductPricesDAO.Close()
}
