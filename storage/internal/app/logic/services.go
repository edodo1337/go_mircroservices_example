package logic

import (
	in "storage_service/internal/app/interfaces"
	"storage_service/internal/pkg/conf"
	"time"

	"github.com/sirupsen/logrus"
)

type StorageService struct {
	storageItemsDAO        in.StorageItemsDAO
	storageTransactionsDAO in.StorageTransactionsDAO
	brokerClient           in.BrokerClient
	transactionsPipe       chan *in.Transaction

	sendMsgTimeout  time.Duration
	consumeLoopTick time.Duration
	logger          *logrus.Entry
}

func NewStorageService(
	storageItemsDAO in.StorageItemsDAO,
	storageTransactionsDAO in.StorageTransactionsDAO,
	brokerClient in.BrokerClient,
	logger *logrus.Entry,
	config *conf.Config,
) *StorageService {
	transactionsPipe := make(chan *in.Transaction, config.Server.TransactionsPipeCapacity)

	return &StorageService{
		storageItemsDAO:        storageItemsDAO,
		storageTransactionsDAO: storageTransactionsDAO,
		brokerClient:           brokerClient,
		transactionsPipe:       transactionsPipe,
		sendMsgTimeout:         time.Duration(config.Kafka.SendMsgTimeout) * time.Second,
		consumeLoopTick:        time.Duration(config.Kafka.ConsumeLoopTick) * time.Millisecond,
		logger:                 logger,
	}
}

func (s *StorageService) Close() {
	close(s.transactionsPipe)
}
