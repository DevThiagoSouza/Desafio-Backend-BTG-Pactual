package rabbitmq


import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	portsin "github.com/DevThiagoSouza/src/application/ports/in"
	portsout "github.com/DevThiagoSouza/src/application/ports/out"
	"github.com/DevThiagoSouza/src/domain"
)

type OrderConsumer struct {
	orderCreator portsin.OrderCreator
	queue        string
	ch           *amqp091.Channel
	logger       portsout.Logger
}

func NewOrderConsumer(queue string, ch *amqp091.Channel, orderCreator portsin.OrderCreator, logger portsout.Logger) *OrderConsumer {
	return &OrderConsumer{
		ch:           ch,
		orderCreator: orderCreator,
		queue:        queue,
		logger:       logger,
	}
}

func (c *OrderConsumer) Consume(ctx context.Context) error {
	_, err := c.ch.QueueDeclare(
		c.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	if err := c.ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}
	
	msgs, err := c.ch.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			var order domain.Order
			if err := json.Unmarshal(msg.Body, &order); err != nil {
				if c.logger != nil {
					c.logger.Error(ctx, "failed to unmarshal message: %v", err)
				}
				msg.Nack(false, false)
				continue
			}

			if err := c.orderCreator.CreateOrder(ctx, order); err != nil {
				if c.logger != nil {
					c.logger.Error(ctx, "failed to save order: %v", err)
				}
				// Persistence issues can be transient, so requeue.
				msg.Nack(false, true)
				continue
			}

			_ = msg.Ack(false)
		}
	}
	
}