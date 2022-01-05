package broker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"
	in "wallet_service/internal/app/interfaces"
	"wallet_service/internal/pkg/conf"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	NewOrdersReader      *kafka.Reader
	RejectedOrdersReader *kafka.Reader

	WriterFails   *kafka.Writer
	WriterSuccess *kafka.Writer

	brokers          []string
	healthCheckTopic string
}

func NewKafkaClient(config *conf.Config) (*KafkaClient, error) {
	c := config.Kafka

	if len(c.Brokers) == 0 ||
		c.Brokers[0] == "" ||
		c.NewOrdersTopic == "" ||
		c.RejectedOrdersTopic == "" ||
		c.SuccessTopic == "" ||
		c.GroupID == "" {
		return nil, in.ErrInvalidBrokerConnParams
	}

	client := KafkaClient{
		brokers:          c.Brokers,
		healthCheckTopic: "healthcheck",
	}

	client.NewOrdersReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.Brokers,
		Topic:    c.NewOrdersTopic,
		GroupID:  c.GroupID,
		MinBytes: 10e1,
		MaxBytes: 10e6,
		MaxWait:  time.Duration(c.MaxWait) * time.Millisecond,
	})

	client.RejectedOrdersReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.Brokers,
		Topic:    c.RejectedOrdersTopic,
		GroupID:  c.GroupID,
		MinBytes: 10e1,
		MaxBytes: 10e6,
		MaxWait:  time.Duration(c.MaxWait) * time.Millisecond,
	})

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	client.WriterFails = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      c.Brokers,
		Topic:        c.RejectedOrdersTopic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		RequiredAcks: -1,
	})

	client.WriterSuccess = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      c.Brokers,
		Topic:        c.SuccessTopic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		RequiredAcks: -1,
	})

	return &client, nil
}

func (c *KafkaClient) SendOrderRejectedMsg(ctx context.Context, msg *in.OrderRejectedMsg) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	data := kafka.Message{
		Value: value,
	}

	err = c.WriterFails.WriteMessages(ctx, data)

	return err
}

func (c *KafkaClient) SendPurchaseSuccess(ctx context.Context, msg *in.OrderSuccessMsg) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	data := kafka.Message{
		Value: value,
	}

	err = c.WriterSuccess.WriteMessages(ctx, data)

	return err
}

func (c *KafkaClient) GetNewOrderMsg(ctx context.Context) (*in.NewOrderMsg, error) {
	data, err := c.NewOrdersReader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var msg in.NewOrderMsg
	err = json.Unmarshal(data.Value, &msg)

	return &msg, err
}

func (c *KafkaClient) GetOrderRejectedMsg(ctx context.Context) (*in.OrderRejectedMsg, error) {
	data, err := c.RejectedOrdersReader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var msg in.OrderRejectedMsg
	err = json.Unmarshal(data.Value, &msg)

	return &msg, err
}

func (c *KafkaClient) CloseReader() error {
	if err := c.NewOrdersReader.Close(); err != nil {
		return err
	}

	if err := c.RejectedOrdersReader.Close(); err != nil {
		return err
	}

	return nil
}

func (c *KafkaClient) CloseWriter() error {
	if err := c.WriterFails.Close(); err != nil {
		return err
	}

	if err := c.WriterSuccess.Close(); err != nil {
		return err
	}

	return nil
}

func (c *KafkaClient) ProduceHealthCheckMsg(ctx context.Context) error {
	log.Println("Produce health check")

	data := kafka.Message{
		Value: []byte{1},
	}

	dialer := &kafka.Dialer{
		Timeout:   30 * time.Second,
		DualStack: true,
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      c.brokers,
		Topic:        c.healthCheckTopic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		RequiredAcks: -1,
	})
	defer writer.Close()

	err := writer.WriteMessages(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (c *KafkaClient) ConsumeHealthCheckMsg(ctx context.Context) ([]byte, error) {
	log.Println("Consume health check")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    c.healthCheckTopic,
		GroupID:  "wallet",
		MinBytes: 10e1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	data, err := reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	if data.Value == nil {
		return nil, errors.New("got nill data")
	}

	return data.Value, nil
}

func (c *KafkaClient) HealthCheck(ctx context.Context) error {
	if err := c.ProduceHealthCheckMsg(ctx); err != nil {
		return err
	}

	if _, err := c.ConsumeHealthCheckMsg(ctx); err != nil {
		return err
	}

	return nil
}
