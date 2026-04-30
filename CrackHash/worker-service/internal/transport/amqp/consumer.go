package amqp

import (
	"context"
	"crackhash/pkg/domain"
	"crackhash/worker/internal/service"
	"encoding/xml"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	svc *service.CrackWorkerService
}

func NewConsumer(svc *service.CrackWorkerService) *Consumer {
	return &Consumer{svc: svc}
}

func (c *Consumer) Start(msgs <-chan amqp.Delivery) {
	log.Println("Worker is waiting for tasks in RabbitMQ...")

	for msg := range msgs {
		log.Printf("Received a task message!")

		var task domain.WorkerTask
		if err := xml.Unmarshal(msg.Body, &task); err != nil {
			log.Printf("Failed to unmarshal task XML: %v", err)
			msg.Nack(false, false)
			continue
		}

		c.svc.ProcessTask(context.Background(), task)

		if err := msg.Ack(false); err != nil {
			log.Printf("Failed to ACK message: %v", err)
		} else {
			log.Printf("Task %s part %d successfully ACKed", task.RequestID, task.PartNumber)
		}
	}
}
