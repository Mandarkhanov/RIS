package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewClient(amqpURL string) (*Client, error) {
	var conn *amqp.Connection
	var err error

	maxRetries := 15
	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			break
		}

		log.Printf("RabbitMQ is not ready yet, retrying in 2s... (Attempt %d/%d)", i, maxRetries)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Failed to open a channel: %w", err)
	}

	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("Failed to set QoS: %w", err)
	}

	return &Client{
		conn: conn,
		ch:   ch,
	}, nil
}

func (c *Client) DeclareQueue(name string) error {
	_, err := c.ch.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Failed to declare a queue: %w", err)
	}
	return nil
}

func (c *Client) PublishXML(ctx context.Context, queueName string, body []byte) error {
	err := c.ch.PublishWithContext(ctx,
		"", // default direct exchange
		queueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // Сообщение персистентное
			ContentType:  "application/xml",
			Body:         body,
		})
	if err != nil {
		return fmt.Errorf("Failed to publish a message: %w", err)
	}
	return nil
}

func (c *Client) Consume(queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := c.ch.Consume(
		queueName,
		"",
		false, // auto-ack = false (ручной ACK)
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to register a consumer: %w", err)
	}
	return msgs, nil
}

func (c *Client) Close() {
	if c.ch != nil {
		if err := c.ch.Close(); err != nil {
			log.Printf("Failed to close RabbitMQ channel: %v", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Failed to close RabbitMQ connection: %v", err)
		}
	}
}
