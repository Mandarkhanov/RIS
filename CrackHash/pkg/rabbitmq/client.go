package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	url   string
	mu    sync.RWMutex
	conn  *amqp.Connection
	ch    *amqp.Channel
	ready bool
}

func NewClient(amqpURL string) *Client {
	c := &Client{url: amqpURL}
	go c.connectLoop()
	return c
}

func (c *Client) connectLoop() {
	for {
		conn, err := amqp.Dial(c.url)
		if err != nil {
			log.Printf("RabbitMQ connection failed, retrying in 2s...")
			time.Sleep(2 * time.Second)
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		err = ch.Qos(1, 0, false)
		if err != nil {
			ch.Close()
			conn.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		c.mu.Lock()
		c.conn = conn
		c.ch = ch
		c.ready = true
		c.mu.Unlock()

		log.Println("Successfully connected to RabbitMQ!")

		err = <-conn.NotifyClose(make(chan *amqp.Error))
		log.Printf("RabbitMQ connection closed: %v. Reconnecting...", err)

		c.mu.Lock()
		c.ready = false
		c.mu.Unlock()
	}
}

func (c *Client) getChannel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.ready || c.ch == nil {
		return nil, fmt.Errorf("RabbitMQ is not connected")
	}
	return c.ch, nil
}

func (c *Client) PublishXML(ctx context.Context, queueName string, body []byte) error {
	ch, err := c.getChannel()
	if err != nil {
		return err
	}
	ch.QueueDeclare(queueName, true, false, false, false, nil)

	return ch.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/xml",
		Body:         body,
	})
}

func (c *Client) Consume(queueName string, handler func(amqp.Delivery)) {
	go func() {
		for {
			ch, err := c.getChannel()
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			ch.QueueDeclare(queueName, true, false, false, false, nil)
			msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			for msg := range msgs {
				handler(msg)
			}

			time.Sleep(2 * time.Second)
		}
	}()
}

func (c *Client) PublishFanoutXML(ctx context.Context, exchangeName string, body []byte) error {
	ch, err := c.getChannel()
	if err != nil {
		return err
	}
	ch.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)

	return ch.PublishWithContext(ctx, exchangeName, "", false, false, amqp.Publishing{
		DeliveryMode: amqp.Transient,
		ContentType:  "application/xml",
		Body:         body,
	})
}

func (c *Client) ConsumeFanout(exchangeName string, handler func(amqp.Delivery)) {
	go func() {
		for {
			ch, err := c.getChannel()
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			ch.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
			q, err := ch.QueueDeclare("", false, true, true, false, nil)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			ch.QueueBind(q.Name, "", exchangeName, false, nil)

			msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			for msg := range msgs {
				handler(msg)
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
