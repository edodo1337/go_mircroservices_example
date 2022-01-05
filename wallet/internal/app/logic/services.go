package logic

import (
	"time"
	in "wallet_service/internal/app/interfaces"
	"wallet_service/internal/pkg/conf"

	"github.com/sirupsen/logrus"
)

type PaymentService struct {
	walletsDAO             in.WalletsDAO
	walletsTransactionsDAO in.WalletTransactionsDAO
	brokerClient           in.BrokerClient
	transactionsPipe       chan *in.Transaction

	sendMsgTimeout  time.Duration
	consumeLoopTick time.Duration
	logger          *logrus.Entry
}

func NewPaymentService(
	walletsDAO in.WalletsDAO,
	walletsTransactionsDAO in.WalletTransactionsDAO,
	brokerClient in.BrokerClient,
	logger *logrus.Entry,
	config *conf.Config,
) *PaymentService {
	transactionsPipe := make(chan *in.Transaction, config.Server.TransactionsPipeCapacity)

	return &PaymentService{
		walletsDAO:             walletsDAO,
		walletsTransactionsDAO: walletsTransactionsDAO,
		brokerClient:           brokerClient,
		transactionsPipe:       transactionsPipe,
		sendMsgTimeout:         time.Duration(config.Kafka.SendMsgTimeout) * time.Second,
		consumeLoopTick:        time.Duration(config.Kafka.ConsumeLoopTick) * time.Millisecond,
		logger:                 logger,
	}
}

func (s *PaymentService) Close() {
	close(s.transactionsPipe)
}
