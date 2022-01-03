package broker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"
	in "wallet_service/internal/app/interfaces"

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

func NewKafkaClient(brokers []string, newOrdersTopic, rejectedOrdersTopic, successTopic, groupID string) (*KafkaClient, error) {
	if len(brokers) == 0 ||
		brokers[0] == "" ||
		newOrdersTopic == "" ||
		rejectedOrdersTopic == "" ||
		successTopic == "" ||
		groupID == "" {
		return nil, in.ErrInvalidBrokerConnParams
	}

	c := KafkaClient{
		brokers:          brokers,
		healthCheckTopic: "healthcheck",
	}

	c.NewOrdersReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    newOrdersTopic,
		GroupID:  groupID,
		MinBytes: 10e1,
		MaxBytes: 10e6,
	})

	c.RejectedOrdersReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    rejectedOrdersTopic,
		GroupID:  groupID,
		MinBytes: 10e1,
		MaxBytes: 10e6,
	})

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	c.WriterFails = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        rejectedOrdersTopic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		RequiredAcks: -1,
	})

	c.WriterSuccess = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        successTopic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		RequiredAcks: -1,
	})

	return &c, nil
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
	data, err := c.NewOrdersReader.ReadMessage(context.Background())
	if err != nil {
		return nil, err
	}

	var msg in.NewOrderMsg
	err = json.Unmarshal(data.Value, &msg)

	return &msg, err
}

func (c *KafkaClient) GetOrderRejectedMsg(ctx context.Context) (*in.OrderRejectedMsg, error) {
	data, err := c.RejectedOrdersReader.ReadMessage(context.Background())
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

	data, err := reader.ReadMessage(context.Background())
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
