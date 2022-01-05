package logic

import (
	in "registry_service/internal/app/interfaces"
	"registry_service/internal/pkg/conf"
	"time"

	"github.com/sirupsen/logrus"
)

type OrdersService struct {
	ordersDAO          in.OrdersDAO
	orderItemsDAO      in.OrderItemsDAO
	productPricesDAO   in.ProductPricesDAO
	brokerClient       in.BrokerClient
	newOrdersPipe      chan *in.NewOrderDTO
	rejectedOrdersPipe chan *in.OrderRejectedMsg
	successOrdersPipe  chan *in.OrderSuccessMsg
	sendMsgTimeout     time.Duration
	consumeLoopTick    time.Duration
	logger             *logrus.Entry
}

func NewOrdersService(
	ordersDAO in.OrdersDAO,
	orderItemsDAO in.OrderItemsDAO,
	productPricesDAO in.ProductPricesDAO,
	brokerClient in.BrokerClient,
	logger *logrus.Entry,
	config *conf.Config,
) *OrdersService {
	newOrdersPipe := make(chan *in.NewOrderDTO, config.Server.NewOrdersPipeCapacity)
	rejectedOrdersPipe := make(chan *in.OrderRejectedMsg, config.Server.NewOrdersPipeCapacity)
	successOrdersPipe := make(chan *in.OrderSuccessMsg, config.Server.NewOrdersPipeCapacity)

	return &OrdersService{
		ordersDAO:          ordersDAO,
		orderItemsDAO:      orderItemsDAO,
		productPricesDAO:   productPricesDAO,
		brokerClient:       brokerClient,
		newOrdersPipe:      newOrdersPipe,
		rejectedOrdersPipe: rejectedOrdersPipe,
		successOrdersPipe:  successOrdersPipe,
		sendMsgTimeout:     time.Duration(config.Kafka.SendMsgTimeout) * time.Second,
		consumeLoopTick:    time.Duration(config.Kafka.ConsumeLoopTick) * time.Millisecond,
		logger:             logger,
	}
}

func (s *OrdersService) Close() {
	close(s.newOrdersPipe)
	close(s.rejectedOrdersPipe)
	close(s.successOrdersPipe)
}
