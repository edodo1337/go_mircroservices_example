package broker

import (
	"context"
	"encoding/json"
	in "registry_service/internal/app/interfaces"
)

type InMemoryBrokerClient struct {
	newOrdersChan      chan []byte
	rejectedOrdersChan chan []byte
}

func NewInMemoryBrokerClient() *InMemoryBrokerClient {
	c := InMemoryBrokerClient{
		newOrdersChan:      make(chan []byte, 10),
		rejectedOrdersChan: make(chan []byte, 10),
	}

	return &c
}

func (c *InMemoryBrokerClient) SendNewOrderMsg(ctx context.Context, msg *in.NewOrderMsg) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.newOrdersChan <- value

	return err
}

func (c *InMemoryBrokerClient) GetOrderRejectedMsg(ctx context.Context) (*in.OrderRejectedMsg, error) {
	data, ok := <-c.rejectedOrdersChan
	if !ok {
		return nil, in.ErrBrokerConnClosed
	}

	var msg in.OrderRejectedMsg
	err := json.Unmarshal(data, &msg)

	return &msg, err
}

func (c *InMemoryBrokerClient) CloseReader() error {
	close(c.rejectedOrdersChan)

	return nil
}

func (c *InMemoryBrokerClient) CloseWriter() error {
	close(c.newOrdersChan)

	return nil
}

func (c *InMemoryBrokerClient) GetSuccessMsg(ctx context.Context) (*in.OrderSuccessMsg, error) {
	panic("not impl")
}

func (c *InMemoryBrokerClient) ProduceHealthCheckMsg(ctx context.Context) error {
	return nil
}

func (c *InMemoryBrokerClient) ConsumeHealthCheckMsg(ctx context.Context) ([]byte, error) {
	return []byte{}, nil
}

func (c *InMemoryBrokerClient) HealthCheck(ctx context.Context) error {
	return nil
}
