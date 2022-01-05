package broker

import (
	"context"
	"encoding/json"
	"errors"
	in "registry_service/internal/app/interfaces"
	"registry_service/internal/pkg/conf"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	ReaderSuccess *kafka.Reader
	ReaderFail    *kafka.Reader

	Writer *kafka.Writer

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

	client.ReaderFail = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.Brokers,
		Topic:    c.RejectedOrdersTopic,
		GroupID:  c.GroupID,
		MinBytes: 10e1,
		MaxBytes: 10e6,
		MaxWait:  time.Duration(c.MaxWait) * time.Millisecond,
	})

	client.ReaderSuccess = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.Brokers,
		Topic:    c.SuccessTopic,
		GroupID:  c.GroupID,
		MinBytes: 10e1,
		MaxBytes: 10e6,
		MaxWait:  time.Duration(c.MaxWait) * time.Millisecond,
	})

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	client.Writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      c.Brokers,
		Topic:        c.NewOrdersTopic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		RequiredAcks: -1,
	})

	return &client, nil
}

func (c *KafkaClient) SendNewOrderMsg(ctx context.Context, msg *in.NewOrderMsg) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	data := kafka.Message{
		Value: value,
	}

	err = c.Writer.WriteMessages(ctx, data)

	return err
}

func (c *KafkaClient) GetOrderRejectedMsg(ctx context.Context) (*in.OrderRejectedMsg, error) {
	data, err := c.ReaderFail.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var msg in.OrderRejectedMsg
	err = json.Unmarshal(data.Value, &msg)

	return &msg, err
}

func (c *KafkaClient) GetSuccessMsg(ctx context.Context) (*in.OrderSuccessMsg, error) {
	data, err := c.ReaderSuccess.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var msg in.OrderSuccessMsg
	err = json.Unmarshal(data.Value, &msg)

	return &msg, err
}

func (c *KafkaClient) CloseReader() error {
	if err := c.ReaderFail.Close(); err != nil {
		return err
	}
	if err := c.ReaderSuccess.Close(); err != nil {
		return err
	}

	return nil
}

func (c *KafkaClient) CloseWriter() error {
	err := c.Writer.Close()

	return err
}

func (c *KafkaClient) ProduceHealthCheckMsg(ctx context.Context) error {
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
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    c.healthCheckTopic,
		GroupID:  "registry",
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
