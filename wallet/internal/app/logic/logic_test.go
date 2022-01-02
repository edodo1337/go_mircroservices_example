package logic

import (
	"context"
	in "registry_service/internal/app/interfaces"
	"registry_service/internal/pkg/broker"
	"registry_service/internal/pkg/conf"
	"registry_service/internal/pkg/db"
	"testing"
	"time"

	"github.com/creasty/defaults"
	"github.com/sirupsen/logrus"
)

func TestMakeOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	config := &conf.Config{}
	if err := defaults.Set(config); err != nil {
		t.Error("err config set defaults", err)
	}

	logger := logrus.New()
	logEntry := logrus.NewEntry(logger)

	config.Kafka.Brokers = []string{"localhost:9093"}
	config.Kafka.NewOrdersTopic = "new_orders"
	config.Kafka.RejectedOrdersTopic = "rejected_orders"
	config.Kafka.GroupID = "registry"
	config.Kafka.ExternalClientsPort = 9092
	config.Kafka.InternalClientsPort = 9093

	orderDAO := db.NewInMemoryOrdersDAO()
	orderItemsDAO := db.NewInMemoryOrderItemsDAO()
	productPricesDAO := db.NewInMemoryProductPricesDAO()
	brokerClient := broker.NewInMemoryBrokerClient()

	service := NewOrdersService(
		orderDAO,
		orderItemsDAO,
		productPricesDAO,
		brokerClient,
		logEntry,
		config,
	)

	stop := make(chan struct{})
	go func(stop chan struct{}) {
		service.NewOrdersPipeline(ctx)
		stop <- struct{}{}
	}(stop)

	makeOrderData := &in.MakeOrderDTO{
		UserID: 1,
		OrderItems: []*in.MakeOrderItemDTO{
			{
				ProductID: 1,
				Count:     1,
			},
		},
	}

	if err := service.MakeOrder(ctx, *makeOrderData); err != nil {
		t.Error("make order error", err)
	}

	cancel()
	<-stop

	if len(orderDAO.OrdersKVStore) == 0 {
		t.Error("empty orders")
	}

	if len(orderItemsDAO.OrderItemsKVStore) == 0 {
		t.Error("empty order items")
	}
}

func TestMakeOrderWithKafka(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	config := &conf.Config{}
	if err := defaults.Set(config); err != nil {
		t.Error("err config set defaults", err)
	}

	logger := logrus.New()
	logEntry := logrus.NewEntry(logger)

	config.Kafka.Brokers = []string{"localhost:9093"}
	config.Kafka.NewOrdersTopic = "new_orders"
	config.Kafka.RejectedOrdersTopic = "rejected_orders"
	config.Kafka.GroupID = "registry"
	config.Kafka.ExternalClientsPort = 9092
	config.Kafka.InternalClientsPort = 9093

	orderDAO := db.NewInMemoryOrdersDAO()
	orderItemsDAO := db.NewInMemoryOrderItemsDAO()
	productPricesDAO := db.NewInMemoryProductPricesDAO()
	brokerClient, err := broker.NewKafkaClient(
		config.Kafka.Brokers,
		config.Kafka.NewOrdersTopic,
		config.Kafka.RejectedOrdersTopic,
		config.Kafka.GroupID,
	)
	if err != nil {
		t.Error("create kafka client err", err)
	}
	service := NewOrdersService(
		orderDAO,
		orderItemsDAO,
		productPricesDAO,
		brokerClient,
		logEntry,
		config,
	)

	stop := make(chan struct{})
	go func(stop chan struct{}) {
		service.NewOrdersPipeline(ctx)
		stop <- struct{}{}
	}(stop)

	makeOrderData := &in.MakeOrderDTO{
		UserID: 1,
		OrderItems: []*in.MakeOrderItemDTO{
			{
				ProductID: 1,
				Count:     1,
			},
		},
	}

	if err := service.MakeOrder(ctx, *makeOrderData); err != nil {
		t.Error("make order error", err)
	}

	// TODO sync problems
	time.Sleep(3 * time.Second)
	cancel()
	<-stop

	if len(orderDAO.OrdersKVStore) == 0 {
		t.Error("empty orders")
	}

	if len(orderItemsDAO.OrderItemsKVStore) == 0 {
		t.Error("empty order items")
	}
}

func TestMakeOrderWithKafkaAndConsumeRejected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	config := &conf.Config{}
	if err := defaults.Set(config); err != nil {
		t.Error("err config set defaults", err)
	}

	logger := logrus.New()
	logEntry := logrus.NewEntry(logger)

	config.Kafka.Brokers = []string{"localhost:9093"}
	config.Kafka.NewOrdersTopic = "new_orders"
	config.Kafka.RejectedOrdersTopic = "rejected_orders"
	config.Kafka.GroupID = "registry"
	config.Kafka.ExternalClientsPort = 9092
	config.Kafka.InternalClientsPort = 9093

	orderDAO := db.NewInMemoryOrdersDAO()
	orderItemsDAO := db.NewInMemoryOrderItemsDAO()
	productPricesDAO := db.NewInMemoryProductPricesDAO()
	brokerClient, err := broker.NewKafkaClient(
		config.Kafka.Brokers,
		config.Kafka.NewOrdersTopic,
		config.Kafka.RejectedOrdersTopic,
		config.Kafka.GroupID,
	)
	if err != nil {
		t.Error("create kafka client err", err)
	}

	service := NewOrdersService(
		orderDAO,
		orderItemsDAO,
		productPricesDAO,
		brokerClient,
		logEntry,
		config,
	)

	stop := make(chan struct{})
	go func(stop chan struct{}) {
		service.NewOrdersPipeline(ctx)
		stop <- struct{}{}
	}(stop)

	go func(stop chan struct{}) {
		go service.ConsumeRejectedOrderMsgLoop(ctx)
		stop <- struct{}{}
	}(stop)

	makeOrderData := &in.MakeOrderDTO{
		UserID: 1,
		OrderItems: []*in.MakeOrderItemDTO{
			{
				ProductID: 1,
				Count:     1,
			},
		},
	}

	if err := service.MakeOrder(ctx, *makeOrderData); err != nil {
		t.Error("make order error", err)
	}

	// TODO sync problems
	time.Sleep(8 * time.Second)
	cancel()
	<-stop
	<-stop

	if len(orderDAO.OrdersKVStore) == 0 {
		t.Error("empty orders")
	}

	if len(orderItemsDAO.OrderItemsKVStore) == 0 {
		t.Error("empty order items")
	}
}
