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

func (c *Consumer) HandleMessage(msg amqp.Delivery) {
	var task domain.WorkerTask
	if err := xml.Unmarshal(msg.Body, &task); err != nil {
		log.Printf("Failed to unmarshal task XML: %v", err)
		msg.Nack(false, false)
		return
	}

	c.svc.ProcessTask(context.Background(), task)

	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to ACK message: %v", err)
	} else {
		log.Printf("Task [%s] part %d successfully ACKed", task.RequestID, task.PartNumber)
	}
}
