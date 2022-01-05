package wallet

import (
	"context"
	in "wallet_service/internal/app/interfaces"
	"wallet_service/internal/app/logic"
	"wallet_service/internal/pkg/broker"
	"wallet_service/internal/pkg/conf"
	"wallet_service/internal/pkg/db"

	"github.com/sirupsen/logrus"
)

type App struct {
	WalletsDAO            in.WalletsDAO
	WalletTransactionsDAO in.WalletTransactionsDAO
	BrokerClient          in.BrokerClient

	PaymentService *logic.PaymentService

	Logger *logrus.Entry
	Config *conf.Config
}

func NewWalletApp(ctx context.Context) *App {
	config := conf.New()
	logger := logrus.New()
	// logger.SetFormatter(&logrus.JSONFormatter{})
	logEntry := logrus.NewEntry(logger)
	logrus.SetLevel(
		parseLogLevel(config.Logger.LogLevel),
	)

	brokerClient, err := broker.NewKafkaClient(config)
	if err != nil {
		panic(err)
	}

	walletsDAO := db.NewPostgresWalletsDAO(ctx, config)
	walletTransDAO := db.NewPostgresWalletTransDAO(ctx, config)

	paymentService := logic.NewPaymentService(
		walletsDAO,
		walletTransDAO,
		brokerClient,
		logEntry,
		config,
	)

	app := App{
		Logger:                logEntry,
		Config:                config,
		BrokerClient:          brokerClient,
		WalletsDAO:            walletsDAO,
		WalletTransactionsDAO: walletTransDAO,
		PaymentService:        paymentService,
	}

	return &app
}

func (app *App) Close() {
	app.PaymentService.Close()
	app.BrokerClient.CloseReader()
	app.BrokerClient.CloseWriter()
	app.WalletsDAO.Close()
	app.WalletTransactionsDAO.Close()
}
